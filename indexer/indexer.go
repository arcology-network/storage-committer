package indexer

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"

	common "github.com/arcology-network/common-lib/common"
	ccmap "github.com/arcology-network/common-lib/container/map"
	"github.com/arcology-network/common-lib/mempool"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type Indexer struct {
	numThreads int
	store      ccurlcommon.DatastoreInterface
	buffer     map[string]ccurlcommon.UnivalueInterface // KV lookup
	byTx       map[uint32][]ccurlcommon.UnivalueInterface
	byPath     *ccmap.ConcurrentMap

	platform      ccurlcommon.PlatformInterface
	updatedKeys   []string      // Keys updated in the circle
	updatedValues []interface{} // Value updated in the circle
	seqPool       *mempool.Mempool
	uniPool       *mempool.Mempool
}

func NewIndexer(store ccurlcommon.DatastoreInterface, platform ccurlcommon.PlatformInterface, args ...interface{}) *Indexer {
	var indexer Indexer
	indexer.numThreads = 8
	indexer.store = store
	indexer.buffer = make(map[string]ccurlcommon.UnivalueInterface)
	indexer.byTx = make(map[uint32][]ccurlcommon.UnivalueInterface)
	indexer.platform = platform
	indexer.byPath = ccmap.NewConcurrentMap()

	indexer.seqPool = mempool.NewMempool("seq", func() interface{} {
		return NewDeltaSequence()
	})

	indexer.uniPool = mempool.NewMempool("univalue", func() interface{} {
		return new(univalue.Univalue)
	})
	return &indexer
}

// func (this *Indexer) Platform() *ccurlcommon.Platform { return this.platform }

// Merge two indexers
func (this *Indexer) MergeFrom(other *Indexer) {
	for k, from := range other.buffer {
		if to, ok := this.buffer[k]; ok { // already exists
			to.IncrementReads(from.Reads())
			to.IncrementWrites(from.Writes())
			to.IncrementDelta(from.DeltaWrites())
			to.SetValue(from.Value())
		}
	}
}

func (this *Indexer) Init(store ccurlcommon.DatastoreInterface) {
	this.store = store
	this.Clear()
}

func (this *Indexer) Store() *ccurlcommon.DatastoreInterface            { return &this.store }
func (this *Indexer) Buffer() *map[string]ccurlcommon.UnivalueInterface { return &this.buffer }
func (this *Indexer) ByPath() interface{}                               { return this.byPath }

func (this *Indexer) IfExists(path string) bool {
	return this.buffer[path] != nil || this.RetriveShallow(path) != nil
}

func (this *Indexer) NewUnivalue() *univalue.Univalue {
	v := this.uniPool.Get().(*univalue.Univalue)
	return v
}

// If the access has been recorded
func (this *Indexer) GetOrInit(tx uint32, path string) ccurlcommon.UnivalueInterface {
	unival := this.buffer[path]
	if unival == nil { // Not in the buffer, check the datastore
		unival = this.NewUnivalue()
		unival.(*univalue.Univalue).Init(tx, path, 0, 0, this.RetriveShallow(path), this)
		this.buffer[path] = unival // Adding to buffer
	}
	return unival
}

func (this *Indexer) Read(tx uint32, path string) interface{} {
	univalue := this.GetOrInit(tx, path)
	return univalue.Get(tx, path, this.Buffer())
}

// Get the value directly, skip the access counting at the univalue level
func (this *Indexer) Peek(path string) (interface{}, bool) {
	if v, ok := this.buffer[path]; ok {
		return v.Value(), true
	}
	return this.RetriveShallow(path), false
}

func (this *Indexer) Write(tx uint32, path string, value interface{}) error {
	parentPath := ccurlcommon.GetParentPath(path)
	if this.IfExists(parentPath) || tx == ccurlcommon.SYSTEM { // The parent path exists or to inject the path directly
		univalue := this.GetOrInit(tx, path) // Get a univalue wrapper

		err := univalue.Set(tx, path, value, this)
		if !this.platform.IsSysPath(parentPath) && tx != ccurlcommon.SYSTEM && err == nil { // System paths don't keep track of child paths
			if parentMeta := this.GetOrInit(tx, parentPath); parentMeta != nil && parentMeta.Value() != nil {
				err = parentMeta.Set(tx, path, univalue.Value(), this)
			}
		}
		return err
		// }
	}
	return errors.New("Error: The parent path doesn't exist: " + parentPath)
}

func (this *Indexer) Insert(path string, value interface{}) {
	this.buffer[path] = value.(ccurlcommon.UnivalueInterface)
}

func (this *Indexer) RetriveShallow(key string) interface{} {
	ret, _ := this.store.Retrive(key)
	return ret
}

func (this *Indexer) Import(txTrans []ccurlcommon.UnivalueInterface, args ...interface{}) {
	ifCommit := true
	if len(args) > 0 && args[0] != nil {
		ifCommit = args[0].(bool)
	}

	nKeys := make([]string, len(txTrans))
	for i, v := range txTrans {
		nKeys[i] = *v.GetPath()
	}

	// Create delta sequences all at once
	deltaSeq := this.byPath.BatchGet(nKeys) // If the entries exist in the buffer already

	inLocalCache := this.store.BatchRetrive(nKeys)
	worker := func(start, end, index int, args ...interface{}) {
		seqPool := this.seqPool.GetTlsMempool(index)
		uniPool := this.uniPool.GetTlsMempool(index)
		for i := start; i < end; i++ {
			if deltaSeq[i] == nil { // The entry does't exist in the buffer
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
func (this *Indexer) WhilteList(whitelist []uint32) []error {
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

func (this *Indexer) SortTransitions() {
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
func (this *Indexer) FinalizeStates() {
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

func (this *Indexer) KVs() ([]string, []interface{}) {
	common.Remove(&this.updatedKeys, "")
	common.RemoveIf(&this.updatedValues, func(v interface{}) bool { return v == nil })
	return this.updatedKeys, this.updatedValues
}

// Clear all
func (this *Indexer) Clear() {
	this.buffer = make(map[string]ccurlcommon.UnivalueInterface)
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

/* Map to array */
func (*Indexer) Vectorize(dict *map[string]ccurlcommon.UnivalueInterface, valBuf *[]ccurlcommon.UnivalueInterface, needToSort bool) {
	*valBuf = (*valBuf)[:0]
	for _, v := range *dict {
		*valBuf = append((*valBuf), v)
	}

	if needToSort { // Sort by path
		sort.SliceStable(*valBuf, func(i, j int) bool {
			return bytes.Compare([]byte(*(*valBuf)[i].GetPath())[:], []byte(*(*valBuf)[j].GetPath())[:]) < 0
		})
	}
}

func (this *Indexer) Equal(other *Indexer) bool {
	cache0 := []ccurlcommon.UnivalueInterface{}
	cache1 := []ccurlcommon.UnivalueInterface{}

	this.Vectorize(&this.buffer, &cache0, true)
	other.Vectorize(&this.buffer, &cache1, true)
	cacheFlag := reflect.DeepEqual(cache0, cache1)
	return cacheFlag
}

func (this *Indexer) Print() {
	values := []ccurlcommon.UnivalueInterface{}
	this.Vectorize(&this.buffer, &values, true)
	for i, elem := range values {
		fmt.Println("Level : ", i)
		elem.Print()
	}
}
