package importer

import (
	"fmt"
	"time"

	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	ccmap "github.com/arcology-network/common-lib/exp/map"
	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/mempool"
	committercommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type Importer struct {
	numThreads int
	store      interfaces.Datastore
	byTx       map[uint32]*[]*univalue.Univalue
	deltaDict  *ccmap.ConcurrentMap[string, *DeltaSequence]

	platform interfaces.Platform

	keyBuffer []string      // Keys updated in the cycle
	valBuffer []interface{} // Value updated in the cycle

	seqPool *mempool.Mempool[*DeltaSequence]
}

func NewImporter(store interfaces.Datastore, platform interfaces.Platform, args ...interface{}) *Importer {
	var importer Importer
	importer.numThreads = 8
	importer.store = store

	importer.byTx = make(map[uint32]*[]*univalue.Univalue)
	importer.platform = platform
	importer.deltaDict = ccmap.NewConcurrentMap(8, func(v *DeltaSequence) bool { return v == nil }, func(k string) uint8 {
		return array.Sum[uint8, uint8]([]byte(k))
	})

	importer.seqPool = mempool.NewMempool[*DeltaSequence](4096, 64, func() *DeltaSequence {
		return NewDeltaSequence("", nil) // Init an empty sequence.
	}, func(_ *DeltaSequence) {})
	return &importer
}

func (this *Importer) Init(store interfaces.Datastore) {
	this.store = store
	this.Clear()
}

func (this *Importer) SetStore(store interfaces.Datastore) { this.store = store }
func (this *Importer) Store() interfaces.Datastore         { return this.store }
func (this *Importer) ByPath() interface{}                 { return this.deltaDict }

func (this *Importer) IfExists(key string) bool {
	if _, ok := this.deltaDict.Get(key); !ok {
		return this.store.IfExists(key)
	}
	return false
}

func (this *Importer) NewSequenceFromPool(k string, store interfaces.Datastore) *DeltaSequence {
	return this.seqPool.New().Init(k, store)
}

func (this *Importer) Import(txTrans []*univalue.Univalue, args ...interface{}) {
	commitIfAbsent := common.IfThenDo1st(len(args) > 0 && args[0] != nil, func() bool { return args[0].(bool) }, true) //Write if absent from local

	t0 := time.Now()
	//Remove entries that preexist but not available locally, it happens with a partial cache
	array.RemoveIf(&txTrans, func(_ int, univ *univalue.Univalue) bool {
		return univ.Preexist() && this.store.IfExists(*univ.GetPath()) && !commitIfAbsent
	})
	fmt.Println("RemoveIf: ", len(txTrans), " in: ", time.Since(t0))

	t0 = time.Now()
	// Scan for paths that aren't in the delta dictionary yet.
	paths := array.ParallelAppend(make([]string, len(txTrans)), this.numThreads, func(i int, _ string) string {
		path := *txTrans[i].GetPath()
		v, _ := this.deltaDict.Get(path)
		return common.IfThen(v == nil, path, "") // Return empty strings for the existing entries.
	})
	paths = array.RemoveIf(&paths, func(_ int, v string) bool { return v == "" }) // Remove paths for existing entries by filtering out empty strings.
	fmt.Println("ParallelAppend + RemoveIf : ", len(txTrans), " in: ", time.Since(t0))

	t0 = time.Now()
	// Create new sequences for the non-existing paths all at once.
	sequences := array.ParallelAppend(paths, this.numThreads, func(i int, k string) *DeltaSequence {
		return this.NewSequenceFromPool(k, this.store)
	})
	fmt.Println("NewSequenceFromPool: ", len(txTrans), " in: ", time.Since(t0))

	t0 = time.Now()
	// Add the new sequences to the delta dictionary in parallel.
	this.deltaDict.BatchSet(paths, sequences)

	array.ParallelForeach(txTrans, this.numThreads, func(i int, _ **univalue.Univalue) {
		seq, _ := this.deltaDict.Get(*txTrans[i].GetPath())
		this.deltaDict.Set(*txTrans[i].GetPath(), seq.Add(txTrans[i])) // Add to the sequence
	})
	fmt.Println("deltaDict.BatchSet: ", len(txTrans), " in: ", time.Since(t0))

	t0 = time.Now()
	txIDs := array.Append(txTrans, func(_ int, v *univalue.Univalue) uint32 { return v.GetTx() })
	mapi.IfNotFoundDo(this.byTx, txIDs, func(k uint32) *[]*univalue.Univalue {
		v := (make([]*univalue.Univalue, 0, 16))
		return &v
	})

	array.ParallelForeach(txIDs, 4, func(i int, _ *uint32) {
		v := txTrans[i]
		tran := this.byTx[v.GetTx()]
		*tran = append(*tran, v)
	})
	fmt.Println("this.byTx[v.GetTx()] = append ", len(txTrans), " in: ", time.Since(t0))
}

// Only keep transation within the whitelist
func (this *Importer) WhilteList(whitelist []uint32) []error {
	if whitelist == nil { // Whiltelist all
		return []error{}
	}

	whitelisted := mapi.FromArray(whitelist, func(_ uint32) bool { return true })

	for txid, vec := range this.byTx {
		if txid == committercommon.SYSTEM {
			continue
		}

		if _, ok := whitelisted[txid]; !ok {
			for _, v := range *vec {
				v.SetPath(nil) // Mark the transition status, so that it can be removed later.
			}
		}
	}
	return []error{}
}

func (this *Importer) SortDeltaSequences() {
	this.keyBuffer = this.deltaDict.Keys()

	array.ParallelForeach(this.keyBuffer, this.numThreads, func(i int, _ *string) {
		deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
		deltaSeq.Sort() // Sort the transitions in the sequence
	})
}

// Merge and finalize state deltas
func (this *Importer) MergeStateDelta() {
	this.valBuffer = array.Resize(this.valBuffer, len(this.keyBuffer))

	array.ParallelForeach(this.keyBuffer, this.numThreads, func(i int, _ *string) {
		deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
		this.valBuffer[i] = deltaSeq.Finalize()

		if this.valBuffer[i] == nil || this.valBuffer[i].(*univalue.Univalue) == nil { // Some sequences may have been deleted with transactions they belong to
			this.keyBuffer[i] = ""
		}
	})

	array.Remove(&this.keyBuffer, "")
	array.RemoveIf(&this.valBuffer, func(_ int, v interface{}) bool { return v.(*univalue.Univalue) == nil })
}

func (this *Importer) KVs() ([]string, []interface{}) {
	array.Remove(&this.keyBuffer, "")
	array.RemoveIf(&this.valBuffer, func(_ int, v interface{}) bool { return v.(*univalue.Univalue) == nil })
	return this.keyBuffer, this.valBuffer
}

// Clear all
func (this *Importer) Clear() {
	for k, v := range this.byTx {
		*v = (*v)[:0]
		this.byTx[k] = v
	}

	this.deltaDict = ccmap.NewConcurrentMap(8, func(v *DeltaSequence) bool { return v == nil }, func(k string) uint8 {
		return array.Sum[uint8, uint8]([]byte(k))
	})

	this.keyBuffer = this.keyBuffer[:0]
	this.valBuffer = this.valBuffer[:0]

	this.seqPool.Reset()

	// this.seqPool.ForEachAllocated(func(obj *DeltaSequence) {
	// 	obj.Reclaim()
	// })

	// this.uniPool.ForEachAllocated(func(obj *univalue.Univalue) {
	// 	obj.Reclaim()
	// })

	this.seqPool.ReclaimRecursive()
	this.store.Clear()
}
