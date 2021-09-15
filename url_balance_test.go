package concurrenturl

import (
	"math/big"
	"testing"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	ccurltype "github.com/arcology-network/concurrenturl/type"
	urltype "github.com/arcology-network/concurrenturl/type"
	commutative "github.com/arcology-network/concurrenturl/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/type/noncommutative"
)

func TestSimpleBalance(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}
	url.Print()

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}

	if err := url.Write(0, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(0))); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	// Add the first delta
	if err := url.Write(1, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(22))); err != nil {
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	// Add the second delta
	if err := url.Write(1, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(11))); err != nil {
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	// // Read Alice's balance
	// v, _ = url.Read(1, "blcc://eth1.0/Alice/balance")
	// if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(233)) != 0 {
	// 	t.Error("blcc://eth1.0/Alice/balance")
	// }

	// Export variables
	accessRecords, transitions := url.Export()
	ccurltype.Univalues(accessRecords).Print()
	ccurltype.Univalues(transitions).Print()
}

func TestBalance(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}

	// create a path
	path, err := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Failed to create the path: blcc://eth1.0/Alice/storage/ctrn-0/")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", path); err != nil {
		t.Error(err, " Failed to MakePath: blcc://eth1.0/Alice/storage/ctrn-0/")
	}

	// create a noncommutative bigint
	inV := noncommutative.NewBigint(100)
	value := (*big.Int)(inV.(*noncommutative.Bigint))
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-0", inV); err != nil {
		t.Error(err, " Failed to Write: blcc://eth1.0/Alice/storage/ctrn-0/elem-0")
	}

	v, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-0")
	outV := (*big.Int)(v.(*noncommutative.Bigint))
	if outV.Cmp(value) != 0 {
		t.Error("Failed to read: blcc://eth1.0/Alice/storage/ctrn-0/elem-0")
	}

	// -------------------Create a commutative bigint ------------------------------
	comtVInit := commutative.NewBalance(big.NewInt(300), big.NewInt(0))
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/comt-0", comtVInit); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/comt-0", commutative.NewBalance(big.NewInt(300), big.NewInt(1))); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/comt-0", commutative.NewBalance(big.NewInt(300), big.NewInt(2))); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	v, _ = url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/comt-0")
	if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(303)) != 0 {
		t.Error("comt-0 has a wrong returned value")
	}

	// ----------------------------Balance ---------------------------------------------------
	if err := url.Write(0, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(0))); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	// Add the first delta
	if err := url.Write(1, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(22))); err != nil {
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	// Add the second delta
	if err := url.Write(1, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(11))); err != nil {
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	// Read Alice's balance
	v, _ = url.Read(1, "blcc://eth1.0/Alice/balance")
	if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(233)) != 0 {
		t.Error("blcc://eth1.0/Alice/balance")
	}

	// Export variables
	accessRecords, transitions := url.Export()
	in := ccurltype.Univalues(accessRecords).Encode()
	out := ccurltype.Univalues{}.Decode(in, &Decoder{}).(ccurltype.Univalues)
	for i := range accessRecords {
		if !accessRecords[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Accesses don't match")
		}
	}

	in = ccurltype.Univalues(transitions).Encode()
	out = ccurltype.Univalues{}.Decode(in, &Decoder{}).(ccurltype.Univalues)
	for i := range transitions {
		if !transitions[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Accesses don't match")
		}
	}

	url.indexer.Import(transitions)
	url.indexer.Commit([]uint32{0, 1})
	//url.indexer.Store().Print()
}
