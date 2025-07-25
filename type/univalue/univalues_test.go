/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package univalue

import (
	"math/rand"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/common-lib/exp/softdeltaset"
	stgcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/holiman/uint256"
)

func RandomAccount() string {
	var letters = []byte("abcdef0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 20)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	addr := hexutil.Encode(b)
	return addr
}

/* Commutative Int64 Test */
func TestUnivaluesCodecPathMeta(t *testing.T) {
	alice := RandomAccount()

	u64 := commutative.NewBoundedUint64(0, 100)
	in0 := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u64-000", 3, 4, 0, u64, nil)
	// in0.reads = 1
	// in0.writes = 2
	// in0.deltaWrites = 3

	u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))
	in1 := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u256-000", 3, 4, 0, u256, nil)
	// in1.reads = 4
	// in1.writes = 5
	// in1.deltaWrites = 6

	meta := commutative.NewPath()
	meta.(*commutative.Path).SetSubPaths([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Path).InsertRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	in2 := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, meta, nil)
	// in2.reads = 7
	// in2.writes = 8
	// in2.deltaWrites = 9

	in := []*Univalue{in0, in1, in2}
	buffer := Univalues([]*Univalue{in0, in1, in2}).Encode()
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
	alice := RandomAccount() /* Commutative Int64 Test */

	// meta:= commutative.NewPath()
	u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))
	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256, nil)
	// in.reads = 1
	// in.writes = 2
	// in.deltaWrites = 3

	bytes := in.Encode()
	v := (&Univalue{}).Decode(bytes).(*Univalue)

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
	alice := RandomAccount()

	path := commutative.NewPath()
	path.(*commutative.Path).SetSubPaths([]string{"e-01", "e-001", "e-002", "e-002"})
	path.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	path.(*commutative.Path).InsertRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	in := NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 11, path, nil)
	// in.reads = 1
	// in.writes = 2
	// in.deltaWrites = 3

	inKeys, _, _ := in.Value().(stgcommon.Type).Get()

	bytes := in.Encode()
	out := (&Univalue{}).Decode(bytes).(*Univalue)
	outSet, _, _ := out.Value().(stgcommon.Type).Get()

	if !slice.EqualSet(inKeys.(*softdeltaset.DeltaSet[string]).Elements(), outSet.(*softdeltaset.DeltaSet[string]).Elements()) {
		t.Error("Error")
	}

	inv := []*Univalue{}
	buffer := Univalues(inv).Encode()
	if v := new(Univalues).Decode(buffer).(Univalues); len(v) != 0 {
		t.Error("Error")
	}
}
