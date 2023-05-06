package indexer

import (
	"sort"

	common "github.com/arcology-network/common-lib/common"
	ccmap "github.com/arcology-network/common-lib/container/map"
	"github.com/arcology-network/common-lib/mempool"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type Importer struct {
	numThreads int
	store      ccurlcommon.DatastoreInterface
	byTx       map[uint32][]ccurlcommon.UnivalueInterface
	byPath     *ccmap.ConcurrentMap

	platform ccurlcommon.PlatformInterface

	updatedKeys   []string      // Keys updated in the circle
	updatedValues []interface{} // Value updated in the circle
	seqPool       *mempool.Mempool
	uniPool       *mempool.Mempool
}

func NewImporter(store ccurlcommon.DatastoreInterface, platform ccurlcommon.PlatformInterface, args ...interface{}) *Importer {
	var indexer Importer
	indexer.numThreads = 8
	indexer.store = store

	indexer.byTx = make(map[uint32][]ccurlcommon.UnivalueInterface)
	indexer.platform = platform
	indexer.byPath = ccmap.NewConcurrentMap()

	indexer.seqPool = mempool.NewMempool("importer-seq", func() interface{} {
		return NewDeltaSequence()
	})

	indexer.uniPool = mempool.NewMempool("importer-univalue", func() interface{} {
		return new(univalue.Univalue)
	})
	return &indexer
}

func (this *Importer) Init(store ccurlcommon.DatastoreInterface) {
	this.store = store
	this.Clear()
}

func (this *Importer) Store() *ccurlcommon.DatastoreInterface { return &this.store }
func (this *Importer) ByPath() interface{}                    { return this.byPath }

func (this *Importer) NewUnivalue() *univalue.Univalue {
	v := this.uniPool.Get().(*univalue.Univalue)
	return v
}

func (this *Importer) RetriveShallow(key string) interface{} {
	ret, _ := this.store.Retrive(key)
	return ret
}

func (this *Importer) Import(txTrans []ccurlcommon.UnivalueInterface, args ...interface{}) {
	ifCommit := true
	if len(args) > 0 && args[0] != nil {
		ifCommit = args[0].(bool)
	}

	nKeys := make([]string, len(txTrans))
	for i, v := range txTrans {
		nKeys[i] = *v.GetPath()
	}

	// Create delta sequences all at once
	deltaSeq := this.byPath.BatchGet(nKeys) // If the entries exist in the RWCache already

	inLocalCache := this.store.BatchRetrive(nKeys)
	worker := func(start, end, index int, args ...interface{}) {
		seqPool := this.seqPool.GetTlsMempool(index)
		uniPool := this.uniPool.GetTlsMempool(index)
		for i := start; i < end; i++ {
			if deltaSeq[i] == nil { // The entry does't exist in the RWCache
				preexist := txTrans[i].Preexist()
				if inLocalCache[i] == nil && ifCommit && preexist { //preexists but not available locally
					nKeys[i] = ""
					txTrans[i] = nil
					continue
				}

				deltaSeq[i] = seqPool.Get().(*DeltaSequence)                // create a new delta sequence from the pool
				deltaSeq[i].(*DeltaSequence).Reset(nKeys[i], this, uniPool) // Reset the sequence in case it was used before

				if preexist {
					deltaSeq[i].(*DeltaSequence).Init(nKeys[i], this, uniPool) // Get the initial value from the cache / persistent DB
				}
			}
		}
	}
	common.ParallelWorker(len(nKeys), this.numThreads, worker)

	this.byPath.BatchSet(nKeys, deltaSeq)                          // Create new empty sequences all at once
	Inserter := func(start, end, index int, args ...interface{}) { // Insert the transitions to the sequences
		for i := start; i < end; i++ {
			if deltaSeq[i] == nil {
				continue
			}

			deltaSeq, _ := this.byPath.Get(*txTrans[i].GetPath())
			deltaSeq.(*DeltaSequence).Insert(txTrans[i])
		}
	}
	common.ParallelWorker(len(txTrans), this.numThreads, Inserter)

	for i := 0; i < len(txTrans); i++ { // Update the transaction ID index
		v := txTrans[i]
		if v == nil {
			continue
		}

		if this.byTx[v.GetTx()] == nil {
			this.byTx[v.GetTx()] = make([]ccurlcommon.UnivalueInterface, 0, 32)
		}
		this.byTx[v.GetTx()] = append(this.byTx[v.GetTx()], v)
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

func (this *Importer) SortTransitions() {
	this.updatedKeys = this.byPath.Keys()
	// var err error
	// this.updatedKeys, err = performance.SortStrings(this.updatedKeys) // Keys should be unique
	// if err != nil {
	// 	panic(err)
	// }
	sort.Strings(this.updatedKeys) // For the later merkle tree calculation

	sorter := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			deltaSeq, _ := this.byPath.Get(this.updatedKeys[i])

			// typeValue := deltaSeq.(*DeltaSequence).base
			// if typeValue == nil {
			// 	typeValue = deltaSeq.(*DeltaSequence).values[0]
			// }

			// if typeValue.Value().(ccurlcommon.TypeInterface).TypeID() == commutative.Path() {
			deltaSeq.(*DeltaSequence).Sort() // Sort the transitions in the sequence
			// }
		}
	}
	common.ParallelWorker(len(this.updatedKeys), this.numThreads, sorter)
}

// Merge and finalize state deltas
func (this *Importer) FinalizeStates() {
	this.updatedValues = this.updatedValues[:0]
	this.updatedValues = append(this.updatedValues, make([]interface{}, len(this.updatedKeys))...)
	finalizer := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			deltaSeq, _ := this.byPath.Get(this.updatedKeys[i])
			deltaSeq.(*DeltaSequence).Finalize()
			if deltaSeq.(*DeltaSequence).Value() == nil { // Some sequences may have been deleted with transactions they belong to
				this.updatedKeys[i] = ""
				this.updatedValues[i] = nil
				continue
			}
			this.updatedValues[i] = deltaSeq.(*DeltaSequence).Value().(ccurlcommon.UnivalueInterface)
		}
	}
	common.ParallelWorker(len(this.updatedKeys), this.numThreads, finalizer)
	common.Remove(&this.updatedKeys, "")
	common.RemoveIf(&this.updatedValues, func(v interface{}) bool { return v == nil })
}

func (this *Importer) KVs() ([]string, []interface{}) {
	common.Remove(&this.updatedKeys, "")
	common.RemoveIf(&this.updatedValues, func(v interface{}) bool { return v == nil })
	return this.updatedKeys, this.updatedValues
}

// Clear all
func (this *Importer) Clear() {
	// this.RWCache = make(map[string]ccurlcommon.UnivalueInterface)
	for k, v := range this.byTx {
		this.byTx[k] = v[:0]
	}

	this.byPath = ccmap.NewConcurrentMap()
	this.updatedKeys = this.updatedKeys[:0]
	this.updatedValues = this.updatedValues[:0]

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
