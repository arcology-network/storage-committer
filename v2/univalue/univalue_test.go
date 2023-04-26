package univalue

import (
	"fmt"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/datacompression"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	"github.com/holiman/uint256"
)

func TestUnivalueCodecUint64(t *testing.T) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()

	// meta:= commutative.NewMeta()
	u256 := commutative.NewUint64(0, 100)
	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	bytes := in.Encode()
	v := (&Univalue{}).Decode(bytes).(*Univalue)
	out := v.Value()

	if *(in.value.(*commutative.Uint64)) != *(out.(*commutative.Uint64)) {
		t.Error("Error")
	}
}

func TestUnivalueCodecU256(t *testing.T) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()

	// meta:= commutative.NewMeta()
	u256 := commutative.NewU256(uint256.NewInt(100), uint256.NewInt(0), uint256.NewInt(100))
	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256)
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
		in.preexists != v.preexists {
		t.Error("Error: mismatch after decoding")
	}
}

func TestUnivalueCodeMeta(t *testing.T) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()

	meta := commutative.NewMeta()

	meta.(*commutative.Meta).SetSubDirs([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Meta).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Meta).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	bytes := in.Encode()
	out := (&Univalue{}).Decode(bytes).(*Univalue)

	if !in.Value().(*commutative.Meta).Equal(out.Value().(*commutative.Meta)) {
		t.Error("Error")
	}
}

func TestCodecMetaUnivalues(t *testing.T) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()

	meta := commutative.NewMeta()
	meta.(*commutative.Meta).SetSubDirs([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Meta).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Meta).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})
	unival := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta)

	in := Univalues([]ccurlcommon.UnivalueInterface{unival, unival})

	bytes := in.Encode()
	out := (&Univalues{}).Decode(bytes).(Univalues)

	// in.

	for i := 0; i < len(out); i++ {
		fmt.Print(in[i].Value().(*commutative.Meta).Value())
		fmt.Println(out[i].Value().(*commutative.Meta).Value())

		fmt.Print(in[i].Value().(*commutative.Meta).Added())
		fmt.Println(out[i].Value().(*commutative.Meta).Added())

		fmt.Print(in[i].Value().(*commutative.Meta).Removed())
		fmt.Println(out[i].Value().(*commutative.Meta).Removed())
		// fmt.Println(out[i])
	}
}

func BenchmarkUnivalueEncodeDecode(t *testing.B) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()
	v := commutative.NewMeta()
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
