package commutative

import (
	"errors"
	"reflect"

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

// Write and afflicated operations
func (this *Meta) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	targetPath := source.([]interface{})[0].(string)
	myPath := source.([]interface{})[1].(string)
	tx := source.([]interface{})[2].(uint32)
	indexer := source.([]interface{})[3].(ccurlcommon.IndexerInterface)

	if ccurlcommon.IsPath(targetPath) && len(targetPath) == len(myPath) { // Delete or rewrite the path
		if value == nil { // Delete the path and all its elements
			for _, subpath := range this.view.Keys() { // Get all the sub paths
				indexer.Write(tx, targetPath+subpath, nil) //FIXME: THIS EMITS SOME ERROR MESSAGE, DOESN't SEEM HARMFUL, BUT WEIRED
			}
			return this, 0, 1, 0, nil
		}
		return this, 0, 1, 0, errors.New("Error: Cannot rewrite a path!")
	}

	subkey := targetPath[len(targetPath)-(len(targetPath)-len(myPath)):] // Extract the sub key from the path
	ok := this.view.Exists(subkey)
	if (ok && value != nil) || (!ok && value == nil) {
		return this, 1, 0, 0, nil //value update only or delete an non existent entry
	}

	if value == nil {
		this.view.DeleteByKey(subkey) // Delete a key
	} else {
		this.view.Insert(subkey)
	}

	preexists := (*indexer.Buffer())[targetPath].Preexist()
	this.addKey(subkey, value, preexists)
	this.delKeys(subkey, value, preexists)

	return this, 0, 0, 1, nil
}

func (this *Meta) addKey(subkey string, value interface{}, preexists bool) bool {
	if this.delDict.Exists(subkey) { // Adding back a preexisting entry
		this.delDict.Delete(subkey) // Cancel out each other
		return true
	}

	if !preexists && value != nil {
		this.addDict.Insert(subkey) // won't be duplicate
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
		this.delDict.Insert(subkey)
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
