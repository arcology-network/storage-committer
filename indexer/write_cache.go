package indexer

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"

	common "github.com/arcology-network/common-lib/common"
	mempool "github.com/arcology-network/common-lib/mempool"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type WriteCache struct {
	store    ccurlcommon.DatastoreInterface
	kvDict   map[string]ccurlcommon.UnivalueInterface // Local KV lookup
	platform ccurlcommon.PlatformInterface
	buffer   []ccurlcommon.UnivalueInterface // Transition + access record buffer
	uniPool  *mempool.Mempool
}

func NewWriteCache(store ccurlcommon.DatastoreInterface, platform ccurlcommon.PlatformInterface, args ...interface{}) *WriteCache {
	var writeCache WriteCache
	writeCache.store = store
	writeCache.kvDict = make(map[string]ccurlcommon.UnivalueInterface)
	writeCache.store = store
	writeCache.platform = platform
	writeCache.buffer = make([]ccurlcommon.UnivalueInterface, 0, 64)

	writeCache.uniPool = mempool.NewMempool("writecache-univalue", func() interface{} { return new(univalue.Univalue) })
	return &writeCache
}

func (this *WriteCache) SetStore(store ccurlcommon.DatastoreInterface)    { this.store = store }
func (this *WriteCache) Store() ccurlcommon.DatastoreInterface            { return this.store }
func (this *WriteCache) Cache() *map[string]ccurlcommon.UnivalueInterface { return &this.kvDict }

// Merge two DB Caches
func (this *WriteCache) MergeFrom(other *WriteCache, txID uint32) {
	for k, from := range other.kvDict {
		// from.SetTx(txID)
		// if to, ok := this.kvDict[k]; ok { // already exists
		// to.IncrementReads(from.Reads())
		// to.IncrementWrites(from.Writes())
		// to.IncrementDelta(from.DeltaWrites())
		// to.SetValue(from.Value())
		// to.Value().(ccurlcommon.TypeInterface).Set()
		this.Write(txID, k, from.Value().(ccurlcommon.TypeInterface).Delta())

		// } else {
		// 	this.kvDict[k] = from
		// }
	}
}

func (this *WriteCache) NewUnivalue() *univalue.Univalue {
	v := this.uniPool.Get().(*univalue.Univalue)
	return v
}

// If the access has been recorded
func (this *WriteCache) GetOrInit(tx uint32, path string) ccurlcommon.UnivalueInterface {
	unival := this.kvDict[path]
	if unival == nil { // Not in the kvDict, check the datastore
		unival = this.NewUnivalue()
		unival.(*univalue.Univalue).Init(tx, path, 0, 0, this.RetriveShallow(path), this)
		this.kvDict[path] = unival // Adding to kvDict
	}
	return unival
}

func (this *WriteCache) Read(tx uint32, path string) interface{} {
	univalue := this.GetOrInit(tx, path)
	return univalue.Get(tx, path, this.Cache())
}

// Get the value directly, skip the access counting at the univalue level
func (this *WriteCache) Peek(path string) (interface{}, bool) {
	if v, ok := this.kvDict[path]; ok {
		return v.Value(), true
	}
	return this.RetriveShallow(path), false
}

func (this *WriteCache) Write(tx uint32, path string, value interface{}) error {
	parentPath := ccurlcommon.GetParentPath(path)
	if this.IfExists(parentPath) || tx == ccurlcommon.SYSTEM { // The parent path exists or to inject the path directly
		univalue := this.GetOrInit(tx, path) // Get a univalue wrapper

		err := univalue.Set(tx, path, value, this)
		if !this.platform.IsSysPath(parentPath) && tx != ccurlcommon.SYSTEM && err == nil { // System paths don't keep track of child paths
			parentMeta := this.GetOrInit(tx, parentPath)
			err = parentMeta.Set(tx, path, univalue.Value(), this)
		}
		return err
		// }
	}
	return errors.New("Error: The parent path doesn't exist: " + parentPath)
}

func (this *WriteCache) IfExists(path string) bool {
	return this.kvDict[path] != nil || this.RetriveShallow(path) != nil
}

func (this *WriteCache) Insert(path string, value interface{}) {
	this.kvDict[path] = value.(ccurlcommon.UnivalueInterface)
}

func (this *WriteCache) RetriveShallow(key string) interface{} {
	ret, _ := this.store.Retrive(key)
	return ret
}

func (this *WriteCache) Clear() {
	this.kvDict = make(map[string]ccurlcommon.UnivalueInterface)
}

func (this *WriteCache) Equal(other *WriteCache) bool {
	cache0 := []ccurlcommon.UnivalueInterface{}
	cache1 := []ccurlcommon.UnivalueInterface{}

	this.Vectorize(&this.kvDict, &cache0, true)
	other.Vectorize(&this.kvDict, &cache1, true)
	cacheFlag := reflect.DeepEqual(cache0, cache1)
	return cacheFlag
}

/* Map to array */
func (*WriteCache) Vectorize(dict *map[string]ccurlcommon.UnivalueInterface, valBuf *[]ccurlcommon.UnivalueInterface, needToSort bool) {
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

func (this *WriteCache) Export(preprocessors ...func([]ccurlcommon.UnivalueInterface) []ccurlcommon.UnivalueInterface) []ccurlcommon.UnivalueInterface {
	this.buffer = this.buffer[:0]
	this.Vectorize(&this.kvDict, &this.buffer, false) // Export records to the buffer

	for _, processor := range preprocessors {
		this.buffer = common.IfThenDo1st(processor != nil, func() []ccurlcommon.UnivalueInterface {
			return processor(this.buffer)
		}, this.buffer)
	}
	return this.buffer
}

func (this *WriteCache) Print() {
	values := []ccurlcommon.UnivalueInterface{}
	this.Vectorize(&this.kvDict, &values, true)
	for i, elem := range values {
		fmt.Println("Level : ", i)
		elem.Print()
	}
}
