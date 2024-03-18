/*
 *   Copyright (c) 2023 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package commutative

import (
	"errors"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/orderedset"
	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/interfaces"
)

type Path struct {
	*deltaset.DeltaSet[string]
	preloaded *orderedset.OrderedSet[string]
}

func NewPath() intf.Type {
	this := &Path{
		DeltaSet: deltaset.NewDeltaSet("", 1000),
	}
	return this
}

func InitNewPaths(newPaths []string) *Path {
	return &Path{
		DeltaSet: deltaset.NewDeltaSet("", 1000, newPaths...),
	}
}

func (this *Path) Length() int                                                { return int(this.DeltaSet.NonNilCount()) }
func (this *Path) View() *deltaset.DeltaSet[string]                           { return this.DeltaSet }
func (this *Path) MemSize() uint32                                            { return uint32(this.DeltaSet.NonNilCount()) * 32 * 2 } // Just an estimate, need to update on fly instead of calculating everytime
func (this *Path) TypeID() uint8                                              { return PATH }
func (this *Path) IsSelf(key interface{}) bool                                { return common.IsPath(key.(string)) }
func (this *Path) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *Path) IsNumeric() bool     { return false }
func (this *Path) IsCommutative() bool { return true }
func (this *Path) IsBounded() bool     { return true }

func (this *Path) Value() interface{} { return this.DeltaSet.Committed() }
func (this *Path) Delta() interface{} { return this.DeltaSet.Delta() }
func (this *Path) DeltaSign() bool    { return true }
func (this *Path) Min() interface{}   { return nil }
func (this *Path) Max() interface{}   { return nil }

func (this *Path) CloneDelta() interface{} { return this.DeltaSet.CloneDelta() }

func (this *Path) IsDeltaApplied() bool       { return this.IsDirty() }
func (this *Path) SetValue(v interface{})     { this.DeltaSet = v.(*deltaset.DeltaSet[string]) }
func (this *Path) ResetDelta()                { this.DeltaSet.ResetDelta() }
func (this *Path) SetDelta(v interface{})     { this.DeltaSet.SetDelta(v.(*deltaset.DeltaSet[string])) }
func (this *Path) SetDeltaSign(v interface{}) {}
func (this *Path) SetMin(v interface{})       {}
func (this *Path) SetMax(v interface{})       {}

func (this *Path) Preload(k string, arg interface{}) {
	if this.preloaded != nil { // Already preloaded
		return
	}

	store := arg.(interface {
		Retrive(string, any) (interface{}, error)
	})

	if v, err := store.Retrive(k, new(deltaset.DeltaSet[string])); v != nil && err == nil && v.(*Path).Committed().Length() > 0 {
		this.preloaded = v.(*Path).Committed()
	}
}

func (this *Path) Clone() interface{} {
	return &Path{
		DeltaSet:  this.DeltaSet.Clone(),
		preloaded: this.preloaded,
	}
}
func (this *Path) Equal(other interface{}) bool { return this.DeltaSet.Equal(other.(*Path).DeltaSet) }

func (this *Path) Get() (interface{}, uint32, uint32) {
	// If the value is touched, there will a write operation associated with it.
	return this.DeltaSet, 1, common.IfThen(!this.DeltaSet.IsDirty(), uint32(0), uint32(1))
	// return this.DeltaSet.Keys(), 1, common.IfThen(!this.DeltaSet.Touched(), uint32(0), uint32(1))
}

// For the codec only
func (this *Path) New(value, delta, sign, min, max interface{}) interface{} {
	deltaSet := &Path{
		DeltaSet: this.DeltaSet.CloneDelta(),
	}
	return deltaSet
}

// ApplyDelta applies all the deltas from the non-conflicting transitions to the original value and returns the new value.
func (this *Path) ApplyDelta(typedVals []intf.Type) (intf.Type, int, error) {
	if idx, _ := slice.FindFirst(typedVals, nil); idx >= 0 {
		return nil, 1, nil //This is a deletion and when this is true, the number of write operations is 1.
	}

	// Due to the async nature of the importing process, the preloaded value may not be in the first element of the slice.
	// If this is the case, we need to find the preloaded value and set it to the preloaded field of the first element.
	if this.preloaded != nil {
		this.DeltaSet.SetCommitted(this.preloaded)
	} else {
		if idx, v := slice.FindFirstIf(typedVals, func(v intf.Type) bool { return v.(*Path).preloaded != nil }); idx >= 0 {
			typedVals[idx].(*Path).preloaded = nil
			this.preloaded = (*v).(*Path).preloaded
		}
	}

	deltaSets := slice.Transform(typedVals, func(_ int, v intf.Type) *deltaset.DeltaSet[string] { return v.(*Path).DeltaSet })
	this.Commit(deltaSets...)
	return this, len(typedVals), nil
}

// Set sets the value of the key to the given value and returns the new value, the number of keys added, the number of
// keys removed and the number of keys updated.
func (this *Path) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	targetPath := source.([]interface{})[0].(string)
	containerRoot := source.([]interface{})[1].(string)
	tx := source.([]interface{})[2].(uint32)
	writeCache := source.([]interface{})[3].(interface {
		Write(tx uint32, key string, value interface{}) (int64, error)
		InCache(path string) (interface{}, bool)
	})

	// Delete or rewrite the path. A rewrite is generally not allowed.
	// The path is the root of the container. It cannot be rewritten. But it can be deleted.
	// When that happens, all the sub paths are also deleted.
	if common.IsPath(targetPath) && len(targetPath) == len(containerRoot) {
		if value == nil { // Delete the path and all its elements
			for _, subpath := range this.DeltaSet.Elements() { // Get all the committed sub paths
				// Delete the sub path
				writeCache.Write(tx, targetPath+subpath, nil) //FIXME: THIS EMITS SOME ERROR MESSAGEES BUT DON't SEEM TO BE HARMFUL
			}
			return this, 0, 1, 0, nil
		}
		return this, 0, 1, 0, errors.New("Error: Cannot rewrite a path!")
	}

	subkey := targetPath[len(targetPath)-(len(targetPath)-len(containerRoot)):] // Extract the sub key from the path
	ok, _ := this.DeltaSet.Exists(subkey)

	if (ok && value != nil) || (!ok && value == nil) {
		return this, 1, 0, 0, nil //value update only or delete an non existent entry
	}

	if value == nil {
		// Delete an existing key
		this.DeltaSet.Delete(subkey)
	} else {
		// Insert a new key
		this.DeltaSet.Insert(subkey)
	}
	return this, 0, 0, 1, nil
}

// data cleaning before saving to storage
func (this *Path) Reset() {
	this.DeltaSet.ResetDelta() // The committed keys are not reset.
}

func (this *Path) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

// For Debug
func (this *Path) SetSubPaths(keys []string)   { this.DeltaSet.InsertCommitted(keys) }
func (this *Path) SetAdded(keys []string)      { this.DeltaSet.InsertUpdated(keys) }
func (this *Path) InsertRemoved(keys []string) { this.DeltaSet.InsertRemoved(keys) }

func (this *Path) Keys() []string { // Committed keys
	return common.IfThenDo1st(this.DeltaSet.Committed() != nil, func() []string { return this.DeltaSet.Committed().Elements() }, []string{})
}

func (this *Path) Added() []string {
	return common.IfThenDo1st(this.DeltaSet.Updated() != nil, func() []string { return this.DeltaSet.Updated().Elements() }, []string{})
}

func (this *Path) Removed() []string {
	return common.IfThenDo1st(this.DeltaSet.Removed() != nil, func() []string { return this.DeltaSet.Removed().Elements() }, []string{})
}
