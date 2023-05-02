package commutative

import (
	"fmt"
	"math"
	"testing"
	"time"

	codec "github.com/arcology-network/common-lib/codec"
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
	if *v.value != 0 {
		t.Error("Wrong value")
	}

	final, _, _ = v.Get()
	if final.(int64) != 3 {
		t.Error("Wrong value")
	}

	v.Set(NewInt64Delta(-3), nil)
	if *v.value != 0 {
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
	val := codec.Int64(2)
	del := codec.Int64(10)
	min := codec.Int64(-2)
	max := codec.Int64(50)

	in := &Int64{&val, &del, &min, &max}
	buffer := in.Encode()
	out := (&Int64{}).Decode(buffer).(*Int64)
	if !in.Equal(out) {
		t.Error("Wrong value")
	}

	if *out.value != *in.value ||
		*out.delta != *in.delta ||
		*out.min != *in.min ||
		*out.max != *in.max {
		t.Error("Error: Wrong value ")
	}

	t0 := time.Now()
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	fmt.Println("Encode() + Decode(): ", time.Since(t0))

	in = (&Int64{}).New(in.Value(), in.Delta(), del >= 0, nil, nil).(*Int64)
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	if *(*out).value != 2 ||
		*(*out).delta != 10 ||
		*(*out).min != math.MinInt64 ||
		*(*out).max != math.MaxInt64 {
		t.Error("Don't match")
	}

	in = &Int64{&val, nil, nil, nil}
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	if *(*out).value != 2 ||
		*(*out).delta != 0 ||
		*(*out).min != math.MinInt64 ||
		*(*out).max != math.MaxInt64 {
		t.Error("Don't match")
	}

	in = &Int64{nil, nil, nil, nil}
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	if *(*out).value != 0 ||
		*(*out).delta != 0 ||
		*(*out).min != math.MinInt64 ||
		*(*out).max != math.MaxInt64 {
		t.Error("Don't match")
	}

	in = (&Int64{}).New(in.Value(), in.Delta(), del >= 0, in.Min(), in.Max()).(*Int64)
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer).(*Int64)
	if *(*out).value != 0 ||
		*(*out).delta != 0 ||
		*(*out).min != math.MinInt64 ||
		*(*out).max != math.MaxInt64 {
		t.Error("Don't match")
	}

	// if out.value != in.value ||
	// 	out.delta != in.delta ||
	// 	out.min != math.MinInt64 ||
	// 	out.max != math.MaxInt64 {
	// 	t.Error("Error: Wrong value ")
	// }
}
