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
	"testing"

	"github.com/holiman/uint256"
)

func TestU256(t *testing.T) {
	v := NewBoundedU256(uint256.NewInt(4), uint256.NewInt(6))
	delta := NewU256Delta(uint256.NewInt(0), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(5), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(0), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(1), true) // 6
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(1), true) // still 6
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	delta = NewU256Delta(uint256.NewInt(1), false) // 5
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	finalized, _, _ := v.(*U256).Get()
	fv := finalized.(uint256.Int)
	if fv.ToBig().Uint64() != 5 {
		t.Error("Error: Should have failed")
	}
}

func TestU256DeltaOutRange(t *testing.T) {
	v := NewBoundedU256(uint256.NewInt(40), uint256.NewInt(60))
	delta := NewU256Delta(uint256.NewInt(0), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(10), false) //  - 10
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	delta = NewU256Delta(uint256.NewInt(40), false) //  - 40
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	delta = NewU256Delta(uint256.NewInt(0), false)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(1), false)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	delta = NewU256Delta(uint256.NewInt(1), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil { // Must bring it to somewhere between the lower and upper limits
		t.Error("Error: Should have failed")
	}

	delta = NewU256Delta(uint256.NewInt(40), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(1), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(2), false)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	delta = NewU256Delta(uint256.NewInt(18), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error("Error: Should have failed")
	}

	delta = NewU256Delta(uint256.NewInt(1), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error("Error: Should have failed")
	}

	delta = NewU256Delta(uint256.NewInt(1), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	// v.(*U256).Get().(*uint256.Int).ToBig().Uint64()
	finalized, _, _ := v.(*U256).Get()
	fv := finalized.(uint256.Int)
	if fv.ToBig().Uint64() != 60 {
		t.Error("Error: Should be", 60)
	}

	delta = NewU256Delta(uint256.NewInt(1), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	finalized, _, _ = v.(*U256).Get()
	fv = finalized.(uint256.Int)
	if fv.ToBig().Uint64() != 60 {
		t.Error("Error: Should be", 60)
	}
}

func TestCloner(t *testing.T) {
	x := uint256.NewInt(100)
	delta := uint256.NewInt(1)
	ret := (&uint256.Int{}).Add(x, delta)
	if x.ToBig().Uint64() != 100 || ret.ToBig().Uint64() != 101 || delta.ToBig().Uint64() != 1 {
		t.Error("Error: Should be unchange")
	}
}

func TestCodec(t *testing.T) {
	in := NewUnboundedU256().(*U256)

	buffer := in.Encode()
	out := (&(U256{})).Decode(buffer).(*U256)
	if out.value.Uint64() != 0 ||
		out.min.Uint64() != (in.min).Uint64() ||
		(out.max).Uint64() != (in.max).Uint64() ||
		out.deltaPositive != in.deltaPositive {
		t.Error("Error: Mismatch after Encode()/Decode()")
	}

	buffer = in.Encode()
	out = (&(U256{})).Decode(buffer).(*U256)
	if (out.delta).Uint64() != (in.delta).Uint64() ||
		out.deltaPositive != in.deltaPositive ||
		(out.min).Uint64() != (in.min).Uint64() ||
		(out.max).Uint64() != (in.max).Uint64() {
		t.Error("Error: Out of range, should have failed")
	}

	in = NewBoundedU256(&U256_MIN, &U256_MAX).(*U256)

	buffer = (&U256{}).New(nil, in.delta, true, nil, nil).(*U256).Encode()
	out = (&(U256{})).Decode(buffer).(*U256)
	if !out.value.Eq((&U256_ZERO)) ||
		!out.delta.Eq(&in.delta) ||
		!out.min.Eq((&U256_MIN)) ||
		!out.max.Eq((&U256_MAX)) {
		t.Error("Error: Out of range, should have failed")
	}
}
