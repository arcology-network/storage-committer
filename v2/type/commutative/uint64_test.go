package commutative

import (
	"testing"
)

func TestNewUint64(t *testing.T) {
	v := NewUint64((5), (0)).(*Uint64)

	final, _, _ := v.Get("", nil)
	if final.(*Uint64).value != 5 || final.(*Uint64).finalized == true {
		t.Error("Wrong value")
	}

	v.Set(NewUint64((0), 1), nil)
	v.Set(NewUint64((0), 1), nil)
	v.Set(NewUint64((0), 1), nil)

	if v.value == 8 {
		t.Error("Wrong value")
	}

	final, _, _ = v.Get("", nil)
	if final.(*Uint64).value != 8 || final.(*Uint64).delta != 0 {
		t.Error("Wrong value")
	}

}

func TestUint64Codec(t *testing.T) {
	in := NewUint64(5, 0)
	buffer := in.(*Uint64).Encode()
	out := (&Uint64{}).Decode(buffer)
	if *(in.(*Uint64)) != *(out.(*Uint64)) {
		t.Error("Wrong value")
	}

	in = NewUint64(0, 0)
	buffer = in.(*Uint64).Encode()
	out = (&Uint64{}).Decode(buffer)
	if *(in.(*Uint64)) != *(out.(*Uint64)) {
		t.Error("Wrong value")
	}
}

// func TestSetU64(t *testing.T) {
// 	v := NewUint64((5), (0), (4), (6), ADDITION)
// 	delta := NewU256Delta((0))
// 	if _, _, err := v.(*U256).Set( delta, nil); err != nil {
// 		t.Error(err)
// 	}

// 	delta = NewU256Delta((1))
// 	if _, _, err := v.(*U256).Set( delta, nil); err != nil {
// 		t.Error(err)
// 	}

// 	delta = NewU256Delta((0))
// 	if _, _, err := v.(*U256).Set( delta, nil); err != nil {
// 		t.Error(err)
// 	}

// 	delta = NewU256Delta((1))
// 	if _, _, err := v.(*U256).Set( delta, nil); err == nil {
// 		t.Error("Error: Should have failed")
// 	}

// 	// value, _, _ := v.(*U256).Get("", nil)
// 	// if value.(*U256).value.ToBig().Uint64() != 6 {
// 	// 	fmt.Println("Error: Value is wrong")
// 	// }
// }

// func TestCodec(t *testing.T) {
// 	b := NewUint64((5), (0), (4), (61), ADDITION)
// 	balance := b.(*U256)
// 	fmt.Println("Value :", balance)

// 	buffer := balance.Encode()
// 	out := (&(U256{})).Decode(buffer).(*U256)
// 	fmt.Println("U256 Encoded size :", out)

// 	if out.value.Uint64() != 5 || (*out.min).Uint64() != 4 || (*out.max).Uint64() != (*balance.max).Uint64() || out.operation != balance.operation {
// 		t.Error("Error: Out of range, should have failed")
// 	}
// }

// func TestDeepCopy(t *testing.T) {
// 	b := NewUint64((5), big.NewInt(0))
// 	b.(*U256).Reset("", NewU256WithLimits((0), big.NewInt(0), (0), (100)), nil)
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
// 	balance := NewUint64((5), big.NewInt(0)).(*U256)

// 	if _, _, err := balance.Set( NewUint64(nil, big.NewInt(-2)), nil); err != nil {
// 		t.Error(err)
// 	}

// 	if _, _, err := balance.Set( NewUint64(nil, big.NewInt(-1)), nil); err != nil {
// 		t.Error(err)
// 	}

// 	if _, _, err := balance.Set( NewUint64(nil, big.NewInt(3)), nil); err != nil {
// 		t.Error(err)
// 	}

// 	v, _, _ := balance.Get("", nil)

// 	u256 := v.(*U256).Value().(*uint256.Int)
// 	fmt.Println("Value :", u256.Uint64())

// 	if u256.Uint64() != 5 {
// 		t.Error("Wrong value")
// 	}
//}
