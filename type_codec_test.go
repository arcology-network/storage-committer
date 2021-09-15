package concurrenturl

import (
	"math/big"
	"reflect"
	"testing"

	commutative "github.com/arcology-network/concurrenturl/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/type/noncommutative"
)

func TestCodecNoncommutative(t *testing.T) {
	/* Noncommutative Path Test*/
	inPath, err := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	if err != nil {
		t.Error(err)
	}

	pathBytes := inPath.(*commutative.Meta).Encode()
	outPath := (&commutative.Meta{}).Decode(pathBytes).(*commutative.Meta)
	if !reflect.DeepEqual(inPath, outPath) {
		t.Error("Error: Path Encoding/decoding error, paths don't match")
	}

	/* Noncommutative String Test */
	inStr := noncommutative.NewString("ctrn-0")
	strBytes := inStr.(*noncommutative.String).Encode()
	stringer := noncommutative.String("")
	outStr := stringer.Decode(strBytes)
	if !reflect.DeepEqual(inStr, outStr) {
		t.Error("Error: String Encoding/decoding error, strings don't match")
	}

	/*  []byte Test */
	inBytes := noncommutative.NewBytes([]byte("test bytes"))
	encoded := inBytes.(*noncommutative.Bytes).Encode()

	result := (&noncommutative.Bytes{}).Decode(encoded)
	outBytes := noncommutative.Bytes(*(result.(*noncommutative.Bytes)))

	if !reflect.DeepEqual(inBytes, &outBytes) {
		t.Error("Error: Bytes Encoding/decoding error, bytes don't match")
	}

	/* Int64 Test */
	inInt64 := noncommutative.NewInt64(12345)
	int64Bytes := inInt64.(*noncommutative.Int64).Encode()
	int64cdc := noncommutative.Int64(0)
	outInt64 := (&int64cdc).Decode(int64Bytes).(*noncommutative.Int64)
	if !reflect.DeepEqual(inInt64, outInt64) {
		t.Error("Error: Int64 Encoding/decoding error, numbers don't match")
	}

	/* Bigint Test */
	inBig := noncommutative.NewBigint(789456)
	bigBytes := inBig.(*noncommutative.Bigint).Encode()
	outBig := (&noncommutative.Bigint{}).Decode(bigBytes).(*noncommutative.Bigint)
	if !reflect.DeepEqual(inBig, outBig) {
		t.Error("Error: Bigint Encoding/decoding error, numbers don't match")
	}
}

func TestCodecCommutative(t *testing.T) {
	/* Commutative Int64 Test */
	inInt64 := commutative.NewInt64(12345, 0)
	int64Bytes := inInt64.(*commutative.Int64).Encode()
	outInt64 := (&commutative.Int64{}).Decode(int64Bytes).(*commutative.Int64)
	if !reflect.DeepEqual(inInt64, outInt64) {
		t.Error("Error: Int64 Encoding/decoding error, numbers don't match")
	}

	/* Commutative Bigint Test */
	inBig := commutative.NewBalance(big.NewInt(789456), big.NewInt(0))
	bigBytes := inBig.(*commutative.Balance).Encode()
	outBig := (&commutative.Balance{}).Decode(bigBytes).(*commutative.Balance)
	if !reflect.DeepEqual(inBig, outBig) {
		t.Error("Error: Bigint Encoding/decoding error, numbers don't match")
	}

}
