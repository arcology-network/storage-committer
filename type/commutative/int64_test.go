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
	"fmt"
	"math"
	"testing"
	"time"
)

func TestNewInt64(t *testing.T) {
	v := NewInt64(0, 5).(*Int64)

	v.Set(NewInt64Delta(3), nil)
	v.Set(NewInt64Delta(2), nil)

	final, _, _ := v.Get()
	if final.(int64) != 5 {
		t.Error("Wrong value")
	}

	v.Set(NewInt64Delta(3), nil)
	v.Set(NewInt64Delta(2), nil)
	final, _, _ = v.Get()
	if final.(int64) != 5 {
		t.Error("Wrong value")
	}

	v.Set(NewInt64Delta(-3), nil)
	final, _, _ = v.Get()
	if final.(int64) != 2 {
		t.Error("Wrong value")
	}

	v.Set(NewInt64Delta(1), nil)
	final, _, _ = v.Get()
	if final.(int64) != 3 {
		t.Error("Wrong value")
	}

	v.Set(NewInt64Delta(1), nil)
	final, _, _ = v.Get()
	if final.(int64) != 4 {
		t.Error("Wrong value")
	}

	v = NewInt64(-5, 0).(*Int64)
	v.Set(NewInt64Delta(-3), nil)
	v.Set(NewInt64Delta(-2), nil)

	final, _, _ = v.Get()
	if final.(int64) != -5 {
		t.Error("Wrong value")
	}

	v.Set(NewInt64Delta(-1), nil)
	v.Set(NewInt64Delta(-2), nil)
	final, _, _ = v.Get()
	if final.(int64) != -5 {
		t.Error("Wrong value")
	}
}

func TestNewInt64Limits(t *testing.T) {
	v := NewInt64(math.MinInt32, math.MaxInt32).(*Int64)
	v.Set(NewInt64Delta(1), nil)

	final, _, _ := v.Get()
	if final.(int64) != 1 {
		t.Error("Wrong value")
	}

	v.Set(NewInt64Delta(2), nil)
	if v.value != 0 {
		t.Error("Wrong value")
	}

	final, _, _ = v.Get()
	if final.(int64) != 3 {
		t.Error("Wrong value")
	}

	v.Set(NewInt64Delta(-3), nil)
	if v.value != 0 {
		t.Error("Wrong value")
	}

	final, _, _ = v.Get()
	if final.(int64) != 0 {
		t.Error("Wrong value")
	}

	v.Set(NewInt64Delta(math.MinInt32), nil)
	final, _, _ = v.Get()
	if final.(int64) != math.MinInt32 {
		t.Error("Wrong value")
	}

	// Out of the lower limit tests
	v.Set(NewInt64Delta(math.MinInt32), nil)
	final, _, _ = v.Get()
	if final.(int64) != math.MinInt32 {
		t.Error("Wrong value")
	}
}

func TestNewInt64MinMax(t *testing.T) {
	v := NewInt64(math.MinInt32, math.MaxInt32).(*Int64)
	v.Set(NewInt64Delta(math.MinInt32), nil)
	final, _, _ := v.Get()
	if final.(int64) != math.MinInt32 {
		t.Error("Error: Wrong value, should be ", math.MinInt32)
	}

	v.Set(NewInt64Delta(math.MaxInt32), nil)
	v.Set(NewInt64Delta(1), nil)
	final, _, _ = v.Get()
	if final.(int64) != 0 {
		t.Error("Error: Wrong value, should be ", 0)
	}

	v = NewInt64(math.MinInt32, math.MaxInt32).(*Int64)
	v.Set(NewInt64Delta(math.MaxInt32), nil)
	final, _, _ = v.Get()
	if final.(int64) != math.MaxInt32 {
		t.Error("Error: Wrong value, should be ", math.MaxInt32)
	}

	v.Set(NewInt64Delta(math.MaxInt32), nil)
	final, _, _ = v.Get()
	if final.(int64) != math.MaxInt32 {
		t.Error("Error: Wrong value, should be ", 0)
	}

	v.Set(NewInt64Delta(1), nil)
	final, _, _ = v.Get()
	if final.(int64) != math.MaxInt32 {
		t.Error("Error: Wrong value, should be ", 0)
	}

	v.Set(NewInt64Delta(-1), nil)
	final, _, _ = v.Get()
	if final.(int64) != math.MaxInt32-1 {
		t.Error("Error: Wrong value, should be ", 0)
	}

	v.Set(NewInt64Delta(math.MinInt32), nil)
	final, _, _ = v.Get()
	if final.(int64) != -2 {
		t.Error("Error: Wrong value, should be ", 0)
	}
}

func TestInt64Codec(t *testing.T) {
	val := int64(2)
	del := int64(10)
	min := int64(-2)
	max := int64(50)

	in := &Int64{val, del, min, max}
	buffer := in.Encode()
	out := (&Int64{}).Decode(buffer).(*Int64)
	if !in.Equal(out) {
		t.Error("Wrong value")
	}

	if out.value != in.value ||
		out.delta != in.delta ||
		out.min != in.min ||
		out.max != in.max {
		t.Error("Error: Wrong value ")
	}

	t0 := time.Now()
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	fmt.Println("Encode() + Decode(): ", time.Since(t0))

	in = (&Int64{}).New(in.Value(), in.Delta(), del >= 0, nil, nil).(*Int64)
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	if (*out).value != 2 ||
		(*out).delta != 10 ||
		(*out).min != math.MinInt64 ||
		(*out).max != math.MaxInt64 {
		t.Error("Don't match")
	}

	in = &Int64{val, int64(0), math.MinInt64, math.MaxInt64}
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	if (*out).value != 2 ||
		(*out).delta != 0 ||
		(*out).min != math.MinInt64 ||
		(*out).max != math.MaxInt64 {
		t.Error("Don't match")
	}

	in = &Int64{int64(0), int64(0), math.MinInt64, math.MaxInt64}
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	if (*out).value != 0 ||
		(*out).delta != 0 ||
		(*out).min != math.MinInt64 ||
		(*out).max != math.MaxInt64 {
		t.Error("Don't match")
	}

	in = (&Int64{}).New(in.Value(), in.Delta(), del >= 0, in.Min(), in.Max()).(*Int64)
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	if (*out).value != 0 ||
		(*out).delta != 0 ||
		(*out).min != math.MinInt64 ||
		(*out).max != math.MaxInt64 {
		t.Error("Don't match")
	}

	// if out.value != in.value ||
	// 	out.delta != in.delta ||
	// 	out.min != math.MinInt64 ||
	// 	out.max != math.MaxInt64 {
	// 	t.Error("Error: Wrong value ")
	// }
}
