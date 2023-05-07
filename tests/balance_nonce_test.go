package ccurltest

import (
	"math"
	"math/big"
	"testing"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/univalue"
	"github.com/holiman/uint256"
)

func TestSimpleBalance(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url.Write(0, "blcc://eth1.0/account/"+alice+"/balance",
		commutative.NewU256(commutative.U256_MIN, commutative.U256_MAX)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the first delta
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/balance",
		commutative.NewU256Delta(uint256.NewInt(22), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the second delta
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/balance",
		commutative.NewU256Delta(uint256.NewInt(11), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Export variables
	// _, in := url.Export(univalue.Sorter)
	in := univalue.Univalues(common.Clone(url.Export(univalue.Sorter))).To(univalue.TransitionFilters()...)

	buffer := univalue.Univalues(in).Encode()
	out := univalue.Univalues{}.Decode(buffer).(univalue.Univalues)
	for i := range in {
		if !in[i].(*univalue.Univalue).Equal(out[i].(*univalue.Univalue)) {
			t.Error("Accesses don't match")
		}
	}

	url.Import(out)
	url.Sort()
	url.Commit([]uint32{0, 1})

	url = ccurl.NewConcurrentUrl(store)
	// Read alice's balance again
	url2 := ccurl.NewConcurrentUrl(store)
	balance, _ := url2.Read(1, "blcc://eth1.0/account/"+alice+"/balance")
	if balance.(*uint256.Int).Cmp(uint256.NewInt(33)) != 0 {
		t.Error("Error: Wrong blcc://eth1.0/account/alice/balance value")
	}

	url2.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(10), true))
	balance, _ = url2.Read(1, "blcc://eth1.0/account/"+alice+"/balance")
	if balance.(*uint256.Int).Cmp(uint256.NewInt(43)) != 0 {
		t.Error("Error: Wrong blcc://eth1.0/account/alice/balance value")
	}

	// records, trans := url2.Export(univalue.Sorter)
	trans := univalue.Univalues(common.Clone(url.Export(univalue.Sorter))).To(univalue.TransitionFilters()...)
	records := univalue.Univalues(common.Clone(url.Export(univalue.Sorter))).To(univalue.AccessFilters()...)

	univalue.Univalues(trans).Encode()
	for _, v := range records {
		if v.Writes() == v.Reads() && v.Writes() == 0 && v.DeltaWrites() == 0 {
			t.Error("Error: Write == Reads == DeltaWrites == 0")
		}
	}
}

func TestBalance(t *testing.T) {
	// compressionLut := datacompression.NewCompressionLut()
	store := cachedstorage.NewDataStore()

	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	// create a path
	path := commutative.NewPath()
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err, " Failed to MakePath: blcc://eth1.0/account/alice/storage/ctrn-0/")
	}

	// create a noncommutative bigint
	inV := noncommutative.NewBigint(100)
	value := (*big.Int)(inV.(*noncommutative.Bigint))
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", inV); err != nil {
		t.Error(err, " Failed to Write: blcc://eth1.0/account/alice/storage/ctrn-0/elem-0")
	}

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0")
	outV := v.(*big.Int)
	if outV.Cmp(value) != 0 {
		t.Error("Failed to read: blcc://eth1.0/account/alice/storage/ctrn-0/elem-0")
	}

	// -------------------Create another commutative bigint ------------------------------
	comtVInit := commutative.NewU256(commutative.U256_MIN, commutative.U256_MAX)
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", comtVInit); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(300), true)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(1), true)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(2), true)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0")
	if v.(*uint256.Int).Cmp(uint256.NewInt(303)) != 0 {
		t.Error("comt-0 has a wrong returned value")
	}

	// ----------------------------U256 ---------------------------------------------------
	if err := url.Write(0, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(0), true)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the first delta
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(22), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the second delta
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(11), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Read alice's balance
	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/balance")
	if v.(*uint256.Int).Cmp(uint256.NewInt(33)) != 0 {
		t.Error("blcc://eth1.0/account/" + alice + "/balance")
	}

	// Export variables
	transitions := univalue.Univalues(common.Clone(url.Export(univalue.Sorter))).To(univalue.TransitionFilters()...)
	// for i := range transitions {
	trans := transitions[9]

	_10 := trans.Encode()
	_10tran := (&univalue.Univalue{}).Decode(_10).(*univalue.Univalue)

	if !trans.(*univalue.Univalue).Equal(_10tran) {
		t.Error("Accesses don't match", trans, _10tran)
	}
}

func TestNonce(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url1 := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url1.CreateAccount(ccurlcommon.SYSTEM, url1.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url1.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64(0, math.MaxInt64)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if err := url1.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if err := url1.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(2)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if err := url1.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(3)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	nonce, _ := url1.Read(0, "blcc://eth1.0/account/"+alice+"/nonce")
	v := nonce.(uint64)
	if v != 6 {
		t.Error("Error: blcc://eth1.0/account/alice/nonce should be ", 6)
	}

	trans := univalue.Univalues(common.Clone(url1.Export(univalue.Sorter))).To(univalue.TransitionFilters()...)
	url1.Import(trans)
	url1.Sort()
	url1.Commit([]uint32{0})

	nonce, _ = url1.Read(0, "blcc://eth1.0/account/"+alice+"/nonce")
	v = nonce.(uint64)
	if v != 6 {
		t.Error("Error: blcc://eth1.0/account/alice/nonce ")
	}
}

func TestMultipleNonces(t *testing.T) {
	store := cachedstorage.NewDataStore()

	url0 := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if err := url0.CreateAccount(ccurlcommon.SYSTEM, url0.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	url0.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64(0, math.MaxInt64))
	if err := url0.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if err := url0.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// _, trans0 := url0.Export(univalue.Sorter)
	// ccurltype.SetInvariate(trans0, "nonce")
	trans0 := univalue.Univalues(common.Clone(url0.Export(univalue.Sorter))).To(univalue.TransitionFilters()...)

	url1 := ccurl.NewConcurrentUrl(store)
	bob := BobAccount()
	if err := url1.CreateAccount(ccurlcommon.SYSTEM, url1.Platform.Eth10(), bob); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	url0.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64(0, math.MaxInt64))

	if err := url1.Write(0, "blcc://eth1.0/account/"+bob+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+bob+"/balance")
	}

	if err := url1.Write(0, "blcc://eth1.0/account/"+bob+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+bob+"/balance")
	}

	nonce, _ := url1.Read(0, "blcc://eth1.0/account/"+bob+"/nonce")
	bobNonce := nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}

	trans1 := univalue.Univalues(common.Clone(url1.Export(univalue.Sorter))).To(univalue.TransitionFilters()...)
	// ccurltype.SetInvariate(trans1, "nonce")

	nonce, _ = url1.Read(0, "blcc://eth1.0/account/"+bob+"/nonce")
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}

	url0.Import(trans0)
	url0.Import(trans1)

	nonce, _ = url1.Read(0, "blcc://eth1.0/account/"+bob+"/nonce")
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}

	url0.Sort()
	url0.Commit([]uint32{0})

	nonce, _ = url1.Read(0, "blcc://eth1.0/account/"+bob+"/nonce")
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}

	nonce, _ = url0.Read(0, "blcc://eth1.0/account/"+bob+"/nonce")
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}
}
