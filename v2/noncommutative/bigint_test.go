package noncommutative

import (
	"testing"
)

func TestNewBigint(t *testing.T) {
	v := NewBigint(100).(*Bigint)

	out, _, _ := v.Get()
	outV := out.(*Bigint)
	if !outV.Equal(NewBigint(100).(*Bigint)) {
		t.Error("Mismatch")
	}
}
