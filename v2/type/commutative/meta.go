package commutative

import (
	"errors"
	"fmt"
	"reflect"

	codec "github.com/arcology/common-lib/codec"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
	orderedmap "github.com/elliotchance/orderedmap"
)

type Meta struct {
	path            string // current path
	keyView         *orderedmap.OrderedMap
	addedCache      *orderedmap.OrderedMap
	removedCache    map[string]ccurlcommon.UnivalueInterface
	finalized       bool
	iterator        *orderedmap.Element
	reverseIterator *orderedmap.Element
	cacheDirty      bool

	// Export only
	keys    []string
	added   []string // added keys
	removed []string // removed keys
}

func NewMeta(path string) (interface{}, error) {
	if !ccurlcommon.IsPath(path) {
		return nil, errors.New("Error: Wrong path format !")
	}

	if !ccurlcommon.CheckDepth(path) {
		return nil, errors.New("Error: Exceeded the maximum depth")
	}

	this := &Meta{
		path:            path,
		keys:            []string{},
		added:           []string{},
		removed:         []string{},
		finalized:       false,
		keyView:         nil,
		addedCache:      orderedmap.NewOrderedMap(),
		removedCache:    make(map[string]ccurlcommon.UnivalueInterface),
		iterator:        nil,
		reverseIterator: nil,
		cacheDirty:      false,
	}
	return this, nil
}

func (this *Meta) Deepcopy() interface{} {
	var keyView *orderedmap.OrderedMap
	if this.keyView != nil {
		keyView = this.keyView.Copy()
	}
	return &Meta{
		path: this.path,
		// keys:         ccurlcommon.Deepcopy(this.keys),
		keys:         this.keys,
		added:        ccurlcommon.Deepcopy(this.added),
		removed:      ccurlcommon.Deepcopy(this.removed),
		keyView:      keyView,
		addedCache:   orderedmap.NewOrderedMap(),
		removedCache: make(map[string]ccurlcommon.UnivalueInterface),
		finalized:    this.finalized,
		cacheDirty:   false,
	}
}

func (this *Meta) Equal(other *Meta) bool {
	return this.path == other.path &&
		reflect.DeepEqual(this.keys, other.keys) &&
		reflect.DeepEqual(this.added, other.added) &&
		reflect.DeepEqual(this.removed, other.removed) &&
		this.finalized == other.finalized
}

func (this *Meta) ToAccess() interface{} {
	return nil
}

func (this *Meta) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	this.finalized = true
	if !this.cacheDirty {
		return this, 1, 0
	}
	return this, 1, 1
}

func (this *Meta) Delta(source interface{}) interface{} {
	return &Meta{
		path:            this.path,
		keys:            []string{},
		added:           this.GetAdded(),
		removed:         this.GetRemoved(),
		finalized:       this.finalized,
		keyView:         this.keyView,
		addedCache:      this.addedCache,
		removedCache:    this.removedCache,
		iterator:        this.iterator,
		reverseIterator: this.reverseIterator,
		cacheDirty:      this.cacheDirty,
	}
}

func (this *Meta) ApplyDelta(tx uint32, others []ccurlcommon.UnivalueInterface) ccurlcommon.TypeInterface {
	if len(this.removedCache) != 0 {
		panic("Error: this.removedCache should be empty !")
	}

	for _, other := range others {
		if other != nil {
			if other.Value() != nil {
				this.added = append(this.added, other.Value().(*Meta).GetAdded()...)
				for _, k := range other.Value().(*Meta).GetRemoved() {
					this.removedCache[k] = other
				}
			} else {
				return nil
			}
		}
	}

	this.keys = this.FinalizeKeySet(this.added, this.removedCache)
	return this
}

// For imported keys only
func (this *Meta) FinalizeKeySet(addedKeys []string, removalDict map[string]ccurlcommon.UnivalueInterface) []string {
	labels := make([]bool, len(this.keys))
	for i := 0; i < len(this.keys); i++ {
		_, labels[i] = removalDict[this.keys[i]]
	}

	newKeys := make([]string, 0, len(this.keys)+len(addedKeys))
	for i := 0; i < len(this.keys); i++ {
		if !labels[i] {
			newKeys = append(newKeys, this.keys[i])
		}
	}
	newKeys = append(newKeys, addedKeys...)

	this.cacheDirty = false
	return newKeys
}

func (this *Meta) GetAdded() []string {
	if this.cacheDirty {
		this.added = this.added[:0]
		for iter := this.addedCache.Front(); iter != nil; iter = iter.Next() {
			this.added = append(this.added, iter.Key.(string))
		}
	}
	return this.added
}

func (this *Meta) GetRemoved() []string {
	if this.cacheDirty {
		this.removed = this.removed[:0]
		for k := range this.removedCache {
			this.removed = append(this.removed, k)
		}
	}
	return this.removed
}

func (this *Meta) Value() interface{} {
	return this.keys
}

// Vectorize keys
func (this *Meta) GetKeys() []string {
	this.LoadKeys()
	this.GetRemoved()
	return this.FinalizeKeySet(this.GetAdded(), this.removedCache)
}

// Load keys into an orderedmap for quick access
func (this *Meta) LoadKeys() {
	if this.keyView != nil {
		return
	}

	this.keyView = orderedmap.NewOrderedMap()
	for _, k := range this.keys {
		if _, ok := this.removedCache[k]; !ok {
			this.keyView.Set(k, true)
		}
	}

	for iter := this.addedCache.Front(); iter != nil; iter = iter.Next() {
		this.keyView.Set(iter.Key, true)
	}
	this.iterator = this.addedCache.Front()
	this.reverseIterator = this.addedCache.Back()
}

func (this *Meta) Length() uint32        { return uint32(this.keyView.Len()) }
func (this *Meta) ResetIterator()        { this.iterator = this.keyView.Front() }
func (this *Meta) ResetReverseIterator() { this.reverseIterator = this.keyView.Back() }

func (this *Meta) Next() string {
	this.LoadKeys()
	if this.iterator == nil {
		return ""
	}

	key := this.iterator.Key.(string)
	this.iterator = this.iterator.Next()
	return key
}

func (this *Meta) Previous() string {
	this.LoadKeys()
	if this.reverseIterator == nil {
		return ""
	}

	key := this.reverseIterator.Key.(string)
	this.reverseIterator = this.reverseIterator.Prev()
	return key
}

func (this *Meta) Peek(source interface{}) interface{} {
	return this
}

func (this *Meta) Set(tx uint32, path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value == nil {
		indexer := source.(ccurlcommon.LocalCacheInterface)
		univalue := indexer.Read(tx, path)
		for _, subpath := range univalue.(*Meta).GetKeys() {
			indexer.Write(tx, path+subpath, nil) // Remove all the sub paths
		}
		return 0, 1, nil
	}
	return 0, 1, errors.New("Error: Path can only be created or deleted !")
}

func (this *Meta) Composite() bool       { return !this.finalized }
func (this *Meta) Path() string          { return this.path }
func (this *Meta) SetKeys(keys []string) { this.keys = keys }
func (this *Meta) Updated() bool         { return len(this.added) > 0 || len(this.removed) > 0 }
func (this *Meta) TypeID() uint8         { return ccurlcommon.CommutativeMeta }

func (this *Meta) RefreshCaches(tx uint32, subelem ccurlcommon.UnivalueInterface, source interface{}) {
	if this.keyView != nil {
		key := subelem.GetPath()
		key = key[len(this.Path()):]
		if subelem.Value() == nil {
			this.keyView.Delete(key) // Add to the key cache as well
		} else {
			this.keyView.Set(key, subelem) // Add to the key cache as well
		}
	}

	this.cacheDirty = true
	this.SaveToAddedCache(tx, subelem, source)
	this.SaveToRemovalCache(tx, subelem, source)
}

func (this *Meta) SaveToAddedCache(tx uint32, subelem ccurlcommon.UnivalueInterface, source interface{}) {
	key := subelem.GetPath()
	key = key[len(this.Path()):]

	if !subelem.Preexist() && subelem.Value() != nil { // A new Elemnet
		if _, ok := this.addedCache.Get(key); !ok {
			this.addedCache.Set(key, subelem)
		}
	}

	if subelem.Value() == nil { // Delete an Element, it is possible the element is also in the cache
		if _, ok := this.addedCache.Get(key); ok {
			this.addedCache.Delete(key)
		}
	}
}

func (this *Meta) SaveToRemovalCache(tx uint32, subelem ccurlcommon.UnivalueInterface, source interface{}) {
	key := subelem.GetPath()
	key = key[len(this.Path()):]

	if subelem.Preexist() && subelem.Value() == nil {
		if _, ok := this.removedCache[key]; !ok {
			this.removedCache[key] = subelem // Add to the deleteion list
		}
	}

	if subelem.Value() != nil { // Possible the element has been added back, remove it from the cache in this case
		if _, ok := this.addedCache.Get(key); ok {
			delete(this.removedCache, key)
		}
	}
}

func (this *Meta) Purge() {
	this.added = []string{}
	this.removed = []string{}
	this.finalized = false
	this.keyView = nil
	this.addedCache = orderedmap.NewOrderedMap()
	this.removedCache = make(map[string]ccurlcommon.UnivalueInterface)
}

func (this *Meta) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Meta) Encode() []byte {
	byteset := [][]byte{
		codec.String(this.path).Encode(),
		codec.Strings(this.keys).Encode(),
		codec.Strings(this.added).Encode(),
		codec.Strings(this.removed).Encode(),
		codec.Bool(this.finalized).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (*Meta) Decode(bytes []byte) interface{} {
	fields := codec.Byteset{}.Decode(bytes)
	return &Meta{
		path:         codec.String("").Decode(fields[0]),
		keys:         codec.Strings([]string{}).Decode(fields[1]),
		added:        codec.Strings([]string{}).Decode(fields[2]),
		removed:      codec.Strings([]string{}).Decode(fields[3]),
		finalized:    bool(codec.Bool(true).Decode(fields[4])),
		keyView:      nil,
		addedCache:   orderedmap.NewOrderedMap(),
		removedCache: make(map[string]ccurlcommon.UnivalueInterface),
		cacheDirty:   false,
	}
}

func (this *Meta) EncodeCompact() []byte {
	byteset := [][]byte{
		codec.String(this.path).Encode(),
		codec.Strings(this.keys).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Meta) DecodeCompact(bytes []byte) interface{} {
	fields := codec.Byteset{}.Decode(bytes)
	return &Meta{
		path:         codec.String("").Decode(fields[0]),
		keys:         codec.Strings([]string{}).Decode(fields[1]),
		added:        []string{},
		removed:      []string{},
		finalized:    false,
		keyView:      nil,
		addedCache:   orderedmap.NewOrderedMap(),
		removedCache: make(map[string]ccurlcommon.UnivalueInterface),
		cacheDirty:   false,
	}
}

func (this *Meta) Print() {
	fmt.Println("Path: ", this.path)
	fmt.Println("Keys: ", this.keys)
	fmt.Println("Added: ", this.added)
	fmt.Println("Removed: ", this.removed)
	fmt.Println()
}
