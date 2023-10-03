package univalue

import (
	"testing"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/datacompression"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
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
	u64 := commutative.NewUint64(0, 100)
	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u64)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3
	in.preexists = true

	bytes := in.Encode()
	v := (&Univalue{}).Decode(bytes).(*Univalue)

	unimeta := v.GetUnimeta().(*Unimeta)
	inUnimeta := in.GetUnimeta().(*Unimeta)
	if !(*inUnimeta).Equal(unimeta) {
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
	u256 := commutative.NewU256(uint256.NewInt(0), uint256.NewInt(100))

	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	bytes := in.Encode()
	v := (&Univalue{}).Decode(bytes).(*Univalue)
	out := v.Value()

	raw := (*uint256.Int)(in.Value().(*commutative.U256).Value().(*codec.Uint256))

	outV := out.(*commutative.U256).Value().(*codec.Uint256)
	deltaV := in.Value().(*commutative.U256).Delta().(*codec.Uint256)

	flag := ((*uint256.Int)(deltaV)).Cmp((*uint256.Int)(out.(*commutative.U256).Delta().(*codec.Uint256))) != 0
	if raw.Cmp((*uint256.Int)(outV)) != 0 || flag {
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
	meta.(*commutative.Path).SetSubs([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Path).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	inKeys, _, _ := in.Value().(interfaces.Type).Get()

	bytes := in.Encode()
	out := (&Univalue{}).Decode(bytes).(*Univalue)
	outKeys, _, _ := out.Value().(interfaces.Type).Get()

	if !common.EqualArray(inKeys.([]string), outKeys.([]string)) {
		t.Error("Error")
	}
}

func TestUnimetaCodecUint64(t *testing.T) {
	/* Commutative Int64 Test */
	alice := datacompression.AliceAccount()

	// meta:= commutative.NewPath()
	u256 := commutative.NewUint64(0, 100).(*commutative.Uint64)
	in := NewUnimeta(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256.TypeID(), true, false)
	in.reads = 1
	in.writes = 2
	in.deltaWrites = 3

	bytes := in.Encode()
	out := (&Unimeta{}).Decode(bytes).(*Unimeta)

	if in == out {
		t.Error("Error")
	}
}
