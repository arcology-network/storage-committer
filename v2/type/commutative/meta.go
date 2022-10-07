package commutative

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	codec "github.com/HPISTechnologies/common-lib/codec"
	common "github.com/HPISTechnologies/common-lib/common"
	performance "github.com/HPISTechnologies/common-lib/mhasher"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	orderedmap "github.com/elliotchance/orderedmap"
)

type Meta struct {
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
	return reflect.DeepEqual(this.keys, other.keys) &&
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
		keys:            []string{},
		added:           this.PeekAdded(),
		removed:         this.PeekRemoved(),
		finalized:       this.finalized,
		keyView:         this.keyView,
		addedCache:      this.addedCache,
		removedCache:    this.removedCache,
		iterator:        this.iterator,
		reverseIterator: this.reverseIterator,
		cacheDirty:      this.cacheDirty,
	}
}

func (this *Meta) ApplyDelta(tx uint32, v interface{}) ccurlcommon.TypeInterface {
	keys := append(this.keys, this.added...)
	toRemove := this.removed
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		if vec[i].GetPath() == nil { // Not in the whitelist
			continue
		}

		v := vec[i].Value()
		if v == nil { // Deletion
			keys = keys[:0]
			toRemove = toRemove[:0]
			this = nil
			continue
		}

		if this == nil {
			this = this.Value().(*Meta)
		}

		keys = append(keys, v.(*Meta).added...)
		toRemove = append(toRemove, v.(*Meta).removed...)
	}

	if this != nil {
		if len(toRemove) > 0 {
			t0 := time.Now()
			keys, _ = performance.RemoveString(keys, toRemove)
			fmt.Println("RemoveBytes ", time.Since(t0))
		}

		this.keys = keys
		this.cacheDirty = false
	}
	//fmt.Println("ApplyDelta :", time.Since(t0))

	if this == nil {
		return nil
	}
	return this
}

func (this *Meta) PeekAdded() []string {
	added := []string{}
	for iter := this.addedCache.Front(); iter != nil; iter = iter.Next() {
		added = append(added, iter.Key.(string))
	}
	return added
}

func (this *Meta) PeekRemoved() []string {
	removed := []string{}
	for k := range this.removedCache {
		removed = append(removed, k)
	}
	return removed
}

// Vectorize keys
func (this *Meta) PeekKeys() []string {
	this.LoadKeys()
	newKeys := make([]string, 0, this.keyView.Len())
	for iter := this.keyView.Front(); iter != nil; iter = iter.Next() {
		if iter.Value != nil {
			newKeys = append(newKeys, (iter.Key.(string)))
		}
	}
	return newKeys
}

func (this *Meta) Value() interface{} {
	return this.PeekKeys()
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
		indexer := source.(ccurlcommon.IndexerInterface)
		univalue := indexer.Read(tx, path)
		for _, subpath := range univalue.(*Meta).PeekKeys() {
			indexer.Write(tx, path+subpath, nil) // Remove all the sub paths
		}
		return 0, 1, nil
	}
	return 0, 1, errors.New("Error: Path can only be created or deleted !")
}

func (this *Meta) Composite() bool          { return !this.finalized }
func (this *Meta) SetKeys(keys []string)    { this.keys = keys }
func (this *Meta) SetAdded(keys []string)   { this.added = keys }   // Debug only
func (this *Meta) SetRemoved(keys []string) { this.removed = keys } // Debug only
func (this *Meta) TypeID() uint8            { return ccurlcommon.CommutativeMeta }

func (this *Meta) UpdateCaches(tx uint32, child ccurlcommon.UnivalueInterface, source interface{}) bool {
	if this.keyView != nil {
		key := *child.GetPath()
		subkey := key[strings.LastIndex(key[:len(key)-1], "/")+1:] // Extract the sub key

		if child.Value() == nil {
			this.keyView.Delete(subkey) // Add to the key cache as well
		} else {
			this.keyView.Set(subkey, child) // Add to the key cache as well
		}
	}
	this.cacheDirty = this.saveToAddedCache(tx, child, source) || this.saveToRemovalCache(tx, child, source)
	return this.cacheDirty
}

func (this *Meta) saveToAddedCache(tx uint32, child ccurlcommon.UnivalueInterface, source interface{}) bool {
	key := *child.GetPath()
	subkey := key[strings.LastIndex(key[:len(key)-1], "/")+1:] // Extract the sub key

	dirtyFlag := false
	if !child.Preexist() && child.Value() != nil { // A new Elemnet
		if _, ok := this.addedCache.Get(subkey); !ok {
			this.addedCache.Set(subkey, child)
			dirtyFlag = true
		}
	}

	if child.Value() == nil { // Delete an Element, it is possible the element is also in the cache
		if _, ok := this.addedCache.Get(subkey); ok {
			this.addedCache.Delete(subkey)
		}
	}
	return dirtyFlag
}

func (this *Meta) saveToRemovalCache(tx uint32, child ccurlcommon.UnivalueInterface, source interface{}) bool {
	key := *child.GetPath()
	subkey := key[strings.LastIndex(key[:len(key)-1], "/")+1:] // Extract the sub key

	dirtyFlag := false
	if child.Preexist() && child.Value() == nil {
		if _, ok := this.removedCache[subkey]; !ok {
			this.removedCache[subkey] = child // Add to the deleteion list
			dirtyFlag = true
		}
	}

	if child.Value() != nil { // Possible the element has been added back, remove it from the cache in this case
		if _, ok := this.addedCache.Get(subkey); ok {
			delete(this.removedCache, subkey)
		}
	}
	return dirtyFlag
}

// data cleaning before saving to storage
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
	this.keys = this.keys[:0] // Clear keys, no need to send

	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Meta) HeaderSize() uint32 {
	return 5 * codec.UINT32_LEN
}

func (this *Meta) Size() uint32 {
	if this == nil {
		return 0
	}

	total := this.HeaderSize() +
		codec.Strings(this.keys).Size() +
		codec.Strings(this.added).Size() +
		codec.Strings(this.removed).Size() +
		uint32(codec.Bool(this.finalized).Size())
	return total
}

func (this *Meta) FillHeader(buffer []byte) {
	total := uint32(0)
	codec.Uint32(4).EncodeToBuffer(buffer[codec.UINT32_LEN*0:])

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*1:])
	total += codec.Strings(this.keys).Size()

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*2:])
	total += codec.Strings(this.added).Size()

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*3:])
	total += codec.Strings(this.removed).Size()

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*4:])
}

func (this *Meta) EncodeToBuffer(buffer []byte) {
	this.FillHeader(buffer)
	headerLen := this.HeaderSize()

	offset := uint32(0)
	codec.Strings(this.keys).EncodeToBuffer(buffer[headerLen+offset:])
	offset += codec.Strings(this.keys).Size()

	codec.Strings(this.added).EncodeToBuffer(buffer[headerLen+offset:])
	offset += codec.Strings(this.added).Size()

	codec.Strings(this.removed).EncodeToBuffer(buffer[headerLen+offset:])
	offset += codec.Strings(this.removed).Size()

	codec.Bool(this.finalized).EncodeToBuffer(buffer[headerLen+offset:])
}

func (this *Meta) Decode(bytes []byte) interface{} {
	buffers := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	this = &Meta{
		keys:         codec.Strings([]string{}).Decode(common.ArrayCopy(buffers[0])).(codec.Strings),
		added:        codec.Strings([]string{}).Decode(common.ArrayCopy(buffers[1])).(codec.Strings),
		removed:      codec.Strings([]string{}).Decode(buffers[2]).(codec.Strings),
		finalized:    bool(codec.Bool(true).Decode(buffers[3]).(codec.Bool)),
		keyView:      nil,
		addedCache:   orderedmap.NewOrderedMap(),
		removedCache: make(map[string]ccurlcommon.UnivalueInterface),
		cacheDirty:   false,
	}
	return this
}

func (this *Meta) EncodeCompact() []byte {
	byteset := [][]byte{
		codec.Strings(this.keys).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Meta) DecodeCompact(bytes []byte) interface{} {
	buffers := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	return &Meta{
		keys:         codec.Strings([]string{}).Decode(buffers[0]).(codec.Strings),
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
	fmt.Println("Keys: ", this.keys)
	fmt.Println("Added: ", this.added)
	fmt.Println("Removed: ", this.removed)
	fmt.Println()
}
