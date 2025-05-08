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
	stgintf "github.com/arcology-network/storage-committer/common"
)

// The Path type is a special commutative type that represents a path in the concurrent storage.
// It keeps track of the all the sub paths that are added, removed or updated.
type Path struct {
	*deltaset.DeltaSet[string]
	preloaded *orderedset.OrderedSet[string]

	//When it is set to non zero, it can only store one type of data, otherwise it can store multiple types.
	//This is no type checking, it is up to the developer to make sure the data type is correct.
	Type uint8
}

func NewPath(newPaths ...string) stgintf.Type {
	this := &Path{
		Type:     0, // The default data type is 0. It can store multiple types of data in the same path.
		DeltaSet: deltaset.NewDeltaSet("", 1000, nil, newPaths...),
	}
	return this
}

func (this *Path) Length() int                                                { return int(this.DeltaSet.NonNilCount()) }
func (this *Path) View() *deltaset.DeltaSet[string]                           { return this.DeltaSet }
func (this *Path) MemSize() uint64                                            { return uint64(this.DeltaSet.NonNilCount()) * 32 * 2 } // Just an estimate, need to update on fly instead of calculating everytime
func (this *Path) TypeID() uint8                                              { return PATH }
func (this *Path) IsSelf(key interface{}) bool                                { return common.IsPath(key.(string)) }
func (this *Path) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *Path) IsNumeric() bool         { return false }
func (this *Path) IsCommutative() bool     { return true }
func (this *Path) IsIdempotent(v any) bool { return false } // To expensive to check, so we just return false.
func (this *Path) IsBounded() bool         { return true }

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

	if v, err := store.Retrive(k, new(Path)); v != nil && err == nil && v.(*Path).Committed().Length() > 0 {
		this.preloaded = v.(*Path).Committed()
	}
}

// Clone creates a new Path object with the same committed and delta sets.
// the delta set is a deep copy of the original delta set, but the committed set is a shallow copy.
// This is because the committed set should never be modified until the commit time.
func (this *Path) Clone() interface{} {
	return &Path{
		DeltaSet:  this.DeltaSet.Clone(),
		preloaded: this.preloaded,
		Type:      this.Type,
	}
}

func (this *Path) Equal(other interface{}) bool { return this.DeltaSet.Equal(other.(*Path).DeltaSet) }

func (this *Path) Get() (interface{}, uint32, uint32) {
	return this.DeltaSet, 1, common.IfThen(this.DeltaSet.IsDirty(), uint32(1), uint32(0))
}

// For the codec only
func (this *Path) New(value, delta, sign, min, max interface{}) interface{} {
	deltaSet := &Path{
		DeltaSet: this.DeltaSet.CloneDelta(),
		Type:     this.Type,
	}
	return deltaSet
}

// Swap swaps two values.
func Swap[T any](lhv, rhv *T) {
	v := *lhv
	*lhv = *rhv
	*rhv = v
}

// ApplyDelta applies all the deltas from the non-conflicting transitions to the original value and returns the new value.
func (this *Path) ApplyDelta(typedVals []stgintf.Type) (stgintf.Type, int, error) {
	if idx, _ := slice.FindFirst(typedVals, nil); idx >= 0 {
		return nil, 1, nil //This is a deletion and when this is true, the number of write operations is 1.
	}

	// Due to the async nature of the importing process, the preloaded value may not be in the first element of the slice.
	// If this is the case, we need to find the preloaded value and set it to the preloaded field of the first element.
	if this.preloaded != nil {
		this.DeltaSet.SetCommitted(this.preloaded) // Set the preloaded value to the committed value so the delta set can be applied on.
	} else {
		if idx, v := slice.FindFirstIf(typedVals, func(_ int, v stgintf.Type) bool { return v.(*Path).preloaded != nil }); idx >= 0 {
			common.Swap(&this.preloaded, &(*v).(*Path).preloaded)
			this.DeltaSet.SetCommitted(this.preloaded)
		}
		// If no ones has the preloaded value, then this is a new path, no preloaded value
	}

	deltaSets := slice.Transform(typedVals, func(_ int, v stgintf.Type) *deltaset.DeltaSet[string] { return v.(*Path).DeltaSet })
	this.Commit(deltaSets) // Apply the delta sets to the committed value，including its own delta set.
	return this, len(typedVals), nil
}

// Set sets the value of the key to the given value and returns the new value, the number of keys added, the number of
// keys removed and the number of keys updated.
func (this *Path) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	targetPath := source.([]interface{})[0].(string)
	containerRoot := source.([]interface{})[1].(string)
	tx := source.([]interface{})[2].(uint64)
	writeCache := source.([]interface{})[3].(interface {
		Write(tx uint64, key string, value interface{}) (int64, error)
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

	// Update an existing key or delete a non-existent key won't change the delta set. So we return 0, 0, 0.
	// This as to be this way otherwise it will cause a lot of conflicts, when multiple transactions are trying
	// to update the same key. It is also logically correct.
	if (ok && value != nil) || (!ok && value == nil) {
		return this, 0, 0, 0, nil
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
	// this.DeltaSet.Reset()
	this.DeltaSet.ResetDelta() // The committed keys are not reset.
}

func (this *Path) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

// This function is mainly for fast comparison.
// The second return value is to tell if the first return value is 100% accurate. If it is not, then the first return value
// is just an estimate. you need to use other function to do comparison
func (this *Path) ShortHash() (uint64, bool) {
	return 0, this.DeltaSet.IsEmpty()
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
