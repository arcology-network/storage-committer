package ccurltest

import (
	"math/big"
	"testing"

	ccurl "github.com/arcology-network/concurrenturl/v2"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/type/noncommutative"
)

func TestSimpleBalance(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // CreateAccount account structure {
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

	in := ccurltype.Univalues(transitions).Encode()
	out := ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	url.Import(out)
	url.Commit([]uint32{0, 1})

	// Read alice's balance again
	balance, _ := url.Read(1, "blcc://eth1.0/account/alice/balance")
	if balance.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(33)) != 0 {
		t.Error("Error: Wrong blcc://eth1.0/account/alice/balance value")
	}

}

func TestBalance(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // CreateAccount account structure {
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
		if !accessRecords[i].(*ccurltype.Univalue).EqualTransition(out[i].(*ccurltype.Univalue)) {
			t.Error("Accesses don't match")
		}
	}

	in = ccurltype.Univalues(transitions).Encode()
	out = ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	for i := range transitions {
		if !transitions[i].(*ccurltype.Univalue).EqualTransition(out[i].(*ccurltype.Univalue)) {
			t.Error("Accesses don't match")
		}
	}

	url.Indexer().Import(out)
	url.Indexer().Commit([]uint32{0, 1})
	//url.Indexer().Store().Print()
}

func TestNonce(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url1 := ccurl.NewConcurrentUrl(store)
	if err := url1.CreateAccount(ccurlcommon.SYSTEM, url1.Platform.Eth10(), "alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url1.Write(0, "blcc://eth1.0/account/alice/nonce", commutative.NewInt64(10, 100)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/alice/balance")
	}

	if err := url1.Write(0, "blcc://eth1.0/account/alice/nonce", commutative.NewInt64(10, 9)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/alice/balance")
	}

	nonce, _ := url1.Read(0, "blcc://eth1.0/account/alice/nonce")
	v := nonce.(ccurlcommon.TypeInterface).(*commutative.Int64).Value().(int64)
	if v != 109 {
		t.Error("Error: blcc://eth1.0/account/alice/nonce ")
	}
}
