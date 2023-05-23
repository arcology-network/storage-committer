package noncommutative

import (
	"math/big"
	"testing"
	// "github.com/HPISTechnologies/concurrenturl/type/noncommutative"
)

func TestNewBigint(t *testing.T) {
	v := NewBigint(100).(*Bigint)

	out, _, _ := v.Get()
	outV := (*Bigint)(out.(*big.Int))
	if !outV.Equal(NewBigint(100).(*Bigint)) {
		t.Error("Mismatch")
	}
}
