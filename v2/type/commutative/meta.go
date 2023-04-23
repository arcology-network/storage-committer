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
	addedDict     *orderedmap.OrderedMap
	removedDict   *orderedmap.OrderedMap
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
		view:          nil,
		addedDict:     orderedmap.NewOrderedMap(),
		removedDict:   orderedmap.NewOrderedMap(),
		snapshotDirty: false,
	}

	return this, nil
}

func (this *Meta) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func (this *Meta) View() *orderedset.OrderedSet { return this.view }
func (this *Meta) IsSelf(key interface{}) bool  { return ccurlcommon.IsPath(key.(string)) }
func (this *Meta) TypeID() uint8                { return ccurlcommon.CommutativeMeta }
func (this *Meta) CommittedLength() int         { return len(this.committedKeys) }
func (this *Meta) Length() int {
	this.RefreshView()
	return int(this.view.Len())
}

// For linear access
func (this *Meta) At(idx uint64) {}

func (this *Meta) Deepcopy() interface{} {
	meta := &Meta{
		committedKeys: this.committedKeys,
		added:         common.DeepCopy(this.added),
		removed:       common.DeepCopy(this.removed),
		view:          this.view.Deepcopy(),
		addedDict:     orderedmap.NewOrderedMap(),
		removedDict:   orderedmap.NewOrderedMap(),
	}
	return meta
}

func (this *Meta) Equal(other *Meta) bool {
	return reflect.DeepEqual(this.added, other.added) &&
		reflect.DeepEqual(this.removed, other.removed)
}

func (this *Meta) ToAccess() interface{} {
	return nil
}

func (this *Meta) Get(source interface{}) (interface{}, uint32, uint32) {
	if !this.snapshotDirty { // cache clean
		return this, 1, 0
	}

	return this, 1, 1
}

func (this *Meta) Delta() interface{} {
	return &Meta{
		committedKeys: []string{},     // committed keys
		added:         this.Added(),   // Get the keys to be added from the dictionary
		removed:       this.Removed(), // Get the keys to be removed from the dictionary
		view:          this.view,
		addedDict:     this.addedDict,
		removedDict:   this.removedDict,
		snapshotDirty: this.snapshotDirty,
	}
}

// Just return the object, do nothing
func (this *Meta) Value() interface{} {
	return this
}

// committed keys + added - removed
func (this *Meta) Latest() interface{} {
	this.RefreshView()
	return &Meta{
		committedKeys: common.To(this.view.Keys(), string("")), // committed keys
	}
}

func (this *Meta) DeltaOnly() interface{} {
	return &MetaDelta{
		common.To(this.addedDict.Keys(), string("")),
		common.To(this.removedDict.Keys(), string("")),
	}
}

func (this *Meta) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	// k := common.To(this.addedDict.Keys(), string(""))
	// if !common.EqualArray(k, this.added) {
	// 	panic("")
	// }

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
			this = this.Value().(*Meta) // A new value
		}

		keys = append(keys, v.(*Meta).AddedArray().([]string)...)
		toRemove = append(toRemove, v.(*Meta).RemovedArray().([]string)...)
	}

	if this != nil {
		if len(toRemove) > 0 {
			// t0 := time.Now()
			// keys, _ = performance.RemoveString(keys, toRemove)
			toRemoveDict := make(map[string]bool)
			for _, v := range toRemove {
				toRemoveDict[v] = true
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

// Load keys into an orderedmap for quick access, only happens at once
func (this *Meta) RefreshView() {
	if this.view != nil { // Keys have been loaded already.
		return
	}
	this.view = orderedset.NewOrderedSet(this.committedKeys)

	this.view.Difference(this.removedDict)
	this.view.Union(this.addedDict)
}

// Write and afflicated operations
func (this *Meta) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	path := source.([3]interface{})[0].(string)
	tx := source.([3]interface{})[1].(uint32)
	indexer := source.([3]interface{})[2].(ccurlcommon.IndexerInterface)
	subkey := path[strings.LastIndex(path[:len(path)-1], "/")+1:] // Extract the  key

	this.RefreshView()             // Initialize the key view if has been done yet.
	ok := this.view.Exists(subkey) // If exists
	if ok && value != nil {
		return this, 1, 0, 0, nil // No meta changes, value update only
	}

	if !ok && value == nil {
		return this, 1, 0, 0, nil // Delete an non existent entry
	}

	if value == nil && ccurlcommon.IsPath(path) { // Delete the whole path
		for _, subpath := range this.Value().([]interface{}) { // Get all the sub paths
			indexer.Write(tx, path+subpath.(string), nil) // Remove all the sub paths.
		}
		return this, 0, 1, 0, nil
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
		return this, 0, 1, 0, errors.New("Error: Impossible to be in both sets!!")
	}
	return this, 0, 1, 0, nil //<<<>>> deltawrits ?????
}

func (this *Meta) includeBuffer(subkey string, value interface{}, preexists bool) bool {
	// if _, ok := this.removedDict[subkey]; ok { // Adding back a preexisting entry
	// 	delete(this.removedDict, subkey) // Cancel out each other
	// 	return true
	// }

	if _, ok := this.removedDict.Get(subkey); ok { // Adding back a preexisting entry
		this.removedDict.Delete(subkey) // Cancel out each other
		return true
	}

	if !preexists && value != nil {
		this.addedDict.Set(subkey, true) //  duplicate is ok
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
		this.removedDict.Set(subkey, true)
		return true
	}
	return this.addedDict.Delete(subkey) // Leave out the entry if it is in the added buffer
}

// data cleaning before saving to storage
func (this *Meta) Purge() {
	this.added = []string{}
	this.removed = []string{}
	this.view = nil
	this.addedDict = orderedmap.NewOrderedMap()
	this.removedDict = orderedmap.NewOrderedMap()
}

func (this *Meta) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

// Debugging interfaces
func (this *Meta) CommittedKeys() []string   { return this.committedKeys } // Check new keys
func (this *Meta) AddedArray() interface{}   { return this.added }         // Check new keys
func (this *Meta) RemovedArray() interface{} { return this.removed }       // Peek the removed keys

func (this *Meta) Added() []string                         { return common.To(this.addedDict.Keys(), "") }   // Check new keys
func (this *Meta) Removed() []string                       { return common.To(this.removedDict.Keys(), "") } // Peek the removed keys
func (this *Meta) SetCommittedKeys(committedKeys []string) { this.committedKeys = committedKeys }
func (this *Meta) SetAdded(keys []string)                  { this.added = keys }
func (this *Meta) SetRemoved(keys []string)                { this.removed = keys }
