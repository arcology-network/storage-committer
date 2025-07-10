/*
*   Copyright (c) 2025 Arcology Network

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
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommon "github.com/arcology-network/storage-committer/common"
)

type GrowOnlySet struct {
	value [][]byte
}

func NewGrowOnlySet(args ...[]byte) stgcommon.Type {
	this := &GrowOnlySet{
		value: make([][]byte, 0),
	}
	this.value = append(this.value, args...)
	return this
}

func (this *GrowOnlySet) New(_, delta, _, _, _ any) any {
	set := NewGrowOnlySet().(*GrowOnlySet)
	set.value = delta.([][]byte)
	return set
}

func (this *GrowOnlySet) Clone() any {
	if this == nil {
		return nil
	}
	return slice.Clone(this.value) // Clone the slice to avoid modifying the original
}

func (this *GrowOnlySet) Equal(other any) bool {
	return codec.Byteset(this.value).Equal(other.(*GrowOnlySet).value)
}

func (this *GrowOnlySet) IsNumeric() bool     { return false }
func (this *GrowOnlySet) IsCommutative() bool { return true }
func (this *GrowOnlySet) IsBounded() bool     { return false }

func (this *GrowOnlySet) MemSize() uint64                            { return 0 }
func (this *GrowOnlySet) CanApply(v any) bool                        { return false } // If the input has the same type as this, return true
func (this *GrowOnlySet) TypeID() uint8                              { return GROWONLY_SET }
func (this *GrowOnlySet) CopyTo(v any) (any, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *GrowOnlySet) Value() any         { return this.value }
func (this *GrowOnlySet) Delta() (any, bool) { return this.value, true }
func (this *GrowOnlySet) DeltaSign() bool    { return true } // delta sign
func (this *GrowOnlySet) Limits() (any, any) { return nil, nil }
func (this *GrowOnlySet) SetLimits(_, _ any) {}

func (this *GrowOnlySet) SetValue(v any) { this.SetDelta(v, true) }
func (this *GrowOnlySet) ResetDelta()    { this.value = [][]byte{} }
func (this *GrowOnlySet) SetDelta(v any, _ bool) {
	this.value = v.([][]byte)
}

func (this *GrowOnlySet) Preload(_ string, _ any) {}
func (this *GrowOnlySet) IsDeltaApplied() bool    { return true }
func (this *GrowOnlySet) CloneDelta() (any, bool) { return this.Clone(), true }

func (this *GrowOnlySet) Get() (any, uint32, uint32) { return this.value, 1, 0 }
func (this *GrowOnlySet) Set(v any, _ any) (any, uint32, uint32, uint32, error) {
	if v == nil {
		this.value = nil
		return this, 0, 1, 0, nil
	}

	// bytes := v.(*GrowOnlySet).Value().([][]byte)
	this.value = append(this.value, v.(stgcommon.Type).Value().([][]byte)...)
	return this, 0, 1, 0, nil
}

func (this *GrowOnlySet) ApplyDelta(otherSets []stgcommon.Type) (stgcommon.Type, int, error) {
	for _, v := range otherSets {
		if this == nil && v != nil { // New value
			this = &GrowOnlySet{}
			this.value = v.(*GrowOnlySet).value
		}

		if this != nil && v != nil {
			this.Set(v, nil)
		}
	}
	return this, len(otherSets), nil
}

// Size() uint64 // Encoded size
func (this *GrowOnlySet) Size() uint64   { return codec.Byteset(this.value).Size() }
func (this *GrowOnlySet) Encode() []byte { return codec.Byteset(this.value).Encode() }
func (this *GrowOnlySet) EncodeToBuffer(buffer []byte) int {
	return codec.Byteset(this.value).EncodeToBuffer(buffer)
}

func (this *GrowOnlySet) Decode(buffer []byte) any {
	set := GrowOnlySet{}
	set.value = codec.Byteset{}.Decode(buffer).(codec.Byteset)
	return &set
}

func (this *GrowOnlySet) StorageEncode(_ string) []byte             { return this.Encode() }
func (this *GrowOnlySet) StorageDecode(_ string, buffer []byte) any { return this.Decode(buffer) }
func (this *GrowOnlySet) Hash(hasher func([]byte) []byte) []byte {
	return codec.Byteset(this.value).Hash(hasher)
}

func (this *GrowOnlySet) ShortHash() (uint64, bool) { return 0, false } // For fast comparison only.
func (this *GrowOnlySet) Print()                    { codec.Byteset(this.value).Print() }
