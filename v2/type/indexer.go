package ccurltype

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"

	common "github.com/arcology/common-lib/common"
	cccontainermap "github.com/arcology/common-lib/concurrentcontainer/map"
	"github.com/arcology/common-lib/mempool"
	performance "github.com/arcology/common-lib/mhasher"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
)

type Indexer struct {
	numThreads int
	store      ccurlcommon.DatastoreInterface
	buffer     map[string]ccurlcommon.UnivalueInterface // KV lookup
	byTx       map[uint32][]ccurlcommon.UnivalueInterface
	byPath     *cccontainermap.ConcurrentMap

	platform      *ccurlcommon.Platform
	updatedKeys   []string      // Keys updated in the circle
	updatedValues []interface{} // Value updated in the circle
	seqPool       *mempool.Mempool
	uniPool       *mempool.Mempool
}

func NewIndexer(store ccurlcommon.DatastoreInterface, platform *ccurlcommon.Platform, args ...interface{}) *Indexer {
	var indexer Indexer
	indexer.numThreads = 8
	indexer.store = store
	indexer.buffer = make(map[string]ccurlcommon.UnivalueInterface)
	indexer.byTx = make(map[uint32][]ccurlcommon.UnivalueInterface)
	indexer.platform = platform
	indexer.byPath = cccontainermap.NewConcurrentMap()

	indexer.seqPool = mempool.NewMempool("seq", func() interface{} {
		return NewDeltaSequence()
	})

	indexer.uniPool = mempool.NewMempool("univalue", func() interface{} {
		return new(Univalue)
	})
	return &indexer
}

func (indexer *Indexer) Init(store ccurlcommon.DatastoreInterface) {
	indexer.store = store
	indexer.Clear()
}

func (this *Indexer) Store() *ccurlcommon.DatastoreInterface            { return &this.store }
func (this *Indexer) Buffer() *map[string]ccurlcommon.UnivalueInterface { return &this.buffer }
func (this *Indexer) ByPath() interface{}                               { return this.byPath }

func (this *Indexer) IfExists(path string) bool {
	return this.buffer[path] != nil || this.RetriveShallow(path) != nil
}

func (this *Indexer) NewUnivalue() *Univalue {
	v := this.uniPool.Get().(*Univalue)
	return v
}

// If the access has been recorded
func (this *Indexer) CheckHistory(tx uint32, path string, ifAddToBuffer bool) ccurlcommon.UnivalueInterface {
	univalue := this.buffer[path]
	if univalue == nil { // Not in the buffer, check the datastore
		univalue = this.NewUnivalue()
		univalue.(*Univalue).Init(ccurlcommon.VARIATE_TRANSITIONS, tx, path, 0, 0, this.RetriveShallow(path), this)

		if ifAddToBuffer {
			this.buffer[path] = univalue
		}
	}
	return univalue
}

func (this *Indexer) Read(tx uint32, path string) interface{} {
	univalue := this.CheckHistory(tx, path, true)
	return univalue.Get(tx, path, this.Buffer())
}

// Get the value directly, bypassing the univalue level
func (this *Indexer) TryRead(tx uint32, path string) interface{} {
	if v, ok := this.buffer[path]; ok {
		return v.Peek(this.Buffer())
	}
	return this.RetriveShallow(path)
}

func (this *Indexer) Write(tx uint32, path string, value interface{}) error {
	parentPath := ccurlcommon.GetParentPath(path)
	if this.IfExists(parentPath) || tx == ccurlcommon.SYSTEM { // The parent path exists or to inject the path directly
		univalue := this.CheckHistory(tx, path, true)
		if univalue.Value() == nil && value == nil { // Try to delete something nonexistent
			return nil
		} else {
			err := univalue.Set(tx, path, value, this)
			if !this.platform.OnControlList(parentPath) && tx != ccurlcommon.SYSTEM && err == nil { // System paths don't keep track of child paths
				if parentValue := this.CheckHistory(tx, parentPath, false); parentValue != nil && parentValue.Value() != nil {
					if parentValue.UpdateParentMeta(tx, univalue, this) {
						this.buffer[parentPath] = parentValue
					}
				}

			}
			return err
		}
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
				deltaSeq[i].(*DeltaSequence).Reset(nKeys[i], this, uniPool) // Reset the sequence

				if !preexist { //new entry
					continue
				}
				deltaSeq[i].(*DeltaSequence).Init(nKeys[i], this, uniPool) // Get the initial value from the cache / persistent DB
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
				v.(*Univalue).path = nil
			}
		}
	}
	return []error{}
}

func (this *Indexer) SortTransitions() {
	this.updatedKeys = this.byPath.Keys()
	var err error
	this.updatedKeys, err = performance.SortStrings(this.updatedKeys) // Keys should be unique
	if err != nil {
		panic(err)
	}

	sorter := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			deltaSeq, _ := this.byPath.Get(this.updatedKeys[i])
			deltaSeq.(*DeltaSequence).Sort()
		}
	}
	common.ParallelWorker(len(this.updatedKeys), this.numThreads, sorter)
}

// 	Merge and finalize state deltas
func (this *Indexer) FinalizeStates() {
	this.updatedValues = this.updatedValues[:0]
	this.updatedValues = append(this.updatedValues, make([]interface{}, len(this.updatedKeys))...)
	finalizer := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			deltaSeq, _ := this.byPath.Get(this.updatedKeys[i])
			deltaSeq.(*DeltaSequence).Finalize()
			if deltaSeq.(*DeltaSequence).Value() == nil { // Some sequences may have been deleted with transactions they belong
				this.updatedKeys[i] = ""
				this.updatedValues[i] = nil
				continue
			}
			this.updatedValues[i] = deltaSeq.(*DeltaSequence).Value().(ccurlcommon.UnivalueInterface)
		}
	}
	common.ParallelWorker(len(this.updatedKeys), this.numThreads, finalizer)
	common.RemoveEmptyStrings(&this.updatedKeys)
	common.RemoveNils(&this.updatedValues)
}

func (this *Indexer) KVs() ([]string, []interface{}) {
	common.RemoveEmptyStrings(&this.updatedKeys)
	common.RemoveNils(&this.updatedValues)
	return this.updatedKeys, this.updatedValues
}

// Clear all
func (this *Indexer) Clear() {
	this.buffer = make(map[string]ccurlcommon.UnivalueInterface)
	for k, v := range this.byTx {
		this.byTx[k] = v[:0]
	}

	this.byPath = cccontainermap.NewConcurrentMap()
	this.updatedKeys = this.updatedKeys[:0]
	this.updatedValues = this.updatedValues[:0]

	this.seqPool.ForEachAllocated(func(obj interface{}) {
		obj.(*DeltaSequence).Reclaim()
	})

	this.uniPool.ForEachAllocated(func(obj interface{}) {
		obj.(*Univalue).Reclaim()
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
