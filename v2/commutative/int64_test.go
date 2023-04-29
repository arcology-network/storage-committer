package commutative

import (
	"math"
	"testing"
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
	in := NewInt64(0, 5).(*Int64)
	in.Set(NewInt64Delta(4), nil)
	buffer := in.Encode()
	out := (&Int64{}).Decode(buffer)
	if *(in) != *(out.(*Int64)) {
		t.Error("Wrong value")
	}

	in = NewInt64(math.MinInt64, 0).(*Int64)
	buffer = in.Encode()
	out = (&Int64{}).Decode(buffer)
	if *(in) != *(out.(*Int64)) {
		t.Error("Wrong value")
	}
}
