package commutative

import (
	"math"
	"testing"
)

func TestNewUint64(t *testing.T) {
	v := NewUint64(0, 8).(*Uint64)

	final, _, _ := v.Get()
	if final.(uint64) != 0 {
		t.Error("Wrong value")
	}

	v.Set(NewUint64Delta(5), nil)
	v.Set(NewUint64Delta(1), nil)
	v.Set(NewUint64Delta(1), nil)
	v.Set(NewUint64Delta(1), nil)
	v.Set(NewUint64Delta(1), nil)

	final, _, _ = v.Get()
	if final.(uint64) != 8 {
		t.Error("Wrong value")
	}

	v = NewUint64(0, 8).(*Uint64)

	final, _, _ = v.Get()
	if final.(uint64) != 0 {
		t.Error("Wrong value")
	}

	v.Set(NewUint64Delta(10), nil)
	if v.value != 0 {
		t.Error("Wrong value")
	}

	final, _, _ = v.Get()
	if final.(uint64) != 0 {
		t.Error("Wrong value")
	}
}

func TestNewUint64Max(t *testing.T) {
	v := NewUint64(0, math.MaxUint64).(*Uint64)

	v.Set(NewUint64Delta(math.MaxUint64-1), nil)
	v.Set(NewUint64Delta(2), nil)

	final, _, _ := v.Get()
	if final.(uint64) != math.MaxUint64-1 {
		t.Error("Error: Wrong value")
	}

	// Overflow test
	v = NewUint64(0, math.MaxUint64).(*Uint64)
	v.Set(NewUint64Delta(math.MaxUint64-1), nil)
	v.Set(NewUint64Delta(1), nil)
	v.Set(NewUint64Delta(1), nil)

	final, _, _ = v.Get()
	if final.(uint64) != math.MaxUint64 {
		t.Error("Wrong value")
	}

	// Overflow test
	v = NewUint64(0, math.MaxUint64).(*Uint64)
	v.Set(NewUint64Delta(math.MaxUint64-1), nil)
	v.Set(NewUint64Delta(math.MaxUint64), nil)

	final, _, _ = v.Get()
	if final.(uint64) != math.MaxUint64-1 {
		t.Error("Wrong value")
	}

	v = NewUint64(0, math.MaxUint64).(*Uint64)
	v.Set(NewUint64Delta(math.MaxUint64), nil)
	if _, _, _, _, err := v.Set(NewUint64Delta(math.MaxUint64), nil); err == nil {
		t.Error("Error: Should report an overflow")
	}

	final, _, _ = v.Get()
	if final.(uint64) != math.MaxUint64 {
		t.Error("Wrong value")
	}
}

func TestUint64Codec(t *testing.T) {
	in := NewUint64(5, 0)
	buffer := in.(*Uint64).Encode()
	out := (&Uint64{}).Decode(buffer)
	if *(in.(*Uint64)) != *(out.(*Uint64)) {
		t.Error("Wrong value")
	}

	in = NewUint64(0, 0)
	buffer = in.(*Uint64).Encode()
	out = (&Uint64{}).Decode(buffer)
	if *(in.(*Uint64)) != *(out.(*Uint64)) {
		t.Error("Wrong value")
	}
}
