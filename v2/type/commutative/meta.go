package commutative

import (
	"errors"
	"reflect"
	"strings"

	common "github.com/arcology-network/common-lib/common"

	// performance "github.com/arcology-network/common-lib/mhasher"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	orderedmap "github.com/elliotchance/orderedmap"
)

type Meta struct {
	committedKeys   []string               // committed keys
	keyView         *orderedmap.OrderedMap // committed keys + added - removed
	addedBuffer     *orderedmap.OrderedMap
	removedBuffer   map[string]interface{}
	finalized       bool
	iterator        *orderedmap.Element
	reverseIterator *orderedmap.Element
	snapshotDirty   bool

	snapshot *Meta // keyView in an array

	// Export only
	added   []string // added keys in the current block
	removed []string // removed keys in the current block
}

func NewMeta(path string) (interface{}, error) {
	if !ccurlcommon.IsPath(path) {
		return nil, errors.New("Error: Wrong path format !")
	}

	if !ccurlcommon.CheckDepth(path) {
		return nil, errors.New("Error: Exceeded the maximum depth")
	}

	this := &Meta{
		committedKeys:   []string{},
		added:           []string{},
		removed:         []string{},
		finalized:       false,
		keyView:         nil,
		addedBuffer:     orderedmap.NewOrderedMap(),
		removedBuffer:   make(map[string]interface{}),
		iterator:        nil,
		reverseIterator: nil,
		snapshotDirty:   false,
		snapshot:        nil,
	}

	return this, nil
}

// For linear access
func (this *Meta) ResetIterator()        { this.iterator = this.keyView.Front() }
func (this *Meta) ResetReverseIterator() { this.reverseIterator = this.keyView.Back() }

func (this *Meta) Deepcopy() interface{} {
	var keyView *orderedmap.OrderedMap
	if this.keyView != nil {
		keyView = this.keyView.Copy()
	}

	return &Meta{
		committedKeys: this.committedKeys,
		added:         common.DeepCopy(this.added),
		removed:       common.DeepCopy(this.removed),
		keyView:       keyView,
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: make(map[string]interface{}),
		finalized:     this.finalized,
		snapshotDirty: false,
	}
}

func (this *Meta) Equal(other *Meta) bool {
	return reflect.DeepEqual(this.committedKeys, other.committedKeys) &&
		reflect.DeepEqual(this.added, other.added) &&
		reflect.DeepEqual(this.removed, other.removed) &&
		this.finalized == other.finalized
}

func (this *Meta) ToAccess() interface{} {
	return nil
}

func (this *Meta) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	this.finalized = true
	if !this.snapshotDirty { // cache clean
		return this, 1, 0
	}

	return this, 1, 1
}

func (this *Meta) Delta(source interface{}) interface{} {
	return &Meta{
		committedKeys:   []string{}, // committed keys
		added:           this.Added(),
		removed:         this.Removed(),
		finalized:       this.finalized,
		keyView:         this.keyView,
		addedBuffer:     this.addedBuffer,
		removedBuffer:   this.removedBuffer,
		iterator:        this.iterator,
		reverseIterator: this.reverseIterator,
		snapshotDirty:   this.snapshotDirty,
	}
}

func (this *Meta) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	keys := append(this.committedKeys, this.added...)
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
			// t0 := time.Now()
			// keys, _ = performance.RemoveString(keys, toRemove)
			toRemoveDict := make(map[string]struct{})
			for _, v := range toRemove {
				toRemoveDict[v] = struct{}{}
			}
			next := 0
			for i := 0; i < len(keys); i++ {
				if _, ok := toRemoveDict[keys[i]]; ok {
					continue
				} else {
					keys[next] = keys[i]
					next++
				}
			}
			keys = keys[:next]
			// fmt.Println("RemoveBytes ", time.Since(t0))
		}

		this.committedKeys = keys
		this.snapshotDirty = false
	}
	//fmt.Println("ApplyDelta :", time.Since(t0))

	if this == nil {
		return nil
	}
	return this
}

// Check new keys
func (this *Meta) Added() []string {
	added := []string{}
	for iter := this.addedBuffer.Front(); iter != nil; iter = iter.Next() {
		added = append(added, iter.Key.(string))
	}
	return added
}

// Check removed keys
func (this *Meta) Removed() []string {
	removed := []string{}
	for k := range this.removedBuffer {
		removed = append(removed, k)
	}
	return removed
}

func (this *Meta) Erase() []string {
	return this.Snapshot().committedKeys
}

func (this *Meta) VectorizeView() []string {
	array := make([]string, 0, this.keyView.Len())
	for iter := this.keyView.Front(); iter != nil; iter = iter.Next() {
		if iter.Value != nil {
			array = append(array, (iter.Key.(string)))
		}
	}
	return array
}

// committed keys + added - removed
func (this *Meta) Value() interface{} {
	this.InitKeyView()
	return this.Snapshot().committedKeys
}

func (this *Meta) Snapshot() *Meta {
	if this.snapshotDirty || this.snapshot == nil { // Remake the cache is dirty of snapshot is empty
		this.snapshot = &Meta{
			committedKeys: this.VectorizeView(),
		}
	}
	return this.snapshot
}

func (this *Meta) Next() string {
	this.InitKeyView()
	if this.iterator == nil {
		return ""
	}

	key := this.iterator.Key.(string)
	this.iterator = this.iterator.Next()
	return key
}

func (this *Meta) Previous() string {
	this.InitKeyView()
	if this.reverseIterator == nil {
		return ""
	}

	key := this.reverseIterator.Key.(string)
	this.reverseIterator = this.reverseIterator.Prev()
	return key
}

// Just return the object, won't do anything
func (this *Meta) This(source interface{}) interface{} {
	return this
}

func (this *Meta) Reset(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	panic("Error: This function should never be called!!")
	// return 0, 1, nil
}

func (this *Meta) Composite() bool { return !this.finalized }
func (this *Meta) TypeID() uint8   { return ccurlcommon.CommutativeMeta }

// Load keys into an orderedmap for quick access, only happens at once
func (this *Meta) InitKeyView() {
	if this.keyView != nil { // Keys have been loaded already.
		return
	}
	this.keyView = orderedmap.NewOrderedMap()

	counter := 0
	for _, k := range this.committedKeys { // Committed keys first
		if _, ok := this.removedBuffer[k]; !ok {
			this.keyView.Set(k, counter) // Not in the removed set
			counter++
		}
	}

	for iter := this.addedBuffer.Front(); iter != nil; iter = iter.Next() {
		this.keyView.Set(iter.Key, counter) // Newly added keys
		counter++
	}
	this.iterator = this.addedBuffer.Front()
	this.reverseIterator = this.addedBuffer.Back()
}

// Write and afflicated operations
func (this *Meta) Set(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	tx := source.([2]interface{})[0].(uint32)
	indexer := source.([2]interface{})[1].(ccurlcommon.IndexerInterface)

	if value == nil { // Remove all the elements in the path completely
		for _, subpath := range this.Value().([]string) {
			indexer.Write(tx, path+subpath, nil, false) // Remove all the sub paths.
		}
		return 0, 1, nil
	}
	return 0, 1, errors.New("Error: A path can only be created or deleted, it cannot be rewritten!")
}

// Preexist => Delete => Readd ==
func (this *Meta) Refresh(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	subkey := path[strings.LastIndex(path[:len(path)-1], "/")+1:] // Extract the element key

	if this.keyView != nil {
		if value == nil {
			this.keyView.Delete(subkey) // Delete a key
		} else {
			this.keyView.Set(subkey, 0)                              // Add a new one if not exists
			if _, preexists := this.keyView.Get(subkey); preexists { // Check if the entry exists already
				return 0, 0, nil // Updated a pre existing element
			}
		}
	}
	indexer := source.([2]interface{})[1].(ccurlcommon.IndexerInterface)
	univ, _ := (*indexer.Buffer())[path]

	addFlag := this.toAddedBuffer(subkey, value, univ.Preexist())
	removedFlag := this.toRemovedBuffer(subkey, value, univ.Preexist())
	this.snapshotDirty = addFlag || removedFlag // Either is dirty

	if removedFlag == addFlag && removedFlag {
		return 0, 0, errors.New("Error: Impossible to be in both sets!!")
	}
	return 0, 1, nil
}

func (this *Meta) toAddedBuffer(subkey string, value interface{}, preexists bool) bool {
	if _, ok := this.removedBuffer[subkey]; ok { // Adding back a preexisting entry
		delete(this.removedBuffer, subkey) // Cancel out each other
		return true
	}

	if !preexists && value != nil {
		this.addedBuffer.Set(subkey, true) //  duplicate is ok
		return true
	}
	return false
}

// Only the preexisting keys are in this buffer, or they will be cancel each other
func (this *Meta) toRemovedBuffer(subkey string, value interface{}, preexists bool) bool {
	if value != nil {
		return false
	}

	if preexists { // Preexists and value == nil
		this.removedBuffer[subkey] = 1 // Add to the deleteion list, duplicate is ok
		return true
	}
	return this.addedBuffer.Delete(subkey) // Leave out the entry if it is in the added buffer
}

// data cleaning before saving to storage
func (this *Meta) Purge() {
	this.added = []string{}
	this.removed = []string{}
	this.finalized = false
	this.keyView = nil
	this.addedBuffer = orderedmap.NewOrderedMap()
	this.removedBuffer = make(map[string]interface{})
}

func (this *Meta) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

// Debugging interfaces
// func (this *Meta) SetKeys(keys []string)    { this.keys = keys }
// func (this *Meta) SetAdded(keys []string)   { this.added = keys }
// func (this *Meta) SetRemoved(keys []string) { this.removed = keys }
