package univalue

import (
	"testing"

	"github.com/arcology-network/common-lib/datacompression"
	commutative "github.com/arcology-network/concurrenturl/v2/commutative"
)

func TestUnimetaCodecUint64(t *testing.T) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()

	// meta:= commutative.NewPath()
	u256 := commutative.NewUint64(0, 100)
	in := NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	bytes := in.Encode()
	out := (&Unimeta{}).Decode(bytes).(*Unimeta)

	if in == out {
		t.Error("Error")
	}
}
