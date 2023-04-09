package commutative

import (
	"fmt"
	"testing"

	// ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	"github.com/holiman/uint256"
)

func TestMaxUint256(t *testing.T) {
	max := uint256.NewInt(0).SetAllOne()
	t.Log(max)
}

func TestNewU256(t *testing.T) {
	if NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(4), uint256.NewInt(6), ADDITION) == nil { // Between
		t.Error("Error: Failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(5), uint256.NewInt(6), ADDITION) == nil { // On lower
		t.Error("Error: Failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(5), uint256.NewInt(5), ADDITION) == nil { // on both
		t.Error("Error: Failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(4), uint256.NewInt(5), ADDITION) == nil { // on upper
		t.Error("Error: Failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(4), uint256.NewInt(4), ADDITION) != nil { // out of the both
		t.Error("Error: Should have failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(1), uint256.NewInt(4), ADDITION) != nil { // out of the upper
		t.Error("Error: Should have failed")
	}

	if NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(5), uint256.NewInt(4), ADDITION) != nil { // lower is greater than the upper
		t.Error("Error: Should have failed")
	}
}

func TestSetU256(t *testing.T) {
	v := NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(4), uint256.NewInt(6), ADDITION)
	delta := NewU256Delta(uint256.NewInt(0))
	if _, _, err := v.(*U256).Set("", delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(1))
	if _, _, err := v.(*U256).Set("", delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(0))
	if _, _, err := v.(*U256).Set("", delta, nil); err != nil {
		t.Error(err)
	}

	delta = NewU256Delta(uint256.NewInt(1))
	if _, _, err := v.(*U256).Set("", delta, nil); err == nil {
		t.Error("Error: Should have failed")
	}

	// value, _, _ := v.(*U256).Get("", nil)
	// if value.(*U256).value.ToBig().Uint64() != 6 {
	// 	fmt.Println("Error: Value is wrong")
	// }
}

func TestCodec(t *testing.T) {
	b := NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(4), uint256.NewInt(61), ADDITION)
	balance := b.(*U256)
	fmt.Println("Value :", balance)

	buffer := balance.Encode()
	out := (&(U256{})).Decode(buffer).(*U256)
	fmt.Println("U256 Encoded size :", out)

	if out.value.Uint64() != 5 || (*out.min).Uint64() != 4 || (*out.max).Uint64() != (*balance.max).Uint64() || out.operation != balance.operation {
		t.Error("Error: Out of range, should have failed")
	}
}

// func TestDeepCopy(t *testing.T) {
// 	b := NewU256(uint256.NewInt(5), big.NewInt(0))
// 	b.(*U256).Reset("", NewU256WithLimits(uint256.NewInt(0), big.NewInt(0), uint256.NewInt(0), uint256.NewInt(100)), nil)
// 	b0 := b.(*U256).Deepcopy()

// 	if b.(*U256).value.Uint64() != uint64(0) {
// 		t.Error("Error: Wrong value")
// 	}

// 	if !bytes.Equal(b0.(*U256).value.Bytes(), b.(*U256).value.Bytes()) ||
// 		!bytes.Equal(b0.(*U256).min.Bytes(), b.(*U256).min.Bytes()) ||
// 		!bytes.Equal(b0.(*U256).max.Bytes(), b.(*U256).max.Bytes()) ||
// 		b0.(*U256).delta.Cmp(b.(*U256).delta) != 0 {
// 		t.Error("Error: Wrong value")
// 	}
// }

// func TestGet(t *testing.T) {
// 	balance := NewU256(uint256.NewInt(5), big.NewInt(0)).(*U256)

// 	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-2)), nil); err != nil {
// 		t.Error(err)
// 	}

// 	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-1)), nil); err != nil {
// 		t.Error(err)
// 	}

// 	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(3)), nil); err != nil {
// 		t.Error(err)
// 	}

// 	v, _, _ := balance.Get("", nil)

// 	u256 := v.(*U256).Value().(*uint256.Int)
// 	fmt.Println("Value :", u256.Uint64())

// 	if u256.Uint64() != 5 {
// 		t.Error("Wrong value")
// 	}
//}
