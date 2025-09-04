/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package noncommutative

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/ethereum/go-ethereum/rlp"
)

func TestNewBigint(t *testing.T) {
	v := NewBigint(100).(*Bigint)

	out, _, _ := v.Get()
	outV := (out.(big.Int))
	v2 := (out.(big.Int))
	if outV.Cmp(&v2) != 0 {
		t.Error("Mismatch")
	}
}

func TestBigintCodecs(t *testing.T) {
	v := NewBigint(100).(*Bigint)

	out, _, _ := v.Get()
	outV := out.(big.Int)
	v2 := (out.(big.Int))
	if outV.Cmp(&v2) != 0 {
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
	buffer := in.StorageEncode("")
	out := new(Int64).StorageDecode("", buffer)

	if *out.(*Int64) != 111 {
		t.Error("Mismatch expecting ", 100)
	}
}

func TestU256RlpCodec(t *testing.T) {
	v := NewBytes([]byte{1, 2, 3, 4})
	buffer := v.StorageEncode("")
	output := (&Bytes{}).StorageDecode("", buffer)

	if v.(*Bytes).placeholder != output.(*Bytes).placeholder || !reflect.DeepEqual(v.(*Bytes).value, output.(*Bytes).value) {
		fmt.Println("Error: Missmatched")
	}
}

func TestInt64RlpCodec(t *testing.T) {
	v := NewInt64(12345)
	buffer := v.StorageEncode("")
	output := new(Int64).StorageDecode("", buffer)

	if *v != *output.(*Int64) {
		fmt.Println("Error: Missmatched")
	}
}

func TestUint64Codec(t *testing.T) {
	v := NewUint64(56789)
	buffer := v.StorageEncode("")
	output := new(Uint64).StorageDecode("", buffer)

	if *v != *output.(*Uint64) {
		fmt.Println("Error: Missmatched")
	}
}

func TestStringRlpCodec(t *testing.T) {
	bytes := []byte{0, 0, 0, 1}
	encoded, _ := rlp.EncodeToBytes(bytes)
	fmt.Println(encoded)

	v := NewString("12345")
	buffer := v.StorageEncode("")
	output := new(String).StorageDecode("", buffer)

	if *(v.(*String)) != *(output.(*String)) {
		fmt.Println("Error: Missmatched")
	}
}

func TestByteRlp(t *testing.T) {
	v2 := slice.New[byte](32, 11)
	encoded, _ := rlp.EncodeToBytes(v2)

	buf := []byte{}
	err := rlp.DecodeBytes(encoded, &buf)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(v2, buf) {
		fmt.Println("Error: Missmatched")
	}
}

func TestBytesRlpCodec(t *testing.T) {
	v2 := slice.New[byte](33, 0)
	v2[32] = 1
	v := NewBytes(v2).(*Bytes)
	buffer := v.StorageEncode("")
	output := new(Bytes).StorageDecode("", buffer).(*Bytes)

	outv := output.Value().(codec.Bytes)
	if !bytes.Equal(v.Value().(codec.Bytes), outv) {
		fmt.Println("Error: Missmatched")
	}

	v2 = slice.New[byte](32, 11)
	v = NewBytes(v2).(*Bytes)
	buffer = v.StorageEncode("")
	output = new(Bytes).StorageDecode("", buffer).(*Bytes)

	outv = output.Value().(codec.Bytes)
	if !bytes.Equal(v.Value().(codec.Bytes), outv) {
		fmt.Println("Error: Missmatched")
	}

	v2 = slice.New[byte](25, 0)
	v2[24] = 1
	v = NewBytes(v2).(*Bytes)
	buffer = v.StorageEncode("")
	output = new(Bytes).StorageDecode("", buffer).(*Bytes)

	outv = output.Value().(codec.Bytes)
	if !bytes.Equal(v.Value().(codec.Bytes), outv) {
		t.Error("Error: Missmatched")
	}

	v2 = slice.New[byte](40, 0)
	v2[0] = 1
	v = NewBytes(v2).(*Bytes)
	buffer = v.StorageEncode("")
	output = new(Bytes).StorageDecode("", buffer).(*Bytes)

	outv = output.Value().(codec.Bytes)
	if !bytes.Equal(v.Value().(codec.Bytes), outv) {
		fmt.Println("Error: Missmatched")
	}

	v2 = slice.New[byte](40, 0)
	v2[39] = 1
	v = NewBytes(v2).(*Bytes)
	buffer = v.StorageEncode("")
	output = new(Bytes).StorageDecode("", buffer).(*Bytes)

	outv = output.Value().(codec.Bytes)
	if !bytes.Equal(v.Value().(codec.Bytes), outv) {
		fmt.Println("Error: Missmatched")
	}

	v2 = slice.New[byte](32, 11)
	encoded, _ := rlp.EncodeToBytes(v2)

	v3 := slice.New[byte](1, 0)
	rlp.DecodeBytes(encoded, &v3)

	if !bytes.Equal(v2, v3) {
		fmt.Println("Error: Missmatched")
	}
}
