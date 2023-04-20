package commutative

import (
	"math"
	"testing"
)

func TestNewInt64(t *testing.T) {
	v := NewInt64((5), (0)).(*Int64)

	final, _, _ := v.Get(nil)
	if final.(*Int64).value != 5 || final.(*Int64).finalized {
		t.Error("Wrong value")
	}

	v.Set(NewInt64((0), -1), nil)
	v.Set(NewInt64((0), 1), nil)
	v.Set(NewInt64((0), 1), nil)

	if v.value == 8 {
		t.Error("Wrong value")
	}

	final, _, _ = v.Get(nil)
	if final.(*Int64).value != 6 || final.(*Int64).delta != 0 || !final.(*Int64).finalized {
		t.Error("Wrong value")
	}
}

func TestInt64Codec(t *testing.T) {
	in := NewInt64(5, 0)
	buffer := in.(*Int64).Encode()
	out := (&Int64{}).Decode(buffer)
	if *(in.(*Int64)) != *(out.(*Int64)) {
		t.Error("Wrong value")
	}

	in = NewInt64(math.MinInt64, 0)
	buffer = in.(*Int64).Encode()
	out = (&Int64{}).Decode(buffer)
	if *(in.(*Int64)) != *(out.(*Int64)) {
		t.Error("Wrong value")
	}
}
