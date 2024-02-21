package importer

import (
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

	//Remove entries that preexist but not available locally, it happens with a partial cache
	array.RemoveIf(&txTrans, func(_ int, univ *univalue.Univalue) bool {
		return univ.Preexist() && this.store.IfExists(*univ.GetPath()) && !commitIfAbsent
	})

	// Create new sequences for the non-existing paths all at once.
	missingKeys := array.ParallelAppend(txTrans, this.numThreads, func(i int, _ *univalue.Univalue) string {
		if _, ok := this.deltaDict.Get(*txTrans[i].GetPath()); !ok {
			return *txTrans[i].GetPath()
		}
		return ""
	})

	array.Remove(&missingKeys, "")
	missingKeys = mapi.Keys(mapi.FromArray(missingKeys, func(k string) bool { return true })) // Get the unique keys only
	this.deltaDict.BatchSetWith(missingKeys, func(k *string) *DeltaSequence { return NewDeltaSequence(*k, this.store) })

	// Update the delta dictionary with the new sequences in parallel.
	this.deltaDict.ParallelFor(0, len(txTrans),
		func(i int) string { return *txTrans[i].GetPath() },
		func(i int, k string, seq *DeltaSequence, _ bool) (*DeltaSequence, bool) {
			return seq.Add(txTrans[i]), false
		})

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
