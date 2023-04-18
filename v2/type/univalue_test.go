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
	univalue := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, balance)
	bytes := univalue.Encode()
	out := (&Univalue{}).Decode(bytes).(*Univalue).Value()

	if univalue.Value().(*commutative.U256).Value().(*uint256.Int).Cmp(out.(*commutative.U256).Value().(*uint256.Int)) != 0 ||
		univalue.Value().(*commutative.U256).GetDelta().(*uint256.Int).Cmp(out.(*commutative.U256).GetDelta().(*uint256.Int)) != 0 {
		t.Error("Error")
	}

	// fmt.Println("Encoded length of one entry:", len(bytes)*4)

	// in := make([]ccurlcommon.UnivalueInterface, 10)
	// for i := 0; i < len(in); i++ {
	// 	in[i] = NewUnivalue( "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, v)
	// }

	// t0 := time.Now()
	// bytes = Univalues(in).Encode()
	// fmt.Println("Encoded 100000 entires in :", time.Since(t0), "Total size: ", len(bytes)*4)

	// t0 = time.Now()
	// out = (Univalues([]ccurlcommon.UnivalueInterface{})).Decode(bytes).(Univalues)

	// for i := 0; i < len(in); i++ {
	// 	if in[i].GetTx() != out[i].GetTx() ||
	// 		*in[i].GetPath() != *out[i].GetPath() ||
	// 		in[i].Reads() != out[i].Reads() ||
	// 		in[i].Writes() != out[i].Writes() ||
	// 		!reflect.DeepEqual(in[i].Value(), out[i].Value()) ||
	// 		in[i].Preexist() != out[i].Preexist() ||
	// 		in[i].ConcurrentWritable() != out[i].ConcurrentWritable() {
	// 		fmt.Println(in[i])
	// 		fmt.Println(out[i])
	// 		t.Error("Error")
	// 	}
	// }
	// fmt.Println("Decoded 100000 entires in :", time.Since(t0))
}

func BenchmarkUnivalueEncodeDecode(t *testing.B) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()
	v, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	univalue := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, v)
	bytes := univalue.Encode()
	fmt.Println("Encoded length of one entry:", len(bytes)*4)

	in := make([]ccurlcommon.UnivalueInterface, 1000000)
	for i := 0; i < len(in); i++ {
		in[i] = NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, v)
	}

	t0 := time.Now()
	bytes = Univalues(in).Encode()
	fmt.Println("Encoded", len(in), "entires in :", time.Since(t0), "Total size: ", len(bytes)*4)

	t0 = time.Now()
	(Univalues([]ccurlcommon.UnivalueInterface{})).Decode(bytes)
	fmt.Println("Decoded 100000 entires in :", time.Since(t0))
}
