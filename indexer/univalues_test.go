package indexer

import (
	"fmt"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/datacompression"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"

	"github.com/holiman/uint256"
)

/* Commutative Int64 Test */
func TestUnivaluesCodecUint64(t *testing.T) {
	alice := datacompression.RandomAccount()

	u64 := commutative.NewUint64(0, 100)
	in0 := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u64-000", 3, 4, 0, u64)
	// in0.reads = 1
	// in0.writes = 2
	// in0.deltaWrites = 3

	u256 := commutative.NewU256(uint256.NewInt(0), uint256.NewInt(100))
	in1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u256-000", 3, 4, 0, u256)
	// in1.reads = 4
	// in1.writes = 5
	// in1.deltaWrites = 6

	meta := commutative.NewPath()
	meta.(*commutative.Path).SetSubs([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Path).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	in2 := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta)
	// in2.reads = 7
	// in2.writes = 8
	// in2.deltaWrites = 9

	in := []interfaces.Univalue{in0, in1, in2}
	buffer := Univalues([]interfaces.Univalue{in0, in1, in2}).Encode()
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
	alice := datacompression.RandomAccount() /* Commutative Int64 Test */

	// meta:= commutative.NewPath()
	u256 := commutative.NewU256(uint256.NewInt(0), uint256.NewInt(100))
	in := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256)
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
	alice := datacompression.RandomAccount()

	meta := commutative.NewPath()
	meta.(*commutative.Path).SetSubs([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Path).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	in := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta)
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

func TestUnivaluesSelectiveEncoding(t *testing.T) {
	// alice := datacompression.RandomAccount()

	// numericU256 := Univalues([]interfaces.Univalue{ // Commutative numeric new with default and specified limits
	// 	(&Univalue{}).New(*NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/balance", 3, 4, 0, commutative.UINT256, false), commutative.NewU256(commutative.U256_MIN, commutative.U256_MAX), []byte{}).(*Univalue),
	// 	(&Univalue{}).New(*NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/balance", 3, 4, 0, commutative.UINT256, false), commutative.NewU256(uint256.NewInt(111), uint256.NewInt(999)), []byte{}).(*Univalue),
	// 	(&Univalue{}).New(*NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/balance", 3, 4, 1, commutative.UINT256, false), commutative.NewU256(commutative.U256_MIN, commutative.U256_MAX), []byte{}).(*Univalue),
	// 	(&Univalue{}).New(*NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/balance", 0, 4, 1, commutative.UINT256, false), commutative.NewU256(uint256.NewInt(111), uint256.NewInt(999)), []byte{}).(*Univalue),

	// 	(&Univalue{}).New(*NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/balance", 3, 4, 0, commutative.UINT256, false),
	// 		(&commutative.U256{}).NewU256(uint256.NewInt(10), uint256.NewInt(20), commutative.U256_MIN, commutative.U256_MAX, true), []byte{}).(*Univalue), // 4

	// 	(&Univalue{}).New(*NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/balance", 2, 0, 1, commutative.UINT256, false),
	// 		(&commutative.U256{}).NewU256(uint256.NewInt(10), uint256.NewInt(20), commutative.U256_MIN, uint256.NewInt(999), true), []byte{}).(*Univalue), // 5

	// 	(&Univalue{}).New(*NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/balance", 2, 0, 1, commutative.UINT256, false),
	// 		(&commutative.U256{}).NewU256(uint256.NewInt(10), uint256.NewInt(20), uint256.NewInt(111), commutative.U256_MAX, true), []byte{}).(*Univalue), // 6

	// 	(&Univalue{}).New(*NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/balance", 2, 0, 1, commutative.UINT256, false),
	// 		(&commutative.U256{}).NewU256(uint256.NewInt(10), uint256.NewInt(20), uint256.NewInt(111), uint256.NewInt(999), true), []byte{}).(*Univalue), // 7

	// 	(&Univalue{}).New(*NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/balance", 2, 0, 1, commutative.UINT256, false),
	// 		(&commutative.U256{}).NewU256(uint256.NewInt(10), uint256.NewInt(20), uint256.NewInt(111), uint256.NewInt(999), false), []byte{}).(*Univalue), // 8
	// })

	// encodedValues := codec.Byteset{}.Decode(numericU256.To(TransitionCodecFilterSet()...).Encode()).(codec.Byteset)

	// if (codec.Byteset{}.Decode(codec.Byteset{}.Decode(encodedValues[0]).(codec.Byteset)[1]).(codec.Byteset).Size()) != commutative.NewU256().(*commutative.U256).HeaderSize() {
	// 	t.Error("Error")
	// }

	// if (codec.Byteset{}.Decode(codec.Byteset{}.Decode(encodedValues[1]).(codec.Byteset)[1]).(codec.Byteset).Size()) != commutative.NewU256().(*commutative.U256).HeaderSize()+64 {
	// 	t.Error("Error")
	// }

	// if (codec.Byteset{}.Decode(codec.Byteset{}.Decode(encodedValues[2]).(codec.Byteset)[1]).(codec.Byteset).Size()) != commutative.NewU256().(*commutative.U256).HeaderSize() {
	// 	t.Error("Error")
	// }

	// if (codec.Byteset{}.Decode(codec.Byteset{}.Decode(encodedValues[3]).(codec.Byteset)[1]).(codec.Byteset).Size()) != commutative.NewU256().(*commutative.U256).HeaderSize()+64 {
	// 	t.Error("Error")
	// }

	// if (codec.Byteset{}.Decode(codec.Byteset{}.Decode(encodedValues[4]).(codec.Byteset)[1]).(codec.Byteset).Size()) != commutative.NewU256().(*commutative.U256).HeaderSize()+32+1 {
	// 	t.Error("Error")
	// }

	// if (codec.Byteset{}.Decode(codec.Byteset{}.Decode(encodedValues[5]).(codec.Byteset)[1]).(codec.Byteset).Size()) != commutative.NewU256().(*commutative.U256).HeaderSize()+32+32+1 {
	// 	t.Error("Error")
	// }

	// if (codec.Byteset{}.Decode(codec.Byteset{}.Decode(encodedValues[6]).(codec.Byteset)[1]).(codec.Byteset).Size()) != commutative.NewU256().(*commutative.U256).HeaderSize()+32+32+1 {
	// 	t.Error("Error")
	// }

	// if (codec.Byteset{}.Decode(codec.Byteset{}.Decode(encodedValues[7]).(codec.Byteset)[1]).(codec.Byteset).Size()) != commutative.NewU256().(*commutative.U256).HeaderSize()+32+32+32+1 {
	// 	t.Error("Error")
	// }
}

func BenchmarkUnivaluesEncodeDecode(t *testing.B) {
	/* Commutative Int64 Test */
	alice := datacompression.RandomAccount()
	v := commutative.NewPath()
	bytes := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 1, v).Encode()
	// bytes := in1.Encode()
	fmt.Println("Encoded length of one entry:", len(bytes)*4)

	in := make([]interfaces.Univalue, 1000000)
	for i := 0; i < len(in); i++ {
		in[i] = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 1, v)
	}

	t0 := time.Now()
	bytes = Univalues(in).Encode()
	fmt.Println("Encoded", len(in), "entires in :", time.Since(t0), "Total size: ", len(bytes)*4)

	t0 = time.Now()
	(Univalues([]interfaces.Univalue{})).Decode(bytes)
	fmt.Println("Decoded 100000 entires in :", time.Since(t0))
}
