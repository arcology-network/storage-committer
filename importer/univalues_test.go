package importer

import (
	"testing"

	datacompressor "github.com/arcology-network/common-lib/addrcompressor"
	"github.com/arcology-network/common-lib/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"

	"github.com/holiman/uint256"
)

/* Commutative Int64 Test */
func TestUnivaluesCodecPathMeta(t *testing.T) {
	alice := datacompressor.RandomAccount()

	u64 := commutative.NewBoundedUint64(0, 100)
	in0 := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u64-000", 3, 4, 0, u64, nil)
	// in0.reads = 1
	// in0.writes = 2
	// in0.deltaWrites = 3

	u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))
	in1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u256-000", 3, 4, 0, u256, nil)
	// in1.reads = 4
	// in1.writes = 5
	// in1.deltaWrites = 6

	meta := commutative.NewPath()
	meta.(*commutative.Path).SetSubs([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Path).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	in2 := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta, nil)
	// in2.reads = 7
	// in2.writes = 8
	// in2.deltaWrites = 9

	in := []*univalue.Univalue{in0, in1, in2}
	buffer := Univalues([]*univalue.Univalue{in0, in1, in2}).Encode()
	out := Univalues{}.Decode(buffer).(Univalues)

	if !Univalues(in).Equal(out) {
		t.Error("Error")
	}

	// Univalues(in).
	buffer = Univalues(in).Encode()
	out2 := Univalues{}.Decode(buffer).(Univalues)
	if !Univalues(in).Equal(out2) {
		t.Error("Error")
	}
}

func TestUnivaluesCodecU256(t *testing.T) {
	alice := datacompressor.RandomAccount() /* Commutative Int64 Test */

	// meta:= commutative.NewPath()
	u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))
	in := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256, nil)
	// in.reads = 1
	// in.writes = 2
	// in.deltaWrites = 3

	bytes := in.Encode()
	v := (&univalue.Univalue{}).Decode(bytes).(*univalue.Univalue)

	if in.TypeID() != v.TypeID() ||
		in.GetTx() != v.GetTx() ||
		*in.GetPath() != *v.GetPath() ||
		in.Writes() != v.Writes() ||
		in.DeltaWrites() != v.DeltaWrites() ||
		in.Preexist() != v.Preexist() {
		t.Error("Error: mismatch after decoding")
	}
}

func TestUnivaluesCodeMeta(t *testing.T) {
	/* Commutative Int64 Test */
	alice := datacompressor.RandomAccount()

	meta := commutative.NewPath()
	meta.(*commutative.Path).SetSubs([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Path).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	in := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta, nil)
	// in.reads = 1
	// in.writes = 2
	// in.deltaWrites = 3

	inKeys, _, _ := in.Value().(interfaces.Type).Get()

	bytes := in.Encode()
	out := (&univalue.Univalue{}).Decode(bytes).(*univalue.Univalue)
	outKeys, _, _ := out.Value().(interfaces.Type).Get()

	if !common.EqualArray(inKeys.([]string), outKeys.([]string)) {
		t.Error("Error")
	}
}
