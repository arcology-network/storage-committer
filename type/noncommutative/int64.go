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
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	intf "github.com/arcology-network/storage-committer/common"
)

type Int64 int64

func NewInt64(v int64) *Int64 {
	var this Int64 = Int64(v)
	return &this
}

func (this *Int64) MemSize() uint64                            { return 8 }
func (this *Int64) IsSelf(key any) bool                        { return true }
func (this *Int64) TypeID() uint8                              { return INT64 }
func (this *Int64) CopyTo(v any) (any, uint32, uint32, uint32) { return v, 0, 1, 0 }
func (this *Int64) Clone() any                                 { return common.New(*this) }
func (this *Int64) Equal(other any) bool                       { return *this == *(other.(*Int64)) }
func (this *Int64) Get() (any, uint32, uint32)                 { return int64(*this), 1, 0 }

func (this *Int64) IsNumeric() bool     { return true }
func (this *Int64) IsCommutative() bool { return false }
func (this *Int64) IsBounded() bool     { return false }

func (this *Int64) Value() any      { return (this) }
func (this *Int64) Delta() any      { return (this) }
func (this *Int64) DeltaSign() bool { return true } // delta sign
func (this *Int64) Min() any        { return nil }
func (this *Int64) Max() any        { return nil }

func (this *Int64) CloneDelta() any         { return this.Clone() }
func (this *Int64) SetValue(v any)          { this.SetDelta(v) }
func (this *Int64) Preload(_ string, _ any) {}

func (this *Int64) IsDeltaApplied() bool { return true }
func (this *Int64) ResetDelta()          { this.SetDelta(common.New[Int64](0)) }
func (this *Int64) SetDelta(v any)       { *this = (*v.(*Int64)) }
func (this *Int64) SetDeltaSign(v any) {
	(*this) *= common.IfThen(v.(bool), Int64(1), Int64(-1))
}
func (this *Int64) SetMin(v any) {}
func (this *Int64) SetMax(v any) {}

func (this *Int64) New(_, delta, _, _, _ any) any {
	if common.IsType[int64](delta) {
		delta = common.New(codec.Int64(delta.(int64)))
	}

	return common.IfThenDo1st(delta != nil && delta.(*Int64) != nil, func() any { return delta.(*Int64).Clone() }, any(this))
}

func (this *Int64) Set(value any, source any) (any, uint32, uint32, uint32, error) {
	if value != nil {
		*this = Int64(*(value.(*Int64)))
	}
	return this, 0, 1, 0, nil
}

func (this *Int64) ApplyDelta(typedVals []intf.Type) (intf.Type, int, error) {
	// vec := v.([]*univalue.Univalue)
	for _, v := range typedVals {
		// v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Int64)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*Int64), nil)
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
