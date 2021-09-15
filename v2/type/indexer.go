package ccurltype

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	orderedmap "github.com/elliotchance/orderedmap"
)

type Indexer struct {
	store        ccurlcommon.DB
	buffer       map[string]ccurlcommon.UnivalueInterface
	byTx         map[uint32][]ccurlcommon.UnivalueInterface
	byPath       map[string]*orderedmap.OrderedMap
	byPathFirst  map[string]*orderedmap.Element
	uniquePaths  []string
	importBuffer map[string]ccurlcommon.UnivalueInterface
	platform     *ccurlcommon.Platform
}

func NewIndexer(store ccurlcommon.DB, platform *ccurlcommon.Platform) *Indexer {
	var indexer Indexer
	indexer.store = store
	indexer.buffer = make(map[string]ccurlcommon.UnivalueInterface, 1024)
	indexer.byTx = make(map[uint32][]ccurlcommon.UnivalueInterface)
	indexer.importBuffer = make(map[string]ccurlcommon.UnivalueInterface)
	indexer.byPath = make(map[string]*orderedmap.OrderedMap)
	indexer.byPathFirst = make(map[string]*orderedmap.Element)
	indexer.uniquePaths = make([]string, 1024)
	indexer.platform = platform
	return &indexer
}

func (this *Indexer) NewValue(tx uint32, path string, value interface{}) ccurlcommon.UnivalueInterface {
	univalue := NewUnivalue(tx, path, 0, 0, value, this)
	return univalue
}

func (this *Indexer) Store() *ccurlcommon.DB                            { return &this.store }
func (this *Indexer) Buffer() *map[string]ccurlcommon.UnivalueInterface { return &this.buffer }
func (this *Indexer) ByPath() *map[string]*orderedmap.OrderedMap        { return &this.byPath }
func (this *Indexer) ByPathFirst() *map[string]*orderedmap.Element      { return &this.byPathFirst }

//func (this *Indexer) Merkles() *map[string]*merkle.Merkle               { return &this.merkles }

func (this *Indexer) IfExists(path string) bool {
	return this.buffer[path] != nil || this.RetriveShallow(path) != nil
}

// If the access has been recorded
func (this *Indexer) CheckHistory(tx uint32, path string, ifAddToBuffer bool) ccurlcommon.UnivalueInterface {
	univalue := this.buffer[path]
	if univalue == nil { // Not in the buffer, check the datastore
		univalue = this.NewValue(tx, path, this.RetriveShallow(path)) // Make a shallow copy only by default
		if ifAddToBuffer {
			this.buffer[path] = univalue
		}
	}

	this.uniquePaths = append(this.uniquePaths, path)
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
			if !this.platform.OnControlList(parentPath) && tx != ccurlcommon.SYSTEM && err == nil {
				if parentValue := this.CheckHistory(tx, parentPath, true); parentValue != nil && parentValue.Value() != nil {
					err = parentValue.UpdateParentMeta(tx, univalue, this)
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
	return this.store.Retrive(key)
}

func (this *Indexer) Save(key string, v interface{}) {
	this.store.Save(key, v)
}

// All transitions from one traxiton
func (this *Indexer) Import(txTrans []ccurlcommon.UnivalueInterface) {
	for _, v := range txTrans {
		this.addToBuffers(v)
	}
}

func (this *Indexer) addToBuffers(v ccurlcommon.UnivalueInterface) {
	path := v.GetPath()
	if _, ok := this.byPath[path]; !ok {
		this.byPath[path] = orderedmap.NewOrderedMap()
		if initialState := this.RetriveShallow(path); initialState != nil {
			v := this.NewValue(ccurlcommon.SYSTEM, path, initialState.(ccurlcommon.TypeInterface).Deepcopy())
			this.byPath[path].Set(-1, v) //Txs are ordered by Tx, -1 will guarantee the initial state is always the first one
		}
	}
	this.byPath[path].Set(int(v.GetTx()), v)
	this.byPathFirst[path] = this.byPath[path].Front()

	// Add to the transation index
	if this.byTx[v.GetTx()] == nil {
		this.byTx[v.GetTx()] = []ccurlcommon.UnivalueInterface{}
	}
	this.byTx[v.GetTx()] = append(this.byTx[v.GetTx()], v)
}

func (this *Indexer) FinalizeStates() {
	finalizer := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			k := this.uniquePaths[i]
			if this.byPath[k] == nil {
				continue
			}

			begin := this.byPathFirst[k] // Use the first non-nil as the base state
			v := begin.Value.(ccurlcommon.UnivalueInterface)
			v.ApplyDelta(ccurlcommon.SYSTEM, begin.Next())
			if v.Value() != nil {
				v.Value().(ccurlcommon.TypeInterface).Purge() // clear transient data
			}
		}
	}
	common.ParallelWorker(len(this.uniquePaths), 6, finalizer)
}

// Only keep transation within the whitelist
func (this *Indexer) WhilteList(whitelist []uint32) []error {
	errs := []error{}
	for _, txID := range whitelist {
		if this.byTx[txID] == nil {
			errs = append(errs, errors.New("Unknown Transaction ID: "+fmt.Sprint(txID)))
		} else {
			for i := range this.byTx[txID] {
				this.byTx[txID][i] = nil
			}
		}
	}
	return errs
}

func (this *Indexer) Commit(whitelist []uint32) ([]string, interface{}, []error) {
	t0 := time.Now()
	errs := this.WhilteList(whitelist)
	fmt.Println("Whitelisting:", time.Since(t0))

	/* Get unique keys */
	t0 = time.Now()
	this.uniquePaths = make([]string, 0, len(this.byPath))
	for k := range this.byPath {
		this.uniquePaths = append(this.uniquePaths, k)
	}
	fmt.Println("UniqueString "+fmt.Sprint(100000*9), time.Since(t0))

	// Merge and finalize
	t0 = time.Now()
	this.FinalizeStates()
	fmt.Println("FinalizeStates:", time.Since(t0))

	return this.uniquePaths, &this.byPath, errs
}

// Clear all
func (this *Indexer) Clear() {
	this.buffer = make(map[string]ccurlcommon.UnivalueInterface)
	this.byTx = make(map[uint32][]ccurlcommon.UnivalueInterface)
	this.importBuffer = make(map[string]ccurlcommon.UnivalueInterface)
	this.byPath = make(map[string]*orderedmap.OrderedMap)
}

/* Map to array */
func (*Indexer) ToArray(dict *map[string]ccurlcommon.UnivalueInterface, needToSort bool) []ccurlcommon.UnivalueInterface {
	array := make([]ccurlcommon.UnivalueInterface, 0, len(*dict))
	for _, v := range *dict {
		array = append(array, v)
	}

	if needToSort { // Sort by path
		sort.SliceStable(array, func(i, j int) bool {
			return bytes.Compare([]byte(array[i].GetPath()[:]), []byte(array[j].GetPath()[:])) < 0
		})
	}
	return array
}

func (this *Indexer) Equal(other *Indexer) bool {
	cache0 := this.ToArray(&this.buffer, true)
	cache1 := other.ToArray(&this.buffer, true)
	cacheFlag := reflect.DeepEqual(cache0, cache1)
	return cacheFlag
}

func (this *Indexer) Print() {
	for i, elem := range this.ToArray(&this.buffer, true) {
		fmt.Println("Level : ", i)
		elem.Print()
	}
}
