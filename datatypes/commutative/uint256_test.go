package commutative

import (
	"testing"

	"github.com/arcology-network/common-lib/codec"
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
	in := NewU256((*uint256.Int)(U256_MIN), (*uint256.Int)(U256_MIN)).(*U256)

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

	in = NewU256((*uint256.Int)(U256_MIN), (*uint256.Int)(U256_MAX)).(*U256)

	buffer = (&U256{}).New(nil, in.delta, true, nil, nil).(*U256).Encode()
	out = (&(U256{})).Decode(buffer).(*U256)
	if !out.value.Eq((*codec.Uint256)(U256_ZERO)) ||
		!out.delta.Eq(in.delta) ||
		!out.min.Eq((*codec.Uint256)(U256_MIN)) ||
		!out.max.Eq((*codec.Uint256)(U256_MAX)) {
		t.Error("Error: Out of range, should have failed")
	}
}

// func TestCodecRlp(t *testing.T) {
// 	in := NewU256(U256_MIN, U256_MIN).(*U256)

// 	in.value = (&codec.Uint256{}).NewInt(111)

// 	buffer, _ := rlp.EncodeToBytes(in.max)
// 	out := (&(U256{})).Decode(buffer).(*U256)
// 	if out.value.Uint64() != 0 ||
// 		(*out.min).Uint64() != (*in.min).Uint64() ||
// 		(*out.max).Uint64() != (*in.max).Uint64() ||
// 		out.deltaPositive != in.deltaPositive {
// 		t.Error("Error: Mismatch after Encode()/Decode()")
// 	}

// 	buffer = in.EncodeRlp()
// 	out = (&(U256{})).DecodeRlp(buffer).(*U256)
// 	if (*out.delta).Uint64() != (*in.delta).Uint64() ||
// 		out.deltaPositive != in.deltaPositive ||
// 		(*out.min).Uint64() != (*in.min).Uint64() ||
// 		(*out.max).Uint64() != (*in.max).Uint64() {
// 		t.Error("Error: Out of range, should have failed")
// 	}

// 	in = NewU256(U256_MIN, U256_MAX).(*U256)

// 	buffer = (&U256{}).New(nil, in.delta, true, nil, nil).(*U256).EncodeRlp()
// 	out = (&(U256{})).DecodeRlp(buffer).(*U256)
// 	if !out.value.Eq((*codec.Uint256)(U256_ZERO)) ||
// 		!out.delta.Eq(in.delta) ||
// 		!out.min.Eq((*codec.Uint256)(U256_MIN)) ||
// 		!out.max.Eq((*codec.Uint256)(U256_MAX)) {
// 		t.Error("Error: Out of range, should have failed")
// 	}
// }
