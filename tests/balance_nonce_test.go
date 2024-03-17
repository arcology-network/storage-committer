package committertest

import (
	"math"
	"math/big"
	"testing"

	cache "github.com/arcology-network/eu/cache"
	stgcommitter "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	importer "github.com/arcology-network/storage-committer/importer"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	platform "github.com/arcology-network/storage-committer/platform"
	univalue "github.com/arcology-network/storage-committer/univalue"
	"github.com/holiman/uint256"
)

func TestSimpleBalance(t *testing.T) {
	store := chooseDataStore()

	committer := stgcommitter.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewUnboundedU256()); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the first delta
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(22), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the second delta
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(11), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Export variables
	in := univalue.Univalues((writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	buffer := univalue.Univalues(in).Encode()
	out := univalue.Univalues{}.Decode(buffer).(univalue.Univalues)
	for i := range in {
		if !in[i].Equal(out[i]) {
			t.Error("Accesses don't match")
		}
	}

	committer.Import(out)

	committer.Precommit([]uint32{0, 1})
	committer.Commit(0)
	// Read alice's balance again
	writeCache.Reset(writeCache)

	balance, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/balance", new(commutative.U256))
	balanceAddr := balance.(uint256.Int)
	if (&balanceAddr).Cmp(uint256.NewInt(33)) != 0 {
		t.Error("Error: Wrong blcc://eth1.0/account/alice/balance value")
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(10), true))
	balance, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/balance", new(commutative.U256))

	balanceAddr = balance.(uint256.Int)
	if (&balanceAddr).Cmp(uint256.NewInt(43)) != 0 {
		t.Error("Error: Wrong blcc://eth1.0/account/alice/balance value")
	}

	trans := univalue.Univalues((writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	records := univalue.Univalues((writeCache.Export(importer.Sorter))).To(importer.ITAccess{})

	univalue.Univalues(trans).Encode()
	for _, v := range records {
		if v.Writes() == v.Reads() && v.Writes() == 0 && v.DeltaWrites() == 0 {
			t.Error("Error: Write == Reads == DeltaWrites == 0")
		}
	}
}

func TestBalance(t *testing.T) {
	store := chooseDataStore()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// create a path
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err, " Failed to MakePath: blcc://eth1.0/account/alice/storage/ctrn-0/")
	}

	// create a noncommutative bigint
	inV := noncommutative.NewBigint(100)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", inV); err != nil {
		t.Error(err, " Failed to Write: blcc://eth1.0/account/alice/storage/ctrn-0/elem-0")
	}

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", new(noncommutative.Bigint))
	outV := v.(big.Int)
	value := (*big.Int)(inV.(*noncommutative.Bigint))
	if outV.Cmp(value) != 0 {
		t.Error("Failed to read: blcc://eth1.0/account/alice/storage/ctrn-0/elem-0")
	}

	// -------------------Create another commutative bigint ------------------------------
	comtVInit := commutative.NewBoundedU256(&commutative.U256_MIN, &commutative.U256_MAX)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", comtVInit); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(300), true)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(1), true)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(2), true)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", new(commutative.Path))
	vAdd := v.(uint256.Int)
	if vAdd.Cmp(uint256.NewInt(303)) != 0 {
		t.Error("comt-0 has a wrong returned value")
	}

	// ----------------------------U256 ---------------------------------------------------
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(0), true)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the first delta
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(22), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the second delta
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(11), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Read alice's balance
	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/balance", new(commutative.U256))
	vAdd = v.(uint256.Int)
	if vAdd.Cmp(uint256.NewInt(33)) != 0 {
		t.Error("blcc://eth1.0/account/" + alice + "/balance")
	}

	// Export variables
	transitions := univalue.Univalues((writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	// for i := range transitions {
	trans := transitions[9]

	_10 := trans.Encode()
	_10tran := (&univalue.Univalue{}).Decode(_10).(*univalue.Univalue)

	if !trans.Equal(_10tran) {
		t.Error("Accesses don't match", trans, _10tran)
	}
}

func TestNonce(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewBoundedUint64(0, math.MaxInt64)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(2)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(3)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	nonce, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/nonce", new(commutative.Uint64))
	v := nonce.(uint64)
	if v != 6 {
		t.Error("Error: blcc://eth1.0/account/alice/nonce should be ", 6)
	}

	trans := univalue.Univalues((writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(trans)

	committer.Precommit([]uint32{0})
	committer.Commit(0)
	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/nonce", new(commutative.Uint64))
	v = nonce.(uint64)
	if v != 6 {
		t.Error("Error: blcc://eth1.0/account/alice/nonce ")
	}
}

func TestMultipleNonces(t *testing.T) {
	store := chooseDataStore()
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUnboundedUint64())
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	trans0 := univalue.Univalues((writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	bob := BobAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, bob); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUnboundedUint64())
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+bob+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+bob+"/balance")
	}

	nonce, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce := nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}

	raw := (writeCache.Export(importer.Sorter))
	trans1 := univalue.Univalues(raw).To(importer.ITTransition{})

	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}

	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(trans0)
	committer.Import(trans1)

	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be 2", " actual: ", bobNonce)
	}

	committer = stgcommitter.NewStorageCommitter(store)

	committer.Precommit([]uint32{0})
	committer.Commit(0)

	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be 2", " actual: ", bobNonce)
	}

	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}
}
