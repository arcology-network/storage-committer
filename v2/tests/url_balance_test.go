package ccurltest

import (
	"math/big"
	"testing"

	ccurl "github.com/HPISTechnologies/concurrenturl/v2"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	ccurltype "github.com/HPISTechnologies/concurrenturl/v2/type"
	urltype "github.com/HPISTechnologies/concurrenturl/v2/type"
	commutative "github.com/HPISTechnologies/concurrenturl/v2/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/v2/type/noncommutative"
)

func TestSimpleBalance(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	_, transitions := url.Export(true)

	if err := url.Write(0, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(0))); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/alice/balance")
	}

	// Add the first delta
	if err := url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(22))); err != nil {
		t.Error(err, "blcc://eth1.0/account/alice/balance")
	}

	// Add the second delta
	if err := url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(11))); err != nil {
		t.Error(err, "blcc://eth1.0/account/alice/balance")
	}

	// Export variables
	_, transitions = url.Export(true)
	url.Indexer().Import(transitions)
	url.Indexer().Commit([]uint32{0, 1})

	// Read alice's balance again
	balance, _ := url.Read(1, "blcc://eth1.0/account/alice/balance")
	if balance.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(33)) != 0 {
		t.Error("Error: Wrong blcc://eth1.0/account/alice/balance value")
	}

}

func TestBalance(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	// create a path
	path, err := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Failed to create the path: blcc://eth1.0/account/alice/storage/ctrn-0/")
	}

	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path); err != nil {
		t.Error(err, " Failed to MakePath: blcc://eth1.0/account/alice/storage/ctrn-0/")
	}

	// create a noncommutative bigint
	inV := noncommutative.NewBigint(100)
	value := (*big.Int)(inV.(*noncommutative.Bigint))
	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-0", inV); err != nil {
		t.Error(err, " Failed to Write: blcc://eth1.0/account/alice/storage/ctrn-0/elem-0")
	}

	v, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-0")
	outV := (*big.Int)(v.(*noncommutative.Bigint))
	if outV.Cmp(value) != 0 {
		t.Error("Failed to read: blcc://eth1.0/account/alice/storage/ctrn-0/elem-0")
	}

	// -------------------Create another commutative bigint ------------------------------
	comtVInit := commutative.NewBalance(big.NewInt(300), big.NewInt(0))
	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/comt-0", comtVInit); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/comt-0", commutative.NewBalance(big.NewInt(300), big.NewInt(1))); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/comt-0", commutative.NewBalance(big.NewInt(300), big.NewInt(2))); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/comt-0")
	if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(303)) != 0 {
		t.Error("comt-0 has a wrong returned value")
	}

	// ----------------------------Balance ---------------------------------------------------
	if err := url.Write(0, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(0))); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/alice/balance")
	}

	// Add the first delta
	if err := url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(22))); err != nil {
		t.Error(err, "blcc://eth1.0/account/alice/balance")
	}

	// Add the second delta
	if err := url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(11))); err != nil {
		t.Error(err, "blcc://eth1.0/account/alice/balance")
	}

	// Read alice's balance
	v, _ = url.Read(1, "blcc://eth1.0/account/alice/balance")
	if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(33)) != 0 {
		t.Error("blcc://eth1.0/account/alice/balance")
	}

	// Export variables
	accessRecords, transitions := url.Export(true)
	in := ccurltype.Univalues(accessRecords).Encode()
	out := ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	for i := range accessRecords {
		if !accessRecords[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Accesses don't match")
		}
	}

	in = ccurltype.Univalues(transitions).Encode()
	out = ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	for i := range transitions {
		if !transitions[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Accesses don't match")
		}
	}

	url.Indexer().Import(transitions)
	url.Indexer().Commit([]uint32{0, 1})
	//url.Indexer().Store().Print()
}
