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
	"math"

	codec "github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	stgintf "github.com/arcology-network/storage-committer/common"
)

// type Selector []bool

type Uint64 struct {
	value uint64
	delta uint64
	min   uint64
	max   uint64
}

func NewUnboundedUint64() stgintf.Type         { return &Uint64{min: 0, max: math.MaxUint64} }
func NewUint64Delta(delta uint64) stgintf.Type { return &Uint64{delta: delta} }

func NewBoundedUint64(min, max uint64) stgintf.Type {
	if max >= min {
		return &Uint64{min: min, max: max}
	}
	return NewUnboundedUint64()
}

func (this *Uint64) Clone() interface{} { return common.New(*this) }

// For the codec only, don't use it for other purposes
func (this *Uint64) New(value, delta, _, min, max interface{}) interface{} {
	return &Uint64{
		common.IfThenDo1st(value != nil, func() uint64 { return value.(uint64) }, 0),
		common.IfThenDo1st(delta != nil, func() uint64 { return delta.(uint64) }, 0),
		common.IfThenDo1st(min != nil, func() uint64 { return min.(uint64) }, 0),
		common.IfThenDo1st(max != nil, func() uint64 { return max.(uint64) }, math.MaxUint64),
	}
}

func (this *Uint64) Equal(other interface{}) bool {
	return this.value == other.(*Uint64).value &&
		this.delta == other.(*Uint64).delta &&
		this.min == other.(*Uint64).min &&
		this.max == other.(*Uint64).max
}

func (this *Uint64) MemSize() uint32 { return 5 * 8 }

func (this *Uint64) IsNumeric() bool     { return true }
func (this *Uint64) IsCommutative() bool { return true }
func (this *Uint64) IsBounded() bool     { return this.min != 0 || this.max != math.MaxInt64 }

func (this *Uint64) Value() interface{} { return this.value }
func (this *Uint64) Delta() interface{} { return this.delta }
func (this *Uint64) DeltaSign() bool    { return true }
func (this *Uint64) Min() interface{}   { return this.min }
func (this *Uint64) Max() interface{}   { return this.max }

func (this *Uint64) Reset()                          { this.delta = 0 }
func (this *Uint64) IsDeltaApplied() bool            { return this.delta == 0 }
func (this *Uint64) CloneDelta() interface{}         { return this.delta }
func (this *Uint64) ResetDelta()                     { this.SetDelta(common.New[codec.Uint64](0)) }
func (this *Uint64) Preload(_ string, _ interface{}) {}

func (this *Uint64) SetValue(v interface{})     { this.value = v.(uint64) }
func (this *Uint64) SetDelta(v interface{})     { this.delta = v.(uint64) }
func (this *Uint64) SetDeltaSign(v interface{}) {}
func (this *Uint64) SetMin(v interface{})       { this.min = v.(uint64) }
func (this *Uint64) SetMax(v interface{})       { this.max = v.(uint64) }

func (this *Uint64) TypeID() uint8                                              { return UINT64 }
func (this *Uint64) IsSelf(key interface{}) bool                                { return true }
func (this *Uint64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *Uint64) Get() (interface{}, uint32, uint32) {
	return this.value + this.delta, 1, common.IfThen(this.delta == 0, uint32(0), uint32(1))
}

func (this *Uint64) Set(v interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if (this.max < v.(*Uint64).delta) || (this.max-v.(*Uint64).delta < this.value+this.delta) {
		return this, 0, 1, 0, errors.New("Error: Value out of range!!")
	}

	this.delta += v.(*Uint64).delta
	return this, 0, 0, 1, nil
}

func (this *Uint64) ApplyDelta(typedVals []stgintf.Type) (stgintf.Type, int, error) {
	// vec := v.([]*univalue.Univalue)
	for i, v := range typedVals {
		// v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Uint64)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			if _, _, _, _, err := this.Set(v, nil); err != nil {
				return nil, i, err
			}
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil, 0, errors.New("Error: Nil value")
	}

	this.value += this.delta
	this.delta = 0
	return this, len(typedVals), nil
}

func (this *Uint64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}
