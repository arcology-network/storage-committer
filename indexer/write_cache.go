package indexer

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	common "github.com/arcology-network/common-lib/common"
	mempool "github.com/arcology-network/common-lib/mempool"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	concurrenturlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type WriteCache struct {
	store    interfaces.Datastore
	kvDict   map[string]interfaces.Univalue // Local KV lookup
	platform interfaces.Platform
	buffer   []interfaces.Univalue // Transition + access record buffer
	uniPool  *mempool.Mempool
}

func NewWriteCache(store interfaces.Datastore, args ...interface{}) *WriteCache {
	var writeCache WriteCache
	writeCache.store = store
	writeCache.kvDict = make(map[string]interfaces.Univalue)
	writeCache.store = store
	writeCache.platform = concurrenturlcommon.NewPlatform()
	writeCache.buffer = make([]interfaces.Univalue, 0, 64)

	writeCache.uniPool = mempool.NewMempool("writecache-univalue", func() interface{} { return new(univalue.Univalue) })
	return &writeCache
}

func (this *WriteCache) SetStore(store interfaces.Datastore)    { this.store = store }
func (this *WriteCache) Store() interfaces.Datastore            { return this.store }
func (this *WriteCache) Cache() *map[string]interfaces.Univalue { return &this.kvDict }

func (this *WriteCache) NewUnivalue() *univalue.Univalue {
	v := this.uniPool.Get().(*univalue.Univalue)
	return v
}

// If the access has been recorded
func (this *WriteCache) GetOrInit(tx uint32, path string) interfaces.Univalue {
	unival := this.kvDict[path]
	if unival == nil { // Not in the kvDict, check the datastore
		unival = this.NewUnivalue()
		unival.(*univalue.Univalue).Init(tx, path, 0, 0, 0, this.RetriveShallow(path), this)
		this.kvDict[path] = unival // Adding to kvDict
	}
	return unival
}

func (this *WriteCache) Read(tx uint32, path string) (interface{}, interface{}) {
	univalue := this.GetOrInit(tx, path)
	return univalue.Get(tx, path, nil), univalue
}

func (this *WriteCache) Do(tx uint32, path string, doer interface{}) interface{} {
	univalue := this.GetOrInit(tx, path)
	return univalue.Do(tx, path, doer)
}

// Get the value directly, skip the access counting at the univalue level
func (this *WriteCache) Peek(path string) (interface{}, interface{}) {
	if v, ok := this.kvDict[path]; ok {
		return v.Value(), v
	}

	v := this.RetriveShallow(path)
	return v, univalue.NewUnivalue(ccurlcommon.SYSTEM, path, 0, 0, 0, v)
}

func (this *WriteCache) Write(tx uint32, path string, value interface{}, persistent bool) error {
	parentPath := common.GetParentPath(path)
	if this.IfExists(parentPath) || tx == ccurlcommon.SYSTEM { // The parent path exists or to inject the path directly
		univalue := this.GetOrInit(tx, path) // Get a univalue wrapper

		err := univalue.Set(tx, path, value, this)
		if err == nil {
			if strings.HasSuffix(parentPath, "container/") || (!this.platform.IsSysPath(parentPath) && tx != ccurlcommon.SYSTEM) { // Don't keep track of the system children
				parentMeta := this.GetOrInit(tx, parentPath)
				err = parentMeta.Set(tx, path, univalue.Value(), this)
			}
		}
		return err
	}
	// strings.HasPrefix(parentPath, "container/") &&
	return errors.New("Error: The parent path doesn't exist: " + parentPath)
}

func (this *WriteCache) IfExists(path string) bool {
	return this.kvDict[path] != nil || this.RetriveShallow(path) != nil
}

func (this *WriteCache) Insert(path string, value interface{}) {
	this.kvDict[path] = value.(interfaces.Univalue)
}

func (this *WriteCache) RetriveShallow(key string) interface{} {
	ret, _ := this.store.Retrive(key)
	return ret
}

func (this *WriteCache) Clear() {
	this.kvDict = make(map[string]interfaces.Univalue)
}

func (this *WriteCache) Equal(other *WriteCache) bool {
	thisBuffer := common.MapValues(this.kvDict)
	sort.SliceStable(thisBuffer, func(i, j int) bool {
		return *thisBuffer[i].GetPath() < *thisBuffer[j].GetPath()
	})

	otherBuffer := common.MapValues(other.kvDict)
	sort.SliceStable(otherBuffer, func(i, j int) bool {
		return *otherBuffer[i].GetPath() < *otherBuffer[j].GetPath()
	})

	cacheFlag := reflect.DeepEqual(thisBuffer, otherBuffer)
	return cacheFlag
}

func (this *WriteCache) Export(preprocessors ...func([]interfaces.Univalue) []interfaces.Univalue) []interfaces.Univalue {
	this.buffer = common.MapValues(this.kvDict) //this.buffer[:0]

	for _, processor := range preprocessors {
		this.buffer = common.IfThenDo1st(processor != nil, func() []interfaces.Univalue {
			return processor(this.buffer)
		}, this.buffer)
	}

	common.RemoveIf(&this.buffer, func(v interfaces.Univalue) bool { return v.Reads() == 0 && v.IsReadOnly() }) // Remove peeks
	return this.buffer
}

func (this *WriteCache) Print() {
	values := common.MapValues(this.kvDict)
	sort.SliceStable(values, func(i, j int) bool {
		return *values[i].GetPath() < *values[j].GetPath()
	})

	for i, elem := range values {
		fmt.Println("Level : ", i)
		elem.Print()
	}
}
