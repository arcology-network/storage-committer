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
package noncommutative

import (
	"github.com/arcology-network/common-lib/codec"
	stgcommon "github.com/arcology-network/storage-committer/common"
)

type GrowonlySet struct {
	value [][]byte
}

func NewGrowonlySet() stgcommon.Type {
	return &GrowonlySet{
		value: make([][]byte, 0),
	}
}

func (this *GrowonlySet) New(_, delta, _, _, _ any) any {
	set := NewGrowonlySet().(*GrowonlySet)
	set.value = delta.([][]byte)
	return set
}

func (this *GrowonlySet) Clone() any { return nil }
func (this *GrowonlySet) Equal(other any) bool {
	return codec.Byteset(this.value).Equal(other.(*GrowonlySet).value)
}

func (this *GrowonlySet) IsNumeric() bool     { return false }
func (this *GrowonlySet) IsCommutative() bool { return true }
func (this *GrowonlySet) IsBounded() bool     { return false }

func (this *GrowonlySet) MemSize() uint64                            { return 0 }
func (this *GrowonlySet) CanApply(v any) bool                        { return false } // If the input has the same type as this, return true
func (this *GrowonlySet) TypeID() uint8                              { return stgcommon.UNKNOWN }
func (this *GrowonlySet) CopyTo(v any) (any, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *GrowonlySet) Value() any         { return this.value }
func (this *GrowonlySet) Delta() (any, bool) { return this.value, true }
func (this *GrowonlySet) DeltaSign() bool    { return true } // delta sign
func (this *GrowonlySet) Limits() (any, any) { return nil, nil }
func (this *GrowonlySet) SetLimits(_, _ any) {}

func (this *GrowonlySet) SetValue(v any)         { this.SetDelta(v, true) }
func (this *GrowonlySet) ResetDelta()            { this.value = [][]byte{} }
func (this *GrowonlySet) SetDelta(v any, _ bool) { this.value = v.([][]byte) }

func (this *GrowonlySet) Preload(_ string, _ any) {}
func (this *GrowonlySet) IsDeltaApplied() bool    { return true }
func (this *GrowonlySet) CloneDelta() (any, bool) { return this.Clone(), true }

func (this *GrowonlySet) Get() (any, uint32, uint32) { return this.value, 1, 0 }
func (this *GrowonlySet) Set(v any, _ any) (any, uint32, uint32, uint32, error) {
	if v == nil {
		this.value = nil
		return this, 0, 1, 0, nil
	}

	this.value = append(this.value, v.([]byte))
	return this, 0, 1, 0, nil
}

func (this *GrowonlySet) ApplyDelta(typedVals []stgcommon.Type) (stgcommon.Type, int, error) {
	for _, v := range typedVals {
		if this == nil && v != nil { // New value
			this.value = v.(*GrowonlySet).value
		}

		if this != nil && v != nil {
			this.Set(v, nil)
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil, 0, nil
	}

	return this, len(typedVals), nil
}

// Size() uint64 // Encoded size
func (this *GrowonlySet) Size() uint64   { return codec.Byteset(this.value).Size() }
func (this *GrowonlySet) Encode() []byte { return codec.Byteset(this.value).Encode() }
func (this *GrowonlySet) EncodeToBuffer(buffer []byte) int {
	return codec.Byteset(this.value).EncodeToBuffer(buffer)
}

func (this *GrowonlySet) Decode(buffer []byte) any {
	set := GrowonlySet{}
	set.value = codec.Byteset{}.Decode(buffer).(codec.Byteset)
	return &set
}

func (this *GrowonlySet) StorageEncode(_ string) []byte             { return this.Encode() }
func (this *GrowonlySet) StorageDecode(_ string, buffer []byte) any { return this.Decode(buffer) }
func (this *GrowonlySet) Hash(hasher func([]byte) []byte) []byte {
	return codec.Byteset(this.value).Hash(hasher)
}

func (this *GrowonlySet) ShortHash() (uint64, bool) { return 0, false } // For fast comparison only.
func (this *GrowonlySet) Print()                    { codec.Byteset(this.value).Print() }
