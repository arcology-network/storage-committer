package commutative

import (
	"errors"
	"reflect"
	"strings"

	common "github.com/arcology-network/common-lib/common"

	orderedset "github.com/arcology-network/common-lib/container/set"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	orderedmap "github.com/elliotchance/orderedmap"
)

type Meta struct {
	committedKeys []string               // committed keys
	view          *orderedset.OrderedSet // committed keys + added - removed
	addedBuffer   *orderedmap.OrderedMap
	removedBuffer *orderedmap.OrderedMap
	finalized     bool
	snapshotDirty bool

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
		committedKeys: []string{},
		added:         []string{},
		removed:       []string{},
		finalized:     false,
		view:          nil,
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: orderedmap.NewOrderedMap(),
		snapshotDirty: false,
	}

	return this, nil
}

func (this *Meta) View() *orderedset.OrderedSet { return this.view }
func (this *Meta) IsSelf(key interface{}) bool  { return ccurlcommon.IsPath(key.(string)) }
func (this *Meta) DeltaWritable() bool          { return !this.finalized }
func (this *Meta) TypeID() uint8                { return ccurlcommon.CommutativeMeta }
func (this *Meta) CommittedLength() int         { return len(this.committedKeys) }
func (this *Meta) Length() int {
	this.InitView()
	return int(this.view.Len())
}

// For linear access
func (this *Meta) At(idx uint64) {}

func (this *Meta) Deepcopy() interface{} {
	return &Meta{
		committedKeys: this.committedKeys,
		added:         common.DeepCopy(this.added),
		removed:       common.DeepCopy(this.removed),
		view:          this.view.Deepcopy(),
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: orderedmap.NewOrderedMap(),
		finalized:     this.finalized,
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

func (this *Meta) Delta() interface{} {
	return &Meta{
		committedKeys: []string{}, // committed keys
		added:         this.Added(),
		removed:       this.Removed(),
		finalized:     this.finalized,
		view:          this.view,
		addedBuffer:   this.addedBuffer,
		removedBuffer: this.removedBuffer,
		snapshotDirty: this.snapshotDirty,
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
	return common.To(this.addedBuffer.Keys(), "")
}

// Peek the removed keys
func (this *Meta) Removed() []string {
	return common.To(this.removedBuffer.Keys(), "")
}

// committed + added - removed
func (this *Meta) Keys() []interface{} {
	this.InitView()
	return this.view.Keys()
}

// committed keys + added - removed
func (this *Meta) Value() interface{} {
	this.InitView()
	return this.Keys()
}

// Just return the object, won't do anything
func (this *Meta) This(source interface{}) interface{} {
	return this
}

func (this *Meta) Reset(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	panic("Error: This function should never be called!!")
	// return 0, 1, nil
}

// Load keys into an orderedmap for quick access, only happens at once
func (this *Meta) InitView() {
	if this.view != nil { // Keys have been loaded already.
		return
	}
	this.view = orderedset.NewOrderedSet(this.committedKeys)

	this.view.Difference(this.removedBuffer)
	this.view.Union(this.addedBuffer)
}

// Write and afflicated operations
func (this *Meta) Set(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	tx := source.([2]interface{})[0].(uint32)
	indexer := source.([2]interface{})[1].(ccurlcommon.IndexerInterface)
	subkey := path[strings.LastIndex(path[:len(path)-1], "/")+1:] // Extract the  key

	this.InitView()                // Initialize the key view if has been done yet.
	ok := this.view.Exists(subkey) // If exists
	if ok && value != nil {
		return 0, 0, nil // No meta changes, value update only
	}

	if !ok && value == nil {
		return 0, 0, nil // Delete an non existent entry
	}

	if value == nil && ccurlcommon.IsPath(path) { // Delete the whole path
		for _, subpath := range this.Value().([]string) { // Get all the sub paths
			indexer.Write(tx, path+subpath, nil, false) // Remove all the sub paths.
		}
		return 0, 1, nil
	}

	if value == nil {
		this.view.DeleteByKey(subkey) // Delete a key
	} else {
		this.view.Insert(subkey)
	}

	univ, _ := (*indexer.Buffer())[path]
	addFlag := this.includeBuffer(subkey, value, univ.Preexist())
	removedFlag := this.excludeBuffer(subkey, value, univ.Preexist())
	this.snapshotDirty = addFlag || removedFlag // Either is dirty

	if removedFlag == addFlag && removedFlag {
		return 0, 0, errors.New("Error: Impossible to be in both sets!!")
	}
	return 0, 1, nil
}

func (this *Meta) includeBuffer(subkey string, value interface{}, preexists bool) bool {
	// if _, ok := this.removedBuffer[subkey]; ok { // Adding back a preexisting entry
	// 	delete(this.removedBuffer, subkey) // Cancel out each other
	// 	return true
	// }

	if _, ok := this.removedBuffer.Get(subkey); ok { // Adding back a preexisting entry
		this.removedBuffer.Delete(subkey) // Cancel out each other
		return true
	}

	if !preexists && value != nil {
		this.addedBuffer.Set(subkey, true) //  duplicate is ok
		return true
	}
	return false
}

// Only the preexisting keys are in this buffer, or they will be cancel each other
func (this *Meta) excludeBuffer(subkey string, value interface{}, preexists bool) bool {
	if value != nil {
		return false
	}

	if preexists { // Preexists and value == nil
		this.removedBuffer.Set(subkey, true)
		return true
	}
	return this.addedBuffer.Delete(subkey) // Leave out the entry if it is in the added buffer
}

// data cleaning before saving to storage
func (this *Meta) Purge() {
	this.added = []string{}
	this.removed = []string{}
	this.finalized = false
	this.view = nil
	this.addedBuffer = orderedmap.NewOrderedMap()
	this.removedBuffer = orderedmap.NewOrderedMap()
}

func (this *Meta) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

// Debugging interfaces
// func (this *Meta) SetKeys(keys []string)    { this.keys = keys }
// func (this *Meta) SetAdded(keys []string)   { this.added = keys }
// func (this *Meta) SetRemoved(keys []string) { this.removed = keys }
