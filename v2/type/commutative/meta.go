package commutative

import (
	"reflect"
	"strings"

	orderedset "github.com/arcology-network/common-lib/container/set"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Meta struct {
	view          *orderedset.OrderedSet // committed keys + added - removed
	addDict       *orderedset.OrderedSet
	delDict       *orderedset.OrderedSet
	snapshotDirty bool
}

func NewMeta() interface{} {
	this := &Meta{
		view:          orderedset.NewOrderedSet([]string{}),
		addDict:       orderedset.NewOrderedSet([]string{}),
		delDict:       orderedset.NewOrderedSet([]string{}),
		snapshotDirty: false,
	}
	return this
}

func (this *Meta) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func (this *Meta) View() *orderedset.OrderedSet { return this.view }
func (this *Meta) IsSelf(key interface{}) bool  { return ccurlcommon.IsPath(key.(string)) }
func (this *Meta) TypeID() uint8                { return ccurlcommon.CommutativeMeta }
func (this *Meta) CommittedLength() int         { return len(this.view.Keys()) }
func (this *Meta) Length() int {
	this.RefreshView()
	return int(this.view.Len())
}

// For linear access
func (this *Meta) At(idx uint64) {}

func (this *Meta) Deepcopy() interface{} {
	meta := &Meta{
		view:    this.view.Clone(),
		addDict: this.addDict.Clone(),
		delDict: this.delDict.Clone(),
	}
	return meta
}

func (this *Meta) Equal(other *Meta) bool {
	return reflect.DeepEqual(this.addDict.Keys(), other.addDict.Keys()) &&
		reflect.DeepEqual(this.delDict.Keys(), other.delDict.Keys())
}

func (this *Meta) ToAccess() interface{} {
	return nil
}

func (this *Meta) Get(source interface{}) (interface{}, uint32, uint32) {
	if this.addDict.Len() > 0 || this.delDict.Len() > 0 { // cache clean
		return this, 1, 0
	}

	return this, 1, 1
}

func (this *Meta) Delta() interface{} {
	return &Meta{
		view:    this.view,
		addDict: this.addDict,
		delDict: this.delDict,
	}
}

// Just return the object, do nothing
func (this *Meta) Value() interface{} {
	return this
}

// committed keys + added - removed
// func (this *Meta) Latest() interface{} {
// 	this.RefreshView()
// 	return this.view.Keys()
// }

func (this *Meta) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	keys := append(this.view.Keys(), this.addDict.Keys()...)
	toRemove := this.delDict.Keys()
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

		keys = append(keys, v.(*Meta).Added()...)
		toRemove = append(toRemove, v.(*Meta).Removed()...)
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

		this.view = orderedset.NewOrderedSet(keys)
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

	this.view.Sync()
	this.view.Difference(this.delDict)
	this.view.Union(this.addDict)
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
	added := this.addKey(subkey, value, univ.Preexist())
	deleted := this.delKeys(subkey, value, univ.Preexist())

	if added != deleted {
		return this, 0, 0, 1, nil
	} else {
		if added {
			panic("Error: This is impossible 222!")
		}
	}
	panic("Error: This is impossible!")
	return this, 0, 1, 0, nil //<<<>>> deltawrits ?????
}

func (this *Meta) addKey(subkey string, value interface{}, preexists bool) bool {
	if _, ok := this.delDict.Get(subkey); ok { // Adding back a preexisting entry
		this.delDict.Delete(subkey) // Cancel out each other
		return true
	}

	if !preexists && value != nil {
		this.addDict.Set(subkey) //  duplicate is ok
		return true
	}
	return false
}

// Only the preexisting keys are in this buffer, or they will be cancel each other
func (this *Meta) delKeys(subkey string, value interface{}, preexists bool) bool {
	if value != nil {
		return false
	}

	if preexists { // Preexists and value == nil
		this.delDict.Set(subkey)
		return true
	}
	return this.addDict.Delete(subkey) // Leave out the entry if it is in the added buffer
}

// data cleaning before saving to storage
func (this *Meta) Purge() {
	this.view = orderedset.NewOrderedSet([]string{})
	this.addDict = orderedset.NewOrderedSet([]string{})
	this.delDict = orderedset.NewOrderedSet([]string{})
}

func (this *Meta) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

// Debugging interfaces
func (this *Meta) SubDirs() []string { return this.view.Keys() }    // Check new keys
func (this *Meta) Added() []string   { return this.addDict.Keys() } // Check new keys
func (this *Meta) Removed() []string { return this.delDict.Keys() } // Peek the removed keys

func (this *Meta) SetSubDirs(keys []string) { this.view = orderedset.NewOrderedSet(keys) }
func (this *Meta) SetAdded(keys []string)   { this.addDict = orderedset.NewOrderedSet(keys) }
func (this *Meta) SetRemoved(keys []string) { this.delDict = orderedset.NewOrderedSet(keys) }
