package commutative

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/holiman/uint256"
)

func TestUint256Basic(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		fmt.Println(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		fmt.Println(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(3), nil); err != nil {
		fmt.Println(err)
	}

	v, _, _ := balance.Get(0, "", nil)

	u256 := v.(*Balance).Value().(uint256.Int)
	fmt.Println("Value :", u256.Uint64())

	if u256.Uint64() != 5 {
		t.Error("Wrong value")
	}
}

func TestUint256LowerLimit(t *testing.T) {
	b, err := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		fmt.Println(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		fmt.Println(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(3), nil); err != nil {
		fmt.Println(err)
	}

	b, err = NewBalance(uint256.NewInt(5), uint256.NewInt(6), uint256.NewInt(100))
	balance = b.(*Balance)

	if err == nil {
		t.Error("Should report an error !")
	}
}

func TestUint256UpperLimit(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		fmt.Println(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		fmt.Println(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(30), nil); err != nil {
		fmt.Println(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(300), nil); err == nil {
		fmt.Println("Error: Out of range, should have failed")
	}
}

func TestCodecBalance(t *testing.T) {
	b, _ := NewBalance(uint256.NewInt(5), uint256.NewInt(0), uint256.NewInt(100))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set(0, "", big.NewInt(-2), nil); err != nil {
		fmt.Println(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(-1), nil); err != nil {
		fmt.Println(err)
	}

	if _, _, err := balance.Set(0, "", big.NewInt(3), nil); err != nil {
		fmt.Println(err)
	}
	v, _, _ := balance.Get(0, "", nil)

	u256 := v.(*Balance).Value().(uint256.Int)
	if u256.Uint64() != 5 {
		fmt.Println("Error: Wrong value!!!")
	}

	buffer := balance.Encode()
	out := (&(Balance{})).Decode(buffer).(*Balance)
	fmt.Println("Balance Encoded size :", out)

	if out.value.Uint64() != 5 || out.delta.Uint64() != 2 || out.delta.Sign() != 1 || (*out.min).Uint64() != 0 || (*out.max).Uint64() != 100 {
		fmt.Println("Error: Out of range, should have failed")
	}
}
