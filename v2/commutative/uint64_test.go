package commutative

import (
	"fmt"
	"math"
	"testing"
	"time"

	codec "github.com/arcology-network/common-lib/codec"
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
	if *v.value != 0 {
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
	val := codec.Uint64(0)
	del := codec.Uint64(10)
	min := codec.Uint64(111)
	max := codec.Uint64(999)
	in := (&Uint64{}).New(&val, &del, nil, &min, &max).(*Uint64)

	buffer := in.Encode()
	out := (&Uint64{}).Decode(buffer)
	fmt.Println(in)
	fmt.Println(out.(*Uint64))

	if !in.Equal(out) {
		t.Error("Wrong value")
	}

	in = (&Uint64{}).New(&val, &del, nil, &min, &max).(*Uint64)
	buffer = in.Encode()
	out = (&Uint64{}).Decode(buffer)
	if !in.Equal(out) {
		t.Error("Wrong value")
	}
}

func TestUint64Codec2(t *testing.T) {
	val := codec.Uint64(2)
	del := codec.Uint64(10)
	min := codec.Uint64(111)
	max := codec.Uint64(999)

	in := &Uint64{&val, &del, &min, &max}

	t0 := time.Now()
	buffer := in.Encode()
	fmt.Println("Encode: ", time.Since(t0))

	t0 = time.Now()
	out := (&Uint64{}).Decode(buffer).(*Uint64)
	fmt.Println("Decode:", time.Since(t0))

	if !in.Equal(out) {
		t.Error("Wrong value")
	}

	if !in.Equal(out) {
		t.Error("Don't match")
	}

	buffer = in.Encode()
	out = (&Uint64{}).Decode(buffer).(*Uint64)
	if *(*out).value != 2 ||
		*(*out).delta != 10 ||
		*(*out).min != 111 ||
		*(*out).max != 999 {
		t.Error("Don't match")
	}

	in = &Uint64{&val, &del, nil, nil}

	buffer = in.Encode()
	out = (&Uint64{}).Decode(buffer).(*Uint64)
	if *(*out).value != 2 ||
		*(*out).delta != 10 ||
		*(*out).min != 0 ||
		*(*out).max != math.MaxUint64 {
		t.Error("Don't match")
	}

	in = (&Uint64{}).New(&val, nil, nil, nil, nil).(*Uint64)

	buffer = in.Encode()
	out = (&Uint64{}).Decode(buffer).(*Uint64)
	if *(*out).value != 2 ||
		*(*out).delta != 0 ||
		*(*out).min != 0 ||
		*(*out).max != math.MaxUint64 {
		t.Error("Don't match")
	}

	in = &Uint64{nil, nil, nil, nil}

	buffer = in.Encode()
	out = (&Uint64{}).Decode(buffer).(*Uint64)
	if *(*out).value != 0 ||
		*(*out).delta != 0 ||
		*(*out).min != 0 ||
		*(*out).max != math.MaxUint64 {
		t.Error("Don't match")
	}
}
