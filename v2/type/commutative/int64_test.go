package commutative

import (
	"fmt"
	"math/big"
	"testing"

	uint256 "github.com/holiman/uint256"
)

func TestMaxint64(t *testing.T) {
	max := uint256.NewInt(0).SetAllOne()
	t.Log(max)
}

func TestInt64Basic(t *testing.T) {
	b := NewBalance(uint256.NewInt(5), big.NewInt(0))
	balance := b.(*Balance)
	fmt.Println("Value :", balance)

	if _, _, err := balance.Set("", NewBalance(nil, big.NewInt(-2)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewBalance(nil, big.NewInt(-1)), nil); err != nil {
		t.Error(err)
	}

	if _, _, err := balance.Set("", NewBalance(nil, big.NewInt(3)), nil); err != nil {
		t.Error(err)
	}

	v, _, _ := balance.Get("", nil)

	u256 := v.(*Balance).Value().(*uint256.Int)
	fmt.Println("Value :", u256.Uint64())

	if u256.Uint64() != 5 {
		t.Error("Wrong value")
	}
}
