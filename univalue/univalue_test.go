package univalue

import (
	"testing"

	"github.com/arcology-network/common-lib/addrcompressor"
	set "github.com/arcology-network/common-lib/container/set"
	"github.com/arcology-network/common-lib/exp/slice"
	commutative "github.com/arcology-network/storage-committer/commutative"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/holiman/uint256"
)

func AliceAccount() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	// rand.Seed(1)
	b := make([]rune, 40)
	for i := range b {
		b[i] = letters[1]
	}
	return string(b)
}

func TestUnivalueCodecUint64(t *testing.T) {
	/* Commutative Int64 Test */
	alice := AliceAccount()

	// meta:= commutative.NewPath()
	u64 := commutative.NewBoundedUint64(0, 100)
	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u64, nil)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3
	in.preexists = true

	bytes := in.Encode()
	v := (&Univalue{}).Decode(bytes).(*Univalue)

	property := v.Property
	inProperty := in.Property
	if !(inProperty).Equal(&property) {
		t.Error("Error")
	}

	out := v.Value()

	if !(in.value.(*commutative.Uint64)).Equal(out.(*commutative.Uint64)) {
		t.Error("Error")
	}
}

func TestUnivalueCodecU256(t *testing.T) {
	alice := AliceAccount() /* Commutative Int64 Test */

	// meta:= commutative.NewPath()
	u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))

	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256, nil)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	bytes := in.Encode()
	v := (&Univalue{}).Decode(bytes).(*Univalue)
	out := v.Value()

	raw := in.Value().(*commutative.U256).Value().(uint256.Int)

	outV := out.(*commutative.U256).Value().(uint256.Int)
	deltaV := in.Value().(*commutative.U256).Delta().(uint256.Int)

	otherv := out.(*commutative.U256).Delta().(uint256.Int)
	flag := (&deltaV).Cmp(&(otherv)) != 0
	if raw.Cmp((*uint256.Int)(&outV)) != 0 || flag {
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
	alice := AliceAccount()

	meta := commutative.NewPath()
	meta.(*commutative.Path).SetSubPaths([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Path).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta, nil)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	inKeys, _, _ := in.Value().(intf.Type).Get()

	bytes := in.Encode()
	out := (&Univalue{}).Decode(bytes).(*Univalue)
	outKeys, _, _ := out.Value().(intf.Type).Get()

	if !slice.EqualSet(inKeys.(*set.OrderedSet).Keys(), outKeys.(*set.OrderedSet).Keys()) {
		t.Error("Error")
	}
}

func TestPropertyCodecUint64(t *testing.T) {
	/* Commutative Int64 Test */
	alice := addrcompressor.AliceAccount()

	// meta:= commutative.NewPath()
	u256 := commutative.NewBoundedUint64(0, 100).(*commutative.Uint64)
	in := NewProperty(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256.TypeID(), true, false)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	bytes := in.Encode()
	out := (&Property{}).Decode(bytes).(*Property)

	if in == out {
		t.Error("Error")
	}
}

// func BenchmarkAccountMerkleImportPerf(t *testing.B) {
// 	data := [][]byte{}
// 	for i := 0; i < 1000000; i++ {
// 		v := sha256.Sum256([]byte(fmt.Sprint(i)))
// 		data = append(data, v[:])
// 	}

// 	t0 := time.Now()
// 	s1 := codec.Byteset(data).Encode()
// 	codec.Byteset{}.Decode(s1)
// 	fmt.Println("Code.Byteset: ", time.Since(t0), len(s1))

// 	t0 = time.Now()
// 	s2 := ethrlp.Bytes{}.Encode(data)
// 	_, err := ethrlp.Bytes{}.Decode(s2)

// 	rlp.EncodeToBytes(s2)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Println("ethrlp.Bytes{}.Encode: ", time.Since(t0), len(s2), float64(len(s1))/float64(len(s2)))
// }
