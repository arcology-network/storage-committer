package noncommutative

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/arcology-network/evm/rlp"
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

func TestBigintCodecs(t *testing.T) {
	v := NewBigint(100).(*Bigint)

	out, _, _ := v.Get()
	outV := (*Bigint)(out.(*big.Int))
	if !outV.Equal(NewBigint(100).(*Bigint)) {
		t.Error("Mismatch")
	}

	encoded, err := rlp.EncodeToBytes(out)
	if err != nil {
		t.Error(err)
	}

	var decoded big.Int
	err = rlp.DecodeBytes(encoded, &decoded)
	if err != nil {
		t.Error(err)
	}

	if decoded.Uint64() != 100 {
		t.Error("Mismatch expecting ", 100)
	}
}

func TestBigintRlpCodecs(t *testing.T) {
	in := NewInt64(111)
	buffer := in.StorageEncode()
	out := new(Int64).StorageDecode(buffer)

	if *out.(*Int64) != 111 {
		t.Error("Mismatch expecting ", 100)
	}
}

func TestU256RlpCodec(t *testing.T) {
	v := NewBytes([]byte{1, 2, 3, 4})
	buffer := v.StorageEncode()
	output := (&Bytes{}).StorageDecode(buffer)

	if v.(*Bytes).placeholder != output.(*Bytes).placeholder || !reflect.DeepEqual(v.(*Bytes).value, output.(*Bytes).value) {
		fmt.Println("Error: Missmatched")
	}
}

func TestInt64RlpCodec(t *testing.T) {
	v := NewInt64(12345)
	buffer := v.StorageEncode()
	output := new(Int64).StorageDecode(buffer)

	if *v != *output.(*Int64) {
		fmt.Println("Error: Missmatched")
	}
}

func TestStringRlpCodec(t *testing.T) {
	v := NewString("12345")
	buffer := v.StorageEncode()
	output := new(String).StorageDecode(buffer)

	if *(v.(*String)) != *(output.(*String)) {
		fmt.Println("Error: Missmatched")
	}
}
