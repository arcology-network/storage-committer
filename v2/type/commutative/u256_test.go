package commutative

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	// ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	"github.com/holiman/uint256"
)

func TestMaxUint256(t *testing.T) {
	max := uint256.NewInt(0).SetAllOne()
	t.Log(max)
}

func TestUint256Basic(t *testing.T) {
	b := NewU256(uint256.NewInt(5), big.NewInt(0))
	balance := b.(*U256)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-2)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-1)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(3)), nil); err != nil {
		t.Error(err)
	}

	v, _, _ := balance.Get("", nil)

	u256 := v.(*U256).Value().(*uint256.Int)
	fmt.Println("Value :", u256.Uint64())

	if u256.Uint64() != 5 {
		t.Error("Wrong value")
	}
}

func TestUint256NoLimit(t *testing.T) {
	b := NewU256(uint256.NewInt(5), big.NewInt(0))
	balance := b.(*U256)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-2)), nil); err != nil {
		t.Error("Should report an error !")
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-1)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(30)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(300)), nil); err != nil {
		t.Error(err)
	}
}

func TestUint256LowerAndUpperLimit(t *testing.T) {
	b := NewU256(uint256.NewInt(5), big.NewInt(0))
	balance := b.(*U256)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Reset("", NewU256WithLimit(uint256.NewInt(0), uint256.NewInt(100)), nil); err != nil {
		t.Error("Failed to reset limit.")
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-2)), nil); err != nil {
		t.Error("Should report an error !")
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-1)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(30)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(300)), nil); err == nil {
		t.Error("Error: Out of range, should have failed")
	}
}

func TestUnderflow(t *testing.T) {
	b := NewU256(uint256.NewInt(0), big.NewInt(0))
	_, _, err := b.(*U256).Set("", NewU256(nil, big.NewInt(-1)), nil)
	if err == nil {
		t.Error("Error: Should have reported out of range error!!")
	}
}

func TestOverflow(t *testing.T) {
	initv := uint256.NewInt(0)
	for i := 0; i < 4; i++ {
		initv[i] = math.MaxUint64
	}

	b := NewU256(initv, big.NewInt(0))
	fmt.Println("Value :", b)

	_, _, err := b.(*U256).Set("", NewU256(nil, big.NewInt(0)), nil)
	if err != nil {
		t.Error(err)
	}

	_, _, err = b.(*U256).Set("", NewU256(nil, big.NewInt(1)), nil)
	if err == nil {
		t.Error("Error: Should have reported out of range error!!")
	}
}

func TestCodec(t *testing.T) {
	b := NewU256(uint256.NewInt(5), big.NewInt(0))
	balance := b.(*U256)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Reset("", NewU256WithLimit(uint256.NewInt(0), uint256.NewInt(100)), nil); err != nil {
		t.Error("Failed to reset limit.")
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-2)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(-1)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewU256(nil, big.NewInt(3)), nil); err != nil {
		t.Error(err)
	}
	v, _, _ := balance.Get("", nil)

	u256 := v.(*U256).Value().(*uint256.Int)
	if u256.Uint64() != 5 {
		t.Error("Error: Wrong value!!!")
	}

	buffer := balance.Encode()
	out := (&(U256{})).Decode(buffer).(*U256)
	fmt.Println("U256 Encoded size :", out)

	if out.value.Uint64() != 5 || (*out.min).Uint64() != 0 || (*out.max).Uint64() != 100 {
		t.Error("Error: Out of range, should have failed")
	}

	buffer = balance.EncodeCompact()
	fmt.Println("U256 Encoded size :", buffer)

	newV := (&U256{}).DecodeCompact(buffer).(*U256)
	if newV.value.Uint64() != 5 || newV.delta.Uint64() != 0 || (*newV.min).Uint64() != 0 || (*newV.max).Uint64() != 100 {
		t.Error("Error: Out of range, should have failed")
	}
}

func TestDeepCopy(t *testing.T) {
	b := NewU256(uint256.NewInt(5), big.NewInt(0))
	b.(*U256).Reset("", NewU256WithLimit(uint256.NewInt(0), uint256.NewInt(100)), nil)
	b0 := b.(*U256).Deepcopy()

	b0.(*U256).value = uint256.NewInt(7)
	b0.(*U256).delta = big.NewInt(11)

	if b.(*U256).value.Uint64() != 5 || b.(*U256).delta.Cmp(big.NewInt(0)) != 0 || b0.(*U256).delta.Cmp(big.NewInt(11)) != 0 {
		t.Error("Error: Wrong value")
	}

	if b0.(*U256).value.Uint64() != 7 {
		t.Error("Error: Wrong value")
	}
}

func TestApplyDelta(t *testing.T) {
	// balanceDeltas := make([]ccurlcommon.UnivalueInterface, 3)

	// if v, err := NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100)); err == nil {
	// 	balanceDeltas[0] = NewUnivalue(uint32(0), "", v.(*U256))

	// 	if err = balanceDeltas[0].Set("", big.NewInt(-2), nil); err != nil {
	// 		t.Error(err)
	// 	}
	// } else {
	// 	t.Error(err)
	// }

	// if v, err := NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100)); err == nil {
	// 	balanceDeltas[1] = v.(*U256)
	// 	if _, _, err = balanceDeltas[1].Set("", big.NewInt(-2), nil); err != nil {
	// 		t.Error(err)
	// 	}
	// } else {
	// 	t.Error(err)
	// }

	// if v, err := NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100)); err == nil {
	// 	balanceDeltas[2] = v.(*U256)
	// 	if _, _, err = balanceDeltas[2].Set("", big.NewInt(-1), nil); err != nil {
	// 		t.Error(err)
	// 	}
	// } else {
	// 	t.Error(err)
	// }

	// base, _ := NewU256(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))

	// base.(*U256).ApplyDelta(balanceDeltas)
}