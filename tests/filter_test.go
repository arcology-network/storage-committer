package ccurltest

import (
	"testing"
)

/* Commutative Int64 Test */
func TestUnivaluesFilter(t *testing.T) {
	// alice := datacompression.RandomAccount()

	// u64 := commutative.NewBoundedUint64(0, 100)
	// in0 := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u64-000", 3, 4, 0, u64)
	// // in0.reads = 1
	// // in0.writes = 2
	// // in0.deltaWrites = 3

	// u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))
	// in1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u256-000", 3, 4, 0, u256)
	// // in1.reads = 4
	// // in1.writes = 5
	// // in1.deltaWrites = 6

	// meta := commutative.NewPath()
	// meta.(*commutative.Path).SetSubs([]string{"e-01", "e-001", "e-002", "e-002"})
	// meta.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	// meta.(*commutative.Path).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	// in2 := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta)

	// trans := indexer.ITCTransition{}.From(in2) // Path
	// trans.(*univalue.Univalue).SetValue(commutative.NewPath())

	// in2.Value().(*commutative.Path).SetAdded([]string{"e-01", "e-001", "e-002", "e-002"})
	// if reflect.DeepEqual(in2.Value().(*commutative.Path).Added(), []string{}) {
	// 	t.Error("Error")
	// }

	// trans = indexer.IPCTransition{}.From(in2) // Path
	// trans.(*univalue.Univalue).SetValue(commutative.NewPath())

	// in2.Value().(*commutative.Path).SetAdded([]string{})
	// if !reflect.DeepEqual(in2.Value().(*commutative.Path).Added(), []string{}) {
	// 	t.Error("Error")
	// }
}
