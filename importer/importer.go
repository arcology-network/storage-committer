package importer

import (
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	ccmap "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/mempool"
	committercommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type Importer struct {
	numThreads int
	store      interfaces.Datastore
	byTx       map[uint32][]*univalue.Univalue
	deltaDict  *ccmap.ConcurrentMap[string, *DeltaSequence]

	platform interfaces.Platform

	keyBuffer []string      // Keys updated in the cycle
	valBuffer []interface{} // Value updated in the cycle

	seqPool *mempool.Mempool[*DeltaSequence]
	uniPool *mempool.Mempool[*univalue.Univalue]
}

func NewImporter(store interfaces.Datastore, platform interfaces.Platform, args ...interface{}) *Importer {
	var importer Importer
	importer.numThreads = 8
	importer.store = store

	importer.byTx = make(map[uint32][]*univalue.Univalue)
	importer.platform = platform
	importer.deltaDict = ccmap.NewConcurrentMap(8, func(v *DeltaSequence) bool { return v == nil }, func(k string) uint8 {
		return common.Sum[uint8, uint8]([]byte(k))
	})

	importer.seqPool = mempool.NewMempool[*DeltaSequence](4096, 64, func() *DeltaSequence {
		return NewDeltaSequence("", store)
	})

	importer.uniPool = mempool.NewMempool[*univalue.Univalue](4096, 64, func() *univalue.Univalue {
		return new(univalue.Univalue)
	})
	return &importer
}

func (this *Importer) Init(store interfaces.Datastore) {
	this.store = store
	this.Clear()
}

func (this *Importer) SetStore(store interfaces.Datastore) { this.store = store }
func (this *Importer) Store() interfaces.Datastore         { return this.store }
func (this *Importer) ByPath() interface{}                 { return this.deltaDict }

func (this *Importer) NewUnivalue() *univalue.Univalue {
	v := this.uniPool.Get()
	return v
}

func (this *Importer) IfExists(key string) bool {
	if _, ok := this.deltaDict.Get(key); !ok {
		return this.store.IfExists(key)
	}
	return false
}

func (this *Importer) Import(txTrans []*univalue.Univalue, args ...interface{}) {
	commitIfAbsent := common.IfThenDo1st(len(args) > 0 && args[0] != nil, func() bool { return args[0].(bool) }, true) //Write if absent from local

	array.RemoveIf(&txTrans, func(univ *univalue.Univalue) bool {
		return univ.Preexist() && this.store.IfExists(*univ.GetPath()) && !commitIfAbsent //preexists but not available locally, happen with a partial cache
	})

	array.Foreach(txTrans, func(_ int, univ **univalue.Univalue) { // Create new sequences all at once
		if v, _ := this.deltaDict.Get(*(*univ).GetPath()); v == nil {
			this.deltaDict.Set(*(*univ).GetPath(), NewDeltaSequence(*(*univ).GetPath(), this.store))
		}
	})

	array.ParallelForeach(txTrans, this.numThreads, func(i int, _ **univalue.Univalue) {
		seq, _ := this.deltaDict.Get(*txTrans[i].GetPath())
		this.deltaDict.Set(*txTrans[i].GetPath(), seq.Add(txTrans[i])) // Add to the sequence
	})

	for i := 0; i < len(txTrans); i++ { // Update the transaction ID index
		if v := txTrans[i]; v != nil {
			tran := this.byTx[v.GetTx()]
			this.byTx[v.GetTx()] = append(common.IfThen(tran == nil, make([]*univalue.Univalue, 0, 32), tran), v)
		}
	}
}

// Only keep transation within the whitelist
func (this *Importer) WhilteList(whitelist []uint32) []error {
	if whitelist == nil { // Whiltelist all
		return []error{}
	}

	whitelisted := make(map[uint32]bool)
	for _, txID := range whitelist {
		whitelisted[txID] = true
	}

	for txid, vec := range this.byTx {
		if txid == committercommon.SYSTEM {
			continue
		}

		if _, ok := whitelisted[txid]; !ok {
			for _, v := range vec {
				v.SetPath(nil) // Mark its status
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
	array.RemoveIf(&this.valBuffer, func(v interface{}) bool { return v.(*univalue.Univalue) == nil })
}

func (this *Importer) KVs() ([]string, []interface{}) {
	array.Remove(&this.keyBuffer, "")
	array.RemoveIf(&this.valBuffer, func(v interface{}) bool { return v.(*univalue.Univalue) == nil })
	return this.keyBuffer, this.valBuffer
}

// Clear all
func (this *Importer) Clear() {
	for k, v := range this.byTx {
		this.byTx[k] = v[:0]
	}

	this.deltaDict = ccmap.NewConcurrentMap(8, func(v *DeltaSequence) bool { return v == nil }, func(k string) uint8 {
		return common.Sum[uint8, uint8]([]byte(k))
	})

	this.keyBuffer = this.keyBuffer[:0]
	this.valBuffer = this.valBuffer[:0]

	this.seqPool.Reclaim()
	this.uniPool.Reclaim()

	// this.seqPool.ForEachAllocated(func(obj *DeltaSequence) {
	// 	obj.Reclaim()
	// })

	// this.uniPool.ForEachAllocated(func(obj *univalue.Univalue) {
	// 	obj.Reclaim()
	// })

	this.seqPool.ReclaimRecursive()
	this.uniPool.ReclaimRecursive()
	this.store.Clear()
}
