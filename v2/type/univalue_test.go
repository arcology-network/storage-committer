package ccurltype

import (
	"fmt"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/datacompression"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	"github.com/holiman/uint256"
)

func TestUnivalueEncodeDecode(t *testing.T) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()
	// v, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	balance := commutative.NewU256(uint256.NewInt(100), uint256.NewInt(0), uint256.NewInt(100))
	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, balance)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	bytes := in.Encode()
	v := (&Univalue{}).Decode(bytes).(*Univalue)
	out := v.Value()

	if in.Value().(*commutative.U256).Value().(*uint256.Int).Cmp(out.(*commutative.U256).Value().(*uint256.Int)) != 0 ||
		in.Value().(*commutative.U256).GetDelta().(*uint256.Int).Cmp(out.(*commutative.U256).GetDelta().(*uint256.Int)) != 0 {
		t.Error("Error")
	}

	if in.vType != v.vType ||
		in.tx != v.tx ||
		*in.path != *v.path ||
		in.writes != v.writes ||
		in.deltaWrites != v.deltaWrites ||
		in.preexists != v.preexists ||
		in.composite != v.composite {
		t.Error("Error: mismatch after decoding")
	}
}

func BenchmarkUnivalueEncodeDecode(t *testing.B) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()
	v, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	bytes := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 1, v).Encode()
	// bytes := in1.Encode()
	fmt.Println("Encoded length of one entry:", len(bytes)*4)

	in := make([]ccurlcommon.UnivalueInterface, 1000000)
	for i := 0; i < len(in); i++ {
		in[i] = NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 1, v)
	}

	t0 := time.Now()
	bytes = Univalues(in).Encode()
	fmt.Println("Encoded", len(in), "entires in :", time.Since(t0), "Total size: ", len(bytes)*4)

	t0 = time.Now()
	(Univalues([]ccurlcommon.UnivalueInterface{})).Decode(bytes)
	fmt.Println("Decoded 100000 entires in :", time.Since(t0))
}
