package commutative

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	// ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	"github.com/holiman/uint256"
)

func TestUint256Basic(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(3), nil); err != nil {
		t.Error(err)
	}

	v, _, _ := balance.Get(0, "", nil)

	u256 := v.(*Balance).Value().(uint256.Int)
	fmt.Println("Value :", u256.Uint64())

	if u256.Uint64() != 5 {
		t.Error("Wrong value")
	}
}

func TestUint256LowerLimitOnly(t *testing.T) {
	b, err := NewBalance(uint256.NewInt(5), uint256.NewInt(0), nil)
	if err != nil {
		t.Error(err)
	}

	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(2), nil); err != nil {
		t.Error(err)
	}

	_, err = NewBalance(uint256.NewInt(5), uint256.NewInt(6), uint256.NewInt(100))
	if err == nil {
		t.Error("Should report an error !")
	}

	if v, _, _ := balance.Get(0, "", nil); v.(*Balance).value.Uint64() != 4 {
		t.Error("Should report an error !")
	}
}

func TestUint256UpperLimitOnly(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(5), nil, uint256.NewInt(100))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		t.Error("Should report an error !")
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(30), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(300), nil); err == nil {
		t.Error("Error: Out of range, should have failed")
	}
}

func TestUint256NoLimit(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(5), nil, nil)
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		t.Error("Should report an error !")
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(30), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(300), nil); err != nil {
		t.Error(err)
	}
}

func TestUint256LowerAndUpperLimit(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		t.Error("Should report an error !")
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(30), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(300), nil); err == nil {
		t.Error("Error: Out of range, should have failed")
	}
}

func TestUint256LowerGreaterThanUpper(t *testing.T) {
	b, err := NewBalance(uint256.NewInt(5), uint256.NewInt(100), uint256.NewInt(0))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)
	if err == nil {
		t.Error("Error: Out of range, should have failed")
	}

	b, err = NewBalance(uint256.NewInt(99), uint256.NewInt(100), uint256.NewInt(0))
	balance = b.(*Balance)
	fmt.Println("Value :", balance)
	if err == nil {
		t.Error("Error: Out of range, should have failed")
	}
}

func TestUnderflow(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(0), nil, uint256.NewInt(100))
	_, _, err := b.(*Balance).Set(0, "", big.NewInt(-1), nil)
	if err == nil {
		t.Error("Error: Should have reported out of range error!!")
	}
}

func TestOverflow(t *testing.T) {
	initv := uint256.NewInt(0)
	for i := 0; i < 4; i++ {
		initv[i] = math.MaxUint64
	}

	b, err := NewBalance(initv, uint256.NewInt(0), uint256.NewInt(100))
	if err == nil {
		t.Error("Error: Should have reported out-of-range error")
	}

	fmt.Println("Value :", b)

	b, err = NewBalance(initv, nil, nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Value :", b)

	_, _, err = b.(*Balance).Set(0, "", big.NewInt(0), nil)
	if err != nil {
		t.Error(err)
	}

	_, _, err = b.(*Balance).Set(0, "", big.NewInt(1), nil)
	if err == nil {
		t.Error("Error: Should have reported out of range error!!")
	}
}

func TestCodec(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(3), nil); err != nil {
		t.Error(err)
	}
	v, _, _ := balance.Get(0, "", nil)

	u256 := v.(*Balance).Value().(uint256.Int)
	if u256.Uint64() != 5 {
		t.Error("Error: Wrong value!!!")
	}

	buffer := balance.Encode()
	out := (&(Balance{})).Decode(buffer).(*Balance)
	fmt.Println("Balance Encoded size :", out)

	if out.value.Uint64() != 5 || (*out.min).Uint64() != 0 || (*out.max).Uint64() != 100 {
		t.Error("Error: Out of range, should have failed")
	}

	buffer = balance.EncodeCompact()
	fmt.Println("Balance Encoded size :", buffer)

	newV := (&Balance{}).DecodeCompact(buffer).(*Balance)
	if newV.value.Uint64() != 5 || newV.delta.Uint64() != 0 || (*newV.min).Uint64() != 0 || (*newV.max).Uint64() != 100 {
		t.Error("Error: Out of range, should have failed")
	}
}

func TestDeepCopy(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))
	b0 := b.(*Balance).Deepcopy()

	if b.(*Balance).value != b0.(*Balance).value {
		t.Error("Error: Out of range, should have failed")
	}

	b0.(*Balance).value = *uint256.NewInt(7)
	b0.(*Balance).delta = big.NewInt(11)

	if b.(*Balance).value.Uint64() != 5 || b.(*Balance).delta.Cmp(big.NewInt(0)) != 0 || b0.(*Balance).delta.Cmp(big.NewInt(11)) != 0 {
		t.Error("Error: Wrong value")
	}

	if b0.(*Balance).value.Uint64() != 7 {
		t.Error("Error: Wrong value")
	}
}

func TestApplyDelta(t *testing.T) {
	// balanceDeltas := make([]ccurlcommon.UnivalueInterface, 3)

	// if v, err := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100)); err == nil {
	// 	balanceDeltas[0] = NewUnivalue(0, uint32(0), "", 0, 0, v.(*Balance))

	// 	if err = balanceDeltas[0].Set(0, "", big.NewInt(-2), nil); err != nil {
	// 		t.Error(err)
	// 	}
	// } else {
	// 	t.Error(err)
	// }

	// if v, err := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100)); err == nil {
	// 	balanceDeltas[1] = v.(*Balance)
	// 	if _, _, err = balanceDeltas[1].Set(0, "", big.NewInt(-2), nil); err != nil {
	// 		t.Error(err)
	// 	}
	// } else {
	// 	t.Error(err)
	// }

	// if v, err := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100)); err == nil {
	// 	balanceDeltas[2] = v.(*Balance)
	// 	if _, _, err = balanceDeltas[2].Set(0, "", big.NewInt(-1), nil); err != nil {
	// 		t.Error(err)
	// 	}
	// } else {
	// 	t.Error(err)
	// }

	// base, _ := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))

	// base.(*Balance).ApplyDelta(0, balanceDeltas)
}
