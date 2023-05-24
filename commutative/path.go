package commutative

import (
	"errors"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
)

type Path struct {
	value *orderedset.OrderedSet // committed keys + added - removed
	delta *PathDelta
}

func NewPath() interface{} {
	this := &Path{
		value: orderedset.NewOrderedSet([]string{}),
		delta: NewPathDelta([]string{}, []string{}),
	}
	return this
}

func (this *Path) Length() int                                                { return this.value.Length() }
func (this *Path) View() *orderedset.OrderedSet                               { return this.value }
func (this *Path) MemSize() uint32                                            { return codec.Strings(this.value.Keys()).Size() * 2 } // Just an estimate, need to update on fly instead of calculating everytime
func (this *Path) TypeID() uint8                                              { return PATH }
func (this *Path) IsSelf(key interface{}) bool                                { return ccurlcommon.IsPath(key.(string)) }
func (this *Path) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *Path) IsNumeric() bool     { return false }
func (this *Path) IsCommutative() bool { return true }

func (this *Path) Value() interface{} { return this.value }
func (this *Path) Delta() interface{} { return this.delta }
func (this *Path) DeltaSign() bool    { return true }
func (this *Path) Min() interface{}   { return nil }
func (this *Path) Max() interface{}   { return nil }

func (this *Path) Clone() interface{} {
	meta := &Path{
		value: this.value.Clone().(*orderedset.OrderedSet),
		delta: this.delta.Clone().(*PathDelta),
	}
	return meta
}

func (this *Path) ReInit() {
	this.value = common.IfThen(this.value == nil, orderedset.NewOrderedSet([]string{}), this.value)
	this.delta = common.IfThen(this.delta == nil, NewPathDelta([]string{}, []string{}), this.delta)
}

func (this *Path) Equal(other interface{}) bool {
	return common.EqualIf(this.value, other.(*Path).value, func(v0, v1 *orderedset.OrderedSet) bool { return v0.Equal(v1) }, func(v *orderedset.OrderedSet) bool { return len(v.Keys()) == 0 }) &&
		common.EqualIf(this.delta, other.(*Path).delta, func(v0, v1 *PathDelta) bool { return v0.Equal(v1) }, func(v *PathDelta) bool { return len(v.Added()) == 0 && len(v.Removed()) == 0 })
}

func (this *Path) Get() (interface{}, uint32, uint32) {
	return this.value.Keys(), 1, common.IfThen(!this.value.Touched(), uint32(0), uint32(1))
}

// For the codec only
func (this *Path) New(value, delta, sign, min, max interface{}) interface{} {
	return &Path{
		value: common.IfThenDo1st(value != nil && value.(*orderedset.OrderedSet) != nil && len(value.(*orderedset.OrderedSet).Keys()) > 0, func() *orderedset.OrderedSet { return value.(*orderedset.OrderedSet) }, nil),
		delta: common.IfThenDo1st(delta != nil && delta.(*PathDelta) != nil && delta.(*PathDelta).Touched(), func() *PathDelta { return delta.(*PathDelta) }, nil),
	}
}

func (this *Path) ApplyDelta(v interface{}) (ccurlcommon.TypeInterface, int, error) { // Apply the transitions to the original value
	this.ReInit()
	toAdd := this.delta.addDict.Keys() // The value should only contain committed keys
	toRemove := this.delta.Removed()
	univals := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(univals); i++ {
		if univals[i].GetPath() == nil { // Not in the whitelist
			continue
		}

		if univals[i].Value() == nil { // Deletion
			return nil, 0, nil
		}

		delta := univals[i].Value().(ccurlcommon.TypeInterface).Delta().(*PathDelta)
		toAdd = append(toAdd, delta.Added()...)
		toRemove = append(toRemove, delta.Removed()...)
	}

	keys := append(this.Keys(), toAdd...)
	if len(toRemove) > 0 {
		dict := make(map[string]bool)
		for _, v := range toRemove {
			dict[v] = true
		}

		common.RemoveIf(&keys, func(v string) bool {
			_, ok := dict[v]
			return ok
		})
	}

	return &Path{
		orderedset.NewOrderedSet(keys), // committed keys + added - removed
		NewPathDelta([]string{}, []string{}),
	}, len(univals), nil
}

// Write and afflicated operations
func (this *Path) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	targetPath := source.([]interface{})[0].(string)
	myPath := source.([]interface{})[1].(string)
	tx := source.([]interface{})[2].(uint32)
	writeCache := source.([]interface{})[3].(ccurlcommon.WriteCacheInterface)

	if ccurlcommon.IsPath(targetPath) && len(targetPath) == len(myPath) { // Delete or rewrite the path
		if value == nil { // Delete the path and all its elements
			for _, subpath := range this.value.Keys() { // Get all the sub paths
				writeCache.Write(tx, targetPath+subpath, nil) //FIXME: THIS EMITS SOME ERROR MESSAGEES BUT DON't SEEM HARMFUL
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

	preexists := (*writeCache.Cache())[targetPath].Preexist()
	this.delta.ProcessKey(subkey, value, preexists)
	return this, 0, 0, 1, nil
}

// data cleaning before saving to storage
func (this *Path) Reset() {
	this.value = orderedset.NewOrderedSet([]string{})
	this.delta = NewPathDelta([]string{}, []string{})
}

func (this *Path) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

// For Debug
func (this *Path) SetSubs(keys []string)    { this.value = orderedset.NewOrderedSet(keys) }
func (this *Path) SetAdded(keys []string)   { this.delta.addDict = orderedset.NewOrderedSet(keys) }
func (this *Path) SetRemoved(keys []string) { this.delta.delDict = orderedset.NewOrderedSet(keys) }

func (this *Path) Keys() []string {
	return common.IfThenDo1st(this.value != nil, func() []string { return this.value.Keys() }, []string{})
}
func (this *Path) Added() []string {
	return common.IfThenDo1st(this.value != nil, func() []string { return this.delta.Added() }, []string{})
}
func (this *Path) Removed() []string {
	return common.IfThenDo1st(this.value != nil, func() []string { return this.delta.Removed() }, []string{})
}
