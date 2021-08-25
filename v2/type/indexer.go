package urltype

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/HPISTechnologies/common-lib/common"
	merkle "github.com/HPISTechnologies/common-lib/merkle"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	"github.com/elliotchance/orderedmap"
)

type Indexer struct {
	store   ccurlcommon.DB
	buffer  map[string]ccurlcommon.UnivalueInterface
	merkles map[string]*merkle.Merkle

	byTx         map[uint32]*map[string]ccurlcommon.UnivalueInterface
	byPath       map[string][]ccurlcommon.UnivalueInterface
	byAcct       *orderedmap.OrderedMap
	baseStates   map[string]ccurlcommon.UnivalueInterface
	importBuffer map[string]ccurlcommon.UnivalueInterface
}

func NewIndexer(store ccurlcommon.DB) *Indexer {
	var indexer Indexer
	indexer.store = store
	indexer.buffer = make(map[string]ccurlcommon.UnivalueInterface, 1024)
	indexer.merkles = make(map[string]*merkle.Merkle)

	indexer.byTx = make(map[uint32]*map[string]ccurlcommon.UnivalueInterface)
	indexer.byPath = make(map[string][]ccurlcommon.UnivalueInterface)
	indexer.byAcct = orderedmap.NewOrderedMap()

	indexer.baseStates = make(map[string]ccurlcommon.UnivalueInterface)
	indexer.importBuffer = make(map[string]ccurlcommon.UnivalueInterface)
	return &indexer
}

func (this *Indexer) NewValue(tx uint32, path string, value interface{}) ccurlcommon.UnivalueInterface {
	univalue := NewUnivalue(tx, path, 0, 0, value, this)
	return univalue
}

func (this *Indexer) Store() *ccurlcommon.DB                            { return &this.store }
func (this *Indexer) Buffer() *map[string]ccurlcommon.UnivalueInterface { return &this.buffer }
func (this *Indexer) Merkles() *map[string]*merkle.Merkle               { return &this.merkles }

func (this *Indexer) IfExists(path string) bool {
	return this.buffer[path] != nil || this.Retrive(path) != nil
}

// If the access has been recorded
func (this *Indexer) CheckHistory(tx uint32, path string, ifSave bool) ccurlcommon.UnivalueInterface {
	univalue := this.buffer[path]
	if univalue == nil { // Not in the buffer, check the datastore
		univalue = this.NewValue(tx, path, this.Retrive(path))
		if ifSave {
			this.buffer[path] = univalue
		}
	}
	return univalue
}

func (this *Indexer) Read(tx uint32, path string) interface{} {
	univalue := this.CheckHistory(tx, path, true)
	return univalue.Get(tx, path, this.Buffer())
}

func (this *Indexer) TryRead(tx uint32, path string) interface{} {
	univalue := this.CheckHistory(tx, path, false)
	return univalue.Peek(this.Buffer())
}

func (this *Indexer) Write(tx uint32, path string, value interface{}) error {
	parentPath := ccurlcommon.GetParentPath(path)
	if this.IfExists(parentPath) || tx == ccurlcommon.SYSTEM { // The parent path exists or inject paths directly
		univalue := this.CheckHistory(tx, path, true)
		return univalue.Set(tx, path, value, this)
	}
	return errors.New("Error: The fe path doesn't exist: " + parentPath)
}

func (this *Indexer) Insert(path string, value interface{}) {
	this.buffer[path] = value.(ccurlcommon.UnivalueInterface)
}

func (this *Indexer) Retrive(key string) interface{} {
	if v := this.store.Retrive(key); v != nil {
		return v.(ccurlcommon.TypeInterface).Deepcopy()
	}
	return nil
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
	if value := this.Retrive(v.GetPath()); value != nil { // Get the initial states
		this.baseStates[v.GetPath()] = this.NewValue(ccurlcommon.SYSTEM, v.GetPath(), value)
	} else {
		if v.GetValue() == nil { // Tried to delete non-existent elements or the directly INJECTED paths
			return
		}
	}

	if this.byTx[v.GetTx()] == nil {
		txMap := make(map[string]ccurlcommon.UnivalueInterface)
		this.byTx[v.GetTx()] = &txMap
	}

	(*this.byTx[v.GetTx()])[v.GetPath()] = v
	this.byPath[v.GetPath()] = append(this.byPath[v.GetPath()], v)

	path := v.GetPath()
	parent := (&ccurlcommon.Platform{}).Eth10Account()
	pos := ccurlcommon.SubpathOf(parent, path)
	if pos >= 0 {
		pathDict, ok := this.byAcct.Get(path[:pos])
		if !ok {
			this.byAcct.Set(path[:pos], orderedmap.NewOrderedMap())
			pathDict, _ = this.byAcct.Get(path[:pos])
		}
		pathDict.(*orderedmap.OrderedMap).Set(path, true)

		// Merkle tree
		if this.merkles[path[:pos]] == nil {
			this.merkles[path[:pos]] = merkle.NewMerkle(8, merkle.Sha256)
		}
	}
}

func (this *Indexer) Commit(whitelist []uint32) ([]string, []interface{}, []error) {
	t0 := time.Now()
	errs := []error{}
	whitelistDict := make(map[uint32]bool, 64)
	for _, txID := range whitelist {
		if this.byTx[txID] == nil {
			errs = append(errs, errors.New("Unknown Transaction ID: "+fmt.Sprint(txID)))
		} else {
			whitelistDict[txID] = true
		}
	}

	// Keep the whitelisted entries only
	for txID, txs := range this.byTx {
		if _, ok := whitelistDict[txID]; !ok {
			for k := range *txs {
				(*txs)[k] = nil
			}
		}
	}
	fmt.Println("Whitelisting:", time.Since(t0))

	t0 = time.Now()
	// Find the new elements having no initial values
	newTrans := make([]ccurlcommon.UnivalueInterface, 0, len(this.byPath))
	for k := range this.byPath {
		if this.baseStates[k] == nil {
			for j, tran := range this.byPath[k] {
				if tran != nil {
					newTrans = append(newTrans, tran) // Use the first non-nil element as the initial value
					this.byPath[k][j] = nil           // Remove it from the buffer
					break
				}
			}
		}
	}
	fmt.Println("Find the new elements having no initial values:", time.Since(t0))

	// Add the initial values back in
	for _, tran := range newTrans {
		this.baseStates[tran.GetPath()] = tran
	}

	fmt.Println("Find the new elements having no initial values:", time.Since(t0))

	// Merge and finalize
	t0 = time.Now()
	this.GetFinalStates()
	fmt.Println("GetFinalStates:", time.Since(t0))

	t0 = time.Now()
	// Strip access info
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		this.Presist()
		defer wg.Done()
	}()

	wg.Add(1)
	go func() {
		this.ComputeMerkle()
		defer wg.Done()
	}()
	wg.Wait()
	fmt.Println("Presist + ComputeMerkle:", time.Since(t0))
	this.clear()
	return nil, nil, errs
}

func (this *Indexer) GetFinalStates() {
	keys := make([]string, 0, len(this.byPath))
	for k := range this.byPath {
		keys = append(keys, k)
	}

	finalizer := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			k := keys[i]
			v := this.baseStates[k]
			for _, tran := range this.byPath[k] {
				if tran == nil {
					continue
				}

				if err := v.ApplyDelta(ccurlcommon.SYSTEM, tran); err != nil {
					panic(err)
				}
			}
			v.Finalize()
		}
	}
	common.ParallelWorker(len(keys), 4, finalizer)
}

// Purge data before persisting to the storage
func (this *Indexer) Presist() {
	keys := make([]string, 0, len(this.byPath))
	for k := range this.byPath {
		keys = append(keys, k)
	}

	paths := make([]string, 0, len(this.baseStates))
	states := make([]interface{}, 0, len(this.baseStates))
	for _, k := range keys {
		if len(k) == 0 || this.baseStates[k] == nil {
			continue
		}

		paths = append(paths, k)
		v := this.baseStates[k].GetValue()
		if v != nil {
			v.(ccurlcommon.TypeInterface).Purge()
		}
		states = append(states, v)
	}
	t0 := time.Now()
	this.store.BatchSave(paths, states)
	fmt.Println("BatchSave():", time.Since(t0))
}

// Build a Merkle for every updated account
func (this *Indexer) ComputeMerkle() []string {
	uniqueAccts := this.byAcct.Keys()

	hasher := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			subDict, ok := this.byAcct.Get(uniqueAccts[i])
			if !ok {
				continue
			}

			// Encode individual transitions
			data := make([][]byte, 0, subDict.(*orderedmap.OrderedMap).Len())
			for _, k := range subDict.(*orderedmap.OrderedMap).Keys() {
				if v := this.baseStates[k.(string)]; v != nil {
					if v.GetValue() != nil {
						data = append(data, v.GetValue().(ccurlcommon.TypeInterface).EncodeCompact())
					} else {
						data = append(data, []byte{})
					}
				}
			}
			this.merkles[uniqueAccts[i].(string)].Init(data)
		}
	}
	common.ParallelWorker(len(uniqueAccts), 6, hasher)

	accounts := make([]string, 0, len(uniqueAccts))
	for _, k := range uniqueAccts {
		if this.baseStates[k.(string)] != nil {
			accounts = append(accounts, k.(string))
		}
	}
	return accounts
}

// Clear all
func (this *Indexer) clear() {
	this.buffer = make(map[string]ccurlcommon.UnivalueInterface)
	this.byTx = make(map[uint32]*map[string]ccurlcommon.UnivalueInterface)
	this.byPath = make(map[string][]ccurlcommon.UnivalueInterface)
	this.byAcct = orderedmap.NewOrderedMap()
	this.baseStates = make(map[string]ccurlcommon.UnivalueInterface)
	this.importBuffer = make(map[string]ccurlcommon.UnivalueInterface)
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
