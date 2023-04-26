package commutative

import (
	"errors"

	orderedset "github.com/arcology-network/common-lib/container/set"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Meta struct {
	value         *orderedset.OrderedSet // committed keys + added - removed
	addDict       *orderedset.OrderedSet
	delDict       *orderedset.OrderedSet
	delta         *MetaDelta
	snapshotDirty bool
}

func NewMeta() interface{} {
	this := &Meta{
		value:         orderedset.NewOrderedSet([]string{}),
		delta:         NewMetaDelta([]string{}, []string{}),
		addDict:       orderedset.NewOrderedSet([]string{}),
		delDict:       orderedset.NewOrderedSet([]string{}),
		snapshotDirty: false,
	}
	return this
}

func (this *Meta) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func (this *Meta) View() *orderedset.OrderedSet { return this.value }
func (this *Meta) IsSelf(key interface{}) bool  { return ccurlcommon.IsPath(key.(string)) }
func (this *Meta) TypeID() uint8                { return ccurlcommon.CommutativeMeta }
func (this *Meta) CommittedLength() int         { return len(this.value.Keys()) }
func (this *Meta) Length() int {
	return int(this.value.Len())
}

// For linear access
func (this *Meta) At(idx uint64) {}

func (this *Meta) Deepcopy() interface{} {
	meta := &Meta{
		value:   this.value.Clone(),
		addDict: this.addDict.Clone(),
		delDict: this.delDict.Clone(),
	}
	return meta
}

func (this *Meta) Equal(other *Meta) bool {
	return this.value.Equal(other.value) && this.addDict.Equal(other.addDict) && this.delDict.Equal(other.delDict)
}

func (this *Meta) ToAccess() interface{} {
	return nil
}

func (this *Meta) Get(source interface{}) (interface{}, uint32, uint32) {
	snapshot := &Meta{
		value: orderedset.NewOrderedSet(this.value.Keys()),
	}

	if !this.addDict.Touched() && !this.delDict.Touched() { // cache clean
		return this, 1, 0 // No key was added or deleted, read only
	}
	return snapshot, 1, 1 //
}

func (this *Meta) Value() interface{} { return this.value.Keys() }
func (this *Meta) Delta() interface{} {
	return &Meta{
		value:   this.value,
		addDict: this.addDict,
		delDict: this.delDict,
	}
}

// Just return the object, do nothing

func (this *Meta) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	keys := append(this.value.Keys(), this.addDict.Keys()...)
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

		this.value = orderedset.NewOrderedSet(keys)
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
			for _, subpath := range this.value.Keys() { // Get all the sub paths
				indexer.Write(tx, targetPath+subpath, nil) //FIXME: THIS EMITS SOME ERROR MESSAGE, DOESN't SEEM HARMFUL, BUT WEIRED
			}
			return this, 0, 1, 0, nil
		}
		return this, 0, 1, 0, errors.New("Error: Cannot rewrite a path!")
	}

	subkey := targetPath[len(targetPath)-(len(targetPath)-len(myPath)):] // Extract the sub key from the path
	ok := this.value.Exists(subkey)
	if (ok && value != nil) || (!ok && value == nil) {
		return this, 1, 0, 0, nil //value update only or delete an non existent entry
	}

	if value == nil {
		this.value.DeleteByKey(subkey) // Delete a key
	} else {
		this.value.Insert(subkey)
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
	this.value = orderedset.NewOrderedSet([]string{})
	this.addDict = orderedset.NewOrderedSet([]string{})
	this.delDict = orderedset.NewOrderedSet([]string{})
}

func (this *Meta) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Meta) Added() []string { // Check new keys
	if this.addDict == nil {
		return []string{}
	}
	return this.addDict.Keys()
}

func (this *Meta) Removed() []string {
	if this.addDict == nil {
		return []string{}
	}
	return this.delDict.Keys()
} // Peek the removed keys

func (this *Meta) SetSubDirs(keys []string) { this.value = orderedset.NewOrderedSet(keys) }
func (this *Meta) SetAdded(keys []string)   { this.addDict = orderedset.NewOrderedSet(keys) }
func (this *Meta) SetRemoved(keys []string) { this.delDict = orderedset.NewOrderedSet(keys) }
