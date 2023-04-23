package ccurltest

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl/v2"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/type/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/v2/univalue"
)

func TestNoncommutativeCodec(t *testing.T) {
	/* Noncommutative Path Test*/
	alice := datacompression.RandomAccount()
	inPath, err := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	if err != nil {
		t.Error(err)
	}

	pathBytes := inPath.(*commutative.Meta).Encode()
	outPath := (&commutative.Meta{}).Decode(pathBytes)
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

func TestUnivalueCodec(t *testing.T) {
	store := cachedstorage.NewDataStore()
	transitions := []ccurlcommon.UnivalueInterface{}

	url := ccurl.NewConcurrentUrl(store)
	url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), fmt.Sprint("rand.Int()"))
	_, transVec := url.Export(nil)
	transitions = append(transitions, transVec...)

	for i := 0; i < len(transitions); i++ {
		buffer := transitions[i].Encode()
		in := transitions[i]
		out := (&univalue.Univalue{}).Decode(buffer).(ccurlcommon.UnivalueInterface)
		out.(*univalue.Univalue).ClearReserve()
		if !reflect.DeepEqual(in, out) {
			fmt.Println("Error: Missmatched")
		}
	}
}

func TestUnivaluesCodec(t *testing.T) {
	store := cachedstorage.NewDataStore()
	transitions := []ccurlcommon.UnivalueInterface{}
	for i := 0; i < 110000; i++ {
		acct := datacompression.RandomAccount()
		url := ccurl.NewConcurrentUrl(store)
		url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), acct)
		_, transVec := url.Export(nil)
		transitions = append(transitions, transVec...)
	}
	t0 := time.Now()
	buffer := univalue.Univalues(transitions).Encode()
	fmt.Println("Encode() ", len(transitions), " univalue in :", time.Since(t0))

	t0 = time.Now()
	out := (univalue.Univalues([]ccurlcommon.UnivalueInterface{})).Decode(buffer).(univalue.Univalues)
	fmt.Println("Decode() ", len(transitions), " univalue in :", time.Since(t0))

	for i := 0; i < len(transitions); i++ {
		tran := transitions[i]
		univ := out[i].(*univalue.Univalue)
		univ.ClearReserve()
		if !reflect.DeepEqual(tran, univ) {
			//fmt.Println("Error: Missmatched")
		}
	}
}
