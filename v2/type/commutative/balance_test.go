package commutative

import (
	"fmt"
	"math/big"
	"testing"
)

func TestCodecBalance(t *testing.T) {
	/* Noncommutative Path Test*/
	balance := NewBalance(big.NewInt(0), big.NewInt(-777)).(*Balance)
	fmt.Println("Balance Encoded size :", len(balance.Encode()))
	fmt.Println("Balance Encoded Compact size :", len(balance.EncodeCompact()))

	buffer0 := balance.Encode()

	buffer := make([]byte, balance.Size())
	balance.EncodeToBuffer(buffer)
	out := (&(Balance{})).Decode(buffer).(*Balance)

	if balance.Value().(*big.Int).Cmp(out.Value().(*big.Int)) != 0 ||
		balance.GetDelta().(*big.Int).Cmp(out.GetDelta().(*big.Int)) != 0 {
		fmt.Println(buffer0)
		fmt.Println(buffer)
		t.Error("Error")
	}
}
