package commutative

import (
	"testing"

	// ccurltype "github.com/arcology-network/concurrenturl/v2/type"

	"github.com/holiman/uint256"
)

func TestMaxUint256(t *testing.T) {
	max := uint256.NewInt(0).SetAllOne()
	t.Log(max)
}

func TestNewU256(t *testing.T) {
	if NewU256(uint256.NewInt(5), uint256.NewInt(4), uint256.NewInt(6)) == nil { // Between
		t.Error("Error: Failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(5), uint256.NewInt(6)) == nil { // On lower
		t.Error("Error: Failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(5), uint256.NewInt(5)) == nil { // on both
		t.Error("Error: Failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(4), uint256.NewInt(5)) == nil { // on upper
		t.Error("Error: Failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(4), uint256.NewInt(4)) != nil { // out of the both
		t.Error("Error: Should have failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(1), uint256.NewInt(4)) != nil { // out of the upper
		t.Error("Error: Should have failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(5), uint256.NewInt(4)) != nil { // lower is greater than the upper
		t.Error("Error: Should have failed")
	}
}

func TestU256(t *testing.T) {
	v := NewU256(uint256.NewInt(5), uint256.NewInt(4), uint256.NewInt(6))
	delta := NewU256Delta(uint256.NewInt(0), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(1), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(0), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(1), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	delta = NewU256Delta(uint256.NewInt(1), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error(err)
	}

	if _, _, _, _, err := v.(*U256).Set(delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	finalized, _, _ := v.(*U256).Get()
	if finalized.(*uint256.Int).ToBig().Uint64() != 6 {
		t.Error("Error: Should have failed")
	}
}

func TestU256DeltaOutRange(t *testing.T) {
	v := NewU256(uint256.NewInt(50), uint256.NewInt(40), uint256.NewInt(60))
	delta := NewU256Delta(uint256.NewInt(0), true)
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(10), false) //  - 10
	if _, _, _, _, err := v.(*U256).Set(delta, nil); err != nil {
		t.Error(err)
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
	in := NewU256(uint256.NewInt(4), uint256.NewInt(0), uint256.NewInt(14)).(*U256)
	buffer := in.Encode()
	out := (&(U256{})).Decode(buffer).(*U256)

	if out.value.Uint64() != 4 ||
		(*out.min).Uint64() != (*in.min).Uint64() ||
		(*out.max).Uint64() != (*in.max).Uint64() ||
		out.deltaPositive != in.deltaPositive {
		t.Error("Error: Out of range, should have failed")
	}

	buffer = in.Encode(true, true, true, true)
	out = (&(U256{})).Decode(buffer).(*U256)

	if out.value.Uint64() != 4 ||
		(*out.delta).Uint64() != (*in.delta).Uint64() ||
		out.deltaPositive != in.deltaPositive ||
		(*out.min).Uint64() != (*in.min).Uint64() ||
		(*out.max).Uint64() != (*in.max).Uint64() {
		t.Error("Error: Out of range, should have failed")
	}

	buffer = in.Encode(false, true, false, false)
	out = (&(U256{})).Decode(buffer).(*U256)
	if !out.delta.Eq(in.delta) || !out.value.Eq(U256MIN) || !out.min.Eq(U256MIN) || !out.max.Eq(U256MAX) {
		t.Error("Error: Out of range, should have failed")
	}

	buffer = in.Encode(true, false, true, true)
	out = (&(U256{})).Decode(buffer).(*U256)
	if !out.delta.Eq(U256MIN) || !out.value.Eq(in.value) || !out.min.Eq(in.min) || !out.max.Eq(in.max) {
		t.Error("Error: Out of range, should have failed")
	}

	buffer = in.Encode(false, false, false, false)
	out = (&(U256{})).Decode(buffer).(*U256)
	if !out.delta.Eq(U256MIN) || !out.value.Eq(U256MIN) || !out.min.Eq(U256MIN) || !out.max.Eq(U256MAX) {
		t.Error("Error: Out of range, should have failed")
	}
}
