package ccurltest

import (
	"fmt"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/codec"
	commutative "github.com/arcology-network/concurrenturl/v2/commutative"
)

func TestNewCodec(t *testing.T) {
	t0 := time.Now()
	// reflect.TypeOf(&commutative.Uint64{}).FieldByName("value")
	fmt.Println("reflect.TypeOf(Uint64{}).FieldByName(): ", time.Since(t0))

	t0 = time.Now()
	(&commutative.Uint64{}).Encode()
	fmt.Println("Uint64.Encode() ", time.Since(t0))

	t0 = time.Now()
	buffer := (&commutative.Uint64{}).Encode()
	fmt.Println("(commutative.Uint64{}).Encode():", time.Since(t0))

	t0 = time.Now()
	(&commutative.Uint64{}).Decode(buffer)
	fmt.Println("(commutative.Uint64{}).Decode():", time.Since(t0))

	v := &commutative.Uint64{}
	t0 = time.Now()
	codec.Encodeables([]codec.Encodeable{v, v}).Encode()
	fmt.Println("codec.Encodeables:", time.Since(t0))
}
