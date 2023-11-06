package indexer

import (
	common "github.com/arcology-network/common-lib/common"
	ccmap "github.com/arcology-network/common-lib/container/map"
	"github.com/arcology-network/common-lib/mempool"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type Importer struct {
	numThreads int
	store      interfaces.Datastore
	byTx       map[uint32][]interfaces.Univalue
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

	importer.byTx = make(map[uint32][]interfaces.Univalue)
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

func (this *Importer) Import(txTrans []interfaces.Univalue, args ...interface{}) {
	commitIfAbsent := common.IfThenDo1st(len(args) > 0 && args[0] != nil, func() bool { return args[0].(bool) }, true) //Write if absent from local

	common.RemoveIf(&txTrans, func(univ interfaces.Univalue) bool {
		return univ.Preexist() && this.store.IfExists(*univ.GetPath()) && !commitIfAbsent //preexists but not available locally, happen with a partial cache
	})

	common.Foreach(txTrans, func(univ *interfaces.Univalue) { // Create new sequences all at once
		if v, _ := this.deltaDict.Get(*(*univ).GetPath()); v == nil {
			this.deltaDict.Set(*(*univ).GetPath(), NewDeltaSequence(*(*univ).GetPath(), this))
		}
	})

	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			seq, _ := this.deltaDict.Get(*txTrans[i].GetPath())
			this.deltaDict.Set(*txTrans[i].GetPath(), seq.(*DeltaSequence).Add(txTrans[i])) // Add to the sequence
		}
	}
	common.ParallelWorker(len(txTrans), this.numThreads, worker)

	for i := 0; i < len(txTrans); i++ { // Update the transaction ID index
		if v := txTrans[i]; v != nil {
			tran := this.byTx[v.GetTx()]
			this.byTx[v.GetTx()] = append(common.IfThen(tran == nil, make([]interfaces.Univalue, 0, 32), tran), v)
		}
	}
}

// Only keep transation within the whitelist
func (this *Importer) WhilteList(whitelist []uint32) []error {
	if whitelist == nil { // Whiltelist all
		return []error{}
	}

	dict := make(map[uint32]bool)
	for _, txID := range whitelist {
		dict[txID] = true
	}

	for k, vec := range this.byTx {
		if k == ccurlcommon.SYSTEM {
			continue
		}

		if _, ok := dict[k]; !ok {
			for _, v := range vec {
				v.(*univalue.Univalue).SetPath(nil)
			}
		}
	}
	return []error{}
}

func (this *Importer) SortDeltaSequences() {
	this.keyBuffer = this.deltaDict.Keys()
	sorter := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
			deltaSeq.(*DeltaSequence).Sort() // Sort the transitions in the sequence
		}
	}
	common.ParallelWorker(len(this.keyBuffer), this.numThreads, sorter)
}

// Merge and finalize state deltas
func (this *Importer) MergeStateDelta() {
	this.valBuffer = this.valBuffer[:0]
	this.valBuffer = append(this.valBuffer, make([]interface{}, len(this.keyBuffer))...)

	finalizer := func(start, end, index int, args ...interface{}) {
		// for i := 0; i < len(this.keyBuffer); i++ {
		for i := start; i < end; i++ {
			deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
			finalized := deltaSeq.(*DeltaSequence).Finalize()
			this.valBuffer[i] = finalized

			if finalized == nil { // Some sequences may have been deleted with transactions they belong to
				this.keyBuffer[i] = ""
				continue
			}
		}
	}
	common.ParallelWorker(len(this.keyBuffer), this.numThreads, finalizer)

	common.Remove(&this.keyBuffer, "")
	common.RemoveIf(&this.valBuffer, func(v interface{}) bool { return v == nil })
}

func (this *Importer) KVs() ([]string, []interface{}) {
	common.Remove(&this.keyBuffer, "")
	common.RemoveIf(&this.valBuffer, func(v interface{}) bool { return v == nil })
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
