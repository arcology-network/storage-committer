package importer

import (
	common "github.com/arcology-network/common-lib/common"
	ccmap "github.com/arcology-network/common-lib/container/map"
	"github.com/arcology-network/common-lib/mempool"
	committercommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type Importer struct {
	numThreads int
	store      interfaces.Datastore
	byTx       map[uint32][]*univalue.Univalue
	deltaDict  *ccmap.ConcurrentMap

	platform interfaces.Platform

	keyBuffer []string      // Keys updated in the cycle
	valBuffer []interface{} // Value updated in the cycle
	seqPool   *mempool.Mempool
	uniPool   *mempool.Mempool
}

func NewImporter(store interfaces.Datastore, platform interfaces.Platform, args ...interface{}) *Importer {
	var importer Importer
	importer.numThreads = 8
	importer.store = store

	importer.byTx = make(map[uint32][]*univalue.Univalue)
	importer.platform = platform
	importer.deltaDict = ccmap.NewConcurrentMap()

	importer.seqPool = mempool.NewMempool("importer-seq", func() interface{} {
		return NewDeltaSequence("", nil)
	})

	importer.uniPool = mempool.NewMempool("importer-univalue", func() interface{} {
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
	v := this.uniPool.Get().(*univalue.Univalue)
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

	common.RemoveIf(&txTrans, func(univ *univalue.Univalue) bool {
		return univ.Preexist() && this.store.IfExists(*univ.GetPath()) && !commitIfAbsent //preexists but not available locally, happen with a partial cache
	})

	common.Foreach(txTrans, func(univ **univalue.Univalue, _ int) { // Create new sequences all at once
		if v, _ := this.deltaDict.Get(*(*univ).GetPath()); v == nil {
			this.deltaDict.Set(*(*univ).GetPath(), NewDeltaSequence(*(*univ).GetPath(), this))
		}
	})

	common.ParallelForeach(txTrans, this.numThreads, func(i int, _ **univalue.Univalue) {
		seq, _ := this.deltaDict.Get(*txTrans[i].GetPath())
		this.deltaDict.Set(*txTrans[i].GetPath(), seq.(*DeltaSequence).Add(txTrans[i])) // Add to the sequence
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
		// common.Foreach(vec, func(v **univalue.Univalue) {
		// 	_, ok := allowDict[k]
		// 	return !ok
		// })
	}
	return []error{}
}

func (this *Importer) SortDeltaSequences() {
	this.keyBuffer = this.deltaDict.Keys()

	common.ParallelForeach(this.keyBuffer, this.numThreads, func(i int, _ *string) {
		deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
		deltaSeq.(*DeltaSequence).Sort() // Sort the transitions in the sequence
	})
}

// Merge and finalize state deltas
func (this *Importer) MergeStateDelta() {
	this.valBuffer = this.valBuffer[:0]
	this.valBuffer = append(this.valBuffer, make([]interface{}, len(this.keyBuffer))...)

	common.ParallelForeach(this.keyBuffer, this.numThreads, func(i int, _ *string) {
		deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
		finalized := deltaSeq.(*DeltaSequence).Finalize()
		this.valBuffer[i] = finalized

		if finalized == nil { // Some sequences may have been deleted with transactions they belong to
			this.keyBuffer[i] = ""
		}
	})

	common.Remove(&this.keyBuffer, "")
	common.RemoveIf(&this.valBuffer, func(v interface{}) bool { return v.(*univalue.Univalue) == nil })
}

func (this *Importer) KVs() ([]string, []interface{}) {
	common.Remove(&this.keyBuffer, "")
	common.RemoveIf(&this.valBuffer, func(v interface{}) bool { return v.(*univalue.Univalue) == nil })
	return this.keyBuffer, this.valBuffer
}

// Clear all
func (this *Importer) Clear() {
	for k, v := range this.byTx {
		this.byTx[k] = v[:0]
	}

	this.deltaDict = ccmap.NewConcurrentMap()
	this.keyBuffer = this.keyBuffer[:0]
	this.valBuffer = this.valBuffer[:0]

	this.seqPool.ForEachAllocated(func(obj interface{}) {
		obj.(*DeltaSequence).Reclaim()
	})

	this.uniPool.ForEachAllocated(func(obj interface{}) {
		obj.(*univalue.Univalue).Reclaim()
	})

	this.seqPool.ReclaimRecursive()
	this.uniPool.ReclaimRecursive()
	this.store.Clear()
}
