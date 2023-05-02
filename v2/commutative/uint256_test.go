package commutative

import (
	"testing"

	"github.com/holiman/uint256"
)

func TestMaxUint256(t *testing.T) {
	max := uint256.NewInt(0).SetAllOne()
	t.Log(max)
}

func TestU256(t *testing.T) {
	v := NewU256(uint256.NewInt(4), uint256.NewInt(6))
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
	if finalized.(*uint256.Int).ToBig().Uint64() != 5 {
		t.Error("Error: Should have failed")
	}
}

func TestU256DeltaOutRange(t *testing.T) {
	v := NewU256(uint256.NewInt(40), uint256.NewInt(60))
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

}

func TestCodec(t *testing.T) {
	in := NewU256(U256MIN, U256MIN).(*U256)

	buffer := in.Encode()
	out := (&(U256{})).Decode(buffer).(*U256)
	if out.value.Uint64() != 0 ||
		(*out.min).Uint64() != (*in.min).Uint64() ||
		(*out.max).Uint64() != (*in.max).Uint64() ||
		out.deltaPositive != in.deltaPositive {
		t.Error("Error: Mismatch after Encode()/Decode()")
	}

	buffer = in.Encode()
	out = (&(U256{})).Decode(buffer).(*U256)
	if (*out.delta).Uint64() != (*in.delta).Uint64() ||
		out.deltaPositive != in.deltaPositive ||
		(*out.min).Uint64() != (*in.min).Uint64() ||
		(*out.max).Uint64() != (*in.max).Uint64() {
		t.Error("Error: Out of range, should have failed")
	}

	in = NewU256(U256MIN, U256MAX).(*U256)

	in = (&U256{}).New(nil, in.delta, true, nil, nil).(*U256)
	buffer = in.Encode()
	out = (&(U256{})).Decode(buffer).(*U256)
	if !out.value.Eq(UINT256ZERO) ||

		!out.delta.Eq(in.delta) ||
		!out.min.Eq(U256MIN) ||
		!out.max.Eq(U256MAX) {
		t.Error("Error: Out of range, should have failed")
	}
}
