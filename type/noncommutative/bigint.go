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
	"math/big"

	// "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/common"
	intf "github.com/arcology-network/storage-committer/common"
)

// type Bigint codec.Bigint

type Bigint big.Int

func NewBigint(v int64) interface{} {
	var value big.Int
	value.SetInt64(v)
	this := Bigint(value)
	return &this
}

func (this *Bigint) MemSize() uint64                                            { return uint64((*big.Int)(this).BitLen()) }
func (this *Bigint) IsSelf(key interface{}) bool                                { return true }
func (this *Bigint) TypeID() uint8                                              { return BIGINT }
func (this *Bigint) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *Bigint) Equal(other interface{}) bool {
	return bytes.Equal((*big.Int)(this).Bytes(), (*big.Int)(other.(*Bigint)).Bytes())
}

func (this *Bigint) Clone() interface{} {
	v := big.Int(*this)
	return (*Bigint)(new(big.Int).Set(&v))
}

func (this *Bigint) IsNumeric() bool     { return true }
func (this *Bigint) IsCommutative() bool { return false }
func (this *Bigint) IsBounded() bool     { return false }

func (this *Bigint) Value() interface{} { return (this) }
func (this *Bigint) Delta() interface{} { return (this) }
func (this *Bigint) DeltaSign() bool    { return true } // delta sign
func (this *Bigint) Min() interface{}   { return nil }
func (this *Bigint) Max() interface{}   { return nil }

func (this *Bigint) CloneDelta() interface{} { return this.Clone() }

func (this *Bigint) SetValue(v interface{})          { this.SetDelta(v) }
func (this *Bigint) Preload(_ string, _ interface{}) {}

func (this *Bigint) IsDeltaApplied() bool       { return true }
func (this *Bigint) ResetDelta()                { this.SetDelta(big.NewInt(0)) }
func (this *Bigint) SetDelta(v interface{})     { (*big.Int)(this).Set((*big.Int)(v.(*Bigint))) }
func (this *Bigint) SetDeltaSign(v interface{}) {}
func (this *Bigint) SetMin(v interface{})       {}
func (this *Bigint) SetMax(v interface{})       {}

func (*Bigint) AffliatedDeletes() []string { return nil }

func (this *Bigint) Get() (interface{}, uint32, uint32) { return *((*big.Int)(this)), 1, 0 }

func (this *Bigint) New(_, delta, _, _, _ interface{}) interface{} {
	return common.IfThenDo1st(delta != nil && delta.(*Bigint) != nil, func() interface{} { return delta.(*Bigint).Clone() }, interface{}(this))
}

func (this *Bigint) Set(value interface{}, _ interface{}) (interface{}, uint32, uint32, uint32, error) {
	if value != nil {
		*this = *(value.(*Bigint))
	}
	return this, 0, 1, 0, nil
}

func (this *Bigint) ApplyDelta(typedVals []intf.Type) (intf.Type, int, error) {
	// vec := v.([]*univalue.Univalue)
	for _, v := range typedVals {
		// v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Bigint)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*Bigint), nil)
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
