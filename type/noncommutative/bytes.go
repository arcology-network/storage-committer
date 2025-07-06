/*
 *   Copyright (c) 2024 Arcology Network

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
	"bytes"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	stgtype "github.com/arcology-network/storage-committer/common"
)

//type Bytes []byte

type Bytes struct {
	placeholder bool //
	value       codec.Bytes
}

func NewBytes(v []byte) stgtype.Type {
	b := make([]byte, len(v))
	copy(b, v)
	return &Bytes{
		placeholder: true, // To separate from nil for encoding and decoding
		value:       b,
	}
}

func (this *Bytes) Assign(v []byte) {
	this.value = v
}

func (this *Bytes) MemSize() uint64     { return uint64(1 + len(this.value)) }
func (this *Bytes) IsSelf(key any) bool { return true }
func (this *Bytes) TypeID() uint8       { return BYTES }

func (this *Bytes) CopyTo(v any) (any, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

// create a new path
func (this *Bytes) Clone() any {
	return &Bytes{
		placeholder: true,
		value:       slice.Clone(this.value),
	}
}

func (this *Bytes) Equal(other any) bool {
	return bytes.Equal(this.value, other.(*Bytes).value)
}

func (this *Bytes) IsNumeric() bool     { return false }
func (this *Bytes) IsCommutative() bool { return false }
func (this *Bytes) IsBounded() bool     { return false }

func (this *Bytes) Value() any      { return this.value }
func (this *Bytes) Delta() any      { return this.value }
func (this *Bytes) DeltaSign() bool { return true } // delta sign
func (this *Bytes) Min() any        { return nil }
func (this *Bytes) Max() any        { return nil }

func (this *Bytes) CloneDelta() any         { return codec.Bytes(slice.Clone(this.value)) }
func (this *Bytes) SetValue(v any)          { this.SetDelta(v) }
func (this *Bytes) Preload(_ string, _ any) {}

func (this *Bytes) IsDeltaApplied() bool       { return true }
func (this *Bytes) ResetDelta()                { this.SetDelta(codec.Bytes([]byte{})) }
func (this *Bytes) SetDelta(v any)             { copy(this.value, v.(codec.Bytes)) }
func (this *Bytes) SetDeltaSign(v any)         {}
func (this *Bytes) SetMin(v any)               {}
func (this *Bytes) SetMax(v any)               {}
func (this *Bytes) Get() (any, uint32, uint32) { return []byte(this.value), 1, 0 }

func (this *Bytes) New(_, delta, _, _, _ any) any {
	v := common.IfThenDo1st(delta != nil && delta.(codec.Bytes) != nil, func() codec.Bytes { return delta.(codec.Bytes).Clone().(codec.Bytes) }, this.value)
	return &Bytes{
		true,
		v,
	}
}

func (this *Bytes) Set(value any, _ any) (any, uint32, uint32, uint32, error) {
	if value != nil && this != value { // Avoid self copy.
		this.value = make([]byte, len(value.(*Bytes).value))
		copy(this.value, value.(*Bytes).value)
	}
	return this, 0, 1, 0, nil
}

func (this *Bytes) ApplyDelta(typedVals []stgtype.Type) (stgtype.Type, int, error) {
	// vec := v.([]*univalue.Univalue)
	for _, v := range typedVals {
		// v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Bytes)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*Bytes), nil)
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
