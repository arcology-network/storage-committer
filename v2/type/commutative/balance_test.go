package commutative

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/holiman/uint256"
)

func TestUint256Basic(t *testing.T) {
	balance := NewBalance(big.NewInt(5), big.NewInt(-2), uint256.NewInt(0), uint256.NewInt(100)).(*Balance)
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

	if u256.Uint64() != 3 {
		t.Error("Wrong value")
	}
}

func TestUint256LowerLimit(t *testing.T) {
	balance := NewBalance(big.NewInt(5), big.NewInt(-2), uint256.NewInt(0), uint256.NewInt(100)).(*Balance)
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

	if _, _, err := balance.Set(0, "", big.NewInt(-4), nil); err == nil {
		fmt.Println("Error: Out of range, should have failed")
	}
}

func TestUint256UpperLimit(t *testing.T) {
	balance := NewBalance(big.NewInt(5), big.NewInt(-2), uint256.NewInt(0), uint256.NewInt(100)).(*Balance)
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
	/* Noncommutative Path Test*/
	balance := NewBalance(big.NewInt(5), big.NewInt(-2), uint256.NewInt(0), uint256.NewInt(100)).(*Balance)
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
	// fmt.Println("Balance Encoded size :", len(balance.Encode()))

	// fmt.Println("Balance Encoded Compact size :", len(balance.EncodeCompact()))

	// buffer0 := balance.Encode()

	// buffer := make([]byte, balance.Size())
	// balance.EncodeToBuffer(buffer)
	// out := (&(Balance{})).Decode(buffer).(*Balance)

	// if balance.Value().(*big.Int).Cmp(out.Value().(*big.Int)) != 0 ||
	// 	balance.GetDelta().(*big.Int).Cmp(out.GetDelta().(*big.Int)) != 0 {
	// 	fmt.Println(buffer0)
	// 	fmt.Println(buffer)
	// 	t.Error("Error")
	// }
}
