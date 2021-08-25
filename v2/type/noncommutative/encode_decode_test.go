package noncommutative

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
)

func TestBigInt(t *testing.T) {
	var value big.Int
	value.SetInt64(-252388)
	avalue := Bigint(value)
	fmt.Printf("avalue=%v\n", avalue)
	data, err := avalue.GobEncode()
	if err != nil {
		t.Error("encode err=", err)
		return
	}
	bvalue := Bigint{}
	err = bvalue.GobDecode(data)
	if err != nil {
		t.Error("decode err=", err)
		return
	}
	fmt.Printf("bvalue=%v\n", bvalue)
	if !reflect.DeepEqual(avalue, bvalue) {
		t.Error("Error: Bigint Encoding/decoding error, value don't match")
	}
}
func TestBytes(t *testing.T) {
	abytes := Bytes{false, []byte{1, 12, 2}}
	data, err := abytes.GobEncode()
	if err != nil {
		t.Error("encode err=", err)
		return
	}
	bvalue := Bytes{}
	err = bvalue.GobDecode(data)
	if err != nil {
		t.Error("decode err=", err)
		return
	}
	fmt.Printf("bvalue=%v\n", bvalue)
	if !reflect.DeepEqual(abytes, bvalue) {
		t.Error("Error: bytes Encoding/decoding error, value don't match")
	}
}
func TestInt64(t *testing.T) {
	avalue := Int64(int64(-100))
	data, err := avalue.GobEncode()
	if err != nil {
		t.Error("encode err=", err)
		return
	}
	bvalue := Int64(0)
	err = bvalue.GobDecode(data)
	if err != nil {
		t.Error("decode err=", err)
		return
	}
	fmt.Printf("bvalue=%v\n", bvalue)
	if !reflect.DeepEqual(avalue, bvalue) {
		t.Error("Error: int64 Encoding/decoding error, value don't match")
	}
}
func TestString(t *testing.T) {
	avalue := String("dfdfdfesdf")
	data, err := avalue.GobEncode()
	if err != nil {
		t.Error("encode err=", err)
		return
	}
	bvalue := String("")
	err = bvalue.GobDecode(data)
	if err != nil {
		t.Error("decode err=", err)
		return
	}
	fmt.Printf("bvalue=%v\n", bvalue)
	if !reflect.DeepEqual(avalue, bvalue) {
		t.Error("Error: int64 Encoding/decoding error, value don't match")
	}
}
