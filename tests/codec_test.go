package ccurltest

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

func TestNoncommutativeCodec(t *testing.T) {
	/* Noncommutative Path Test*/
	inPath := commutative.NewPath()

	pathBytes := inPath.(*commutative.Path).Encode()
	outPath := (&commutative.Path{}).Decode(pathBytes)
	inPath.(*commutative.Path).Equal(outPath.(*commutative.Path))

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
	transitions := []interfaces.Univalue{}

	url := ccurl.NewConcurrentUrl(store)
	url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), fmt.Sprint("rand.Int()"))
	transVec := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	transitions = append(transitions, transVec...)

	for i := 0; i < len(transitions); i++ {
		buffer := transitions[i].Encode()
		in := transitions[i]
		out := (&univalue.Univalue{}).Decode(buffer).(interfaces.Univalue)
		// out.(*univalue.Univalue).ClearReserve()

		if !in.Equal(out) {
			t.Error("Error: Missmatched")
		}
	}
}

func TestUnivaluesCodec(t *testing.T) {
	store := cachedstorage.NewDataStore()
	transitions := []interfaces.Univalue{}
	for i := 0; i < 110000; i++ {
		acct := datacompression.RandomAccount()
		url := ccurl.NewConcurrentUrl(store)
		url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), acct)
		transVec := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
		transitions = append(transitions, transVec...)
	}
	t0 := time.Now()
	buffer := indexer.Univalues(transitions).Encode()
	fmt.Println("Encode() ", len(transitions), " univalue in :", time.Since(t0))

	t0 = time.Now()
	out := (indexer.Univalues([]interfaces.Univalue{})).Decode(buffer).(indexer.Univalues)
	fmt.Println("Decode() ", len(transitions), " univalue in :", time.Since(t0))

	for i := 0; i < len(transitions); i++ {
		tran := transitions[i]
		univ := out[i].(*univalue.Univalue)
		univ.ClearCache()
		if !reflect.DeepEqual(tran, univ) {
			//fmt.Println("Error: Missmatched")
		}
	}
}
