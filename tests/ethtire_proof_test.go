package committertest

import (
	"bytes"
	"fmt"
	"testing"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/exp/slice"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	common "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	importer "github.com/arcology-network/storage-committer/importer"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	opadapter "github.com/arcology-network/storage-committer/op"
	platform "github.com/arcology-network/storage-committer/platform"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	cache "github.com/arcology-network/storage-committer/storage/writecache"
	eucache "github.com/arcology-network/storage-committer/storage/writecache"
	univalue "github.com/arcology-network/storage-committer/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	ethmpt "github.com/ethereum/go-ethereum/trie"
)

func TestEthWorldTrieProof(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(common.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	bob := BobAccount()
	if _, err := adaptorcommon.CreateNewAccount(common.SYSTEM, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	FlushToStore(writeCache, store)

	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000009", noncommutative.NewString("435")); err != nil {
		t.Error(err)
	}
	FlushToStore(writeCache, store)

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil); err == nil {
		t.Error("Deleting an non-existing entry should've flaged an error", err)
	}

	raw := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	if acctTrans := raw; len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", noncommutative.NewString("124")); err != nil {
		t.Error("Failed to write", err)
	}

	if v, _, err := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil); v != "124" {
		t.Error("Wrong return value", err)
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.String)); v != "124" {
		t.Error("Error: Wrong return value")
	}

	FlushToStore(writeCache, store)

	/* Get Account Proofs */
	dstore := store.(*stgproxy.StorageProxy).EthStore()
	if _, err := dstore.IsAccountProvable(alice); err != nil {
		t.Error(err)
	}

	if d, err := dstore.IsAccountProvable(bob); err != nil || len(d) == 0 {
		t.Error(err)
	}

	if _, err := dstore.IsAccountProvable(CarolAccount()); err == nil {
		t.Error("Error: Should've flagged an error")
	}

	// bobStr := []byte(hexutil.MustDecode(bob))
	bobAddr := ethcommon.BytesToAddress(hexutil.MustDecode(bob))
	bobCache, _ := dstore.GetAccount(bobAddr, &ethmpt.AccessListCache{})
	if _, _, err := bobCache.IsStorageProvable("0x0000000000000000000000000000000000000000000000000000000000000000"); err != nil {
		t.Error(err)
	}

	aliceAddr := ethcommon.BytesToAddress(hexutil.MustDecode(alice))
	aliceCache, _ := dstore.GetAccount(aliceAddr, &ethmpt.AccessListCache{})
	if _, _, err := aliceCache.IsStorageProvable("0x0000000000000000000000000000000000000000000000000000000000000009"); err != nil {
		t.Error(err)
	}

	// Through the provider

	// roothash := store.(*stgproxy.StorageProxy).EthStore().Root()                               // Get the proof provider by a root hash.
	// ethdb := dstore.EthDB()                                                                  // Get the proof provider by a root hash.
	// provider, err := stgcommstorage.NewMerkleProofCache(2, ethdb).GetProofProvider(roothash) // Initiate the proof cache, maximum 2 blocks
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// // Verify Bob's storage for a big int value.
	// bobAddr = ethcommon.BytesToAddress([]byte(hexutil.MustDecode(bob)))
	// baccountResult, _ := provider.GetProof(bobAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000000")})
	// if err := baccountResult.Validate(provider.Root()); err != nil {
	// 	t.Error(err)
	// }
}

func TestGetProofAPI(t *testing.T) {
	// store := stgcommstorage.NewParallelEthMemDataStore()
	store := chooseDataStore()
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	bob := BobAccount()
	adaptorcommon.CreateNewAccount(common.SYSTEM, bob, writeCache)
	FlushToStore(writeCache, store)

	/* Bob updates */
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", noncommutative.NewBigint(1999))
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000002", noncommutative.NewBigint(1))
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003", noncommutative.NewBytes(ethcommon.BytesToHash([]byte{1}).Bytes()))
	FlushToStore(writeCache, store)

	// (stgproxy.StorageProxy(store).EthStore()
	roothash := store.(*stgproxy.StorageProxy).EthStore().Root()                     // Get the proof provider by a root hash.
	ethdb := store.(*stgproxy.StorageProxy).EthStore().EthDB()                       // Get the proof provider by a root hash.
	provider, err := ethstg.NewMerkleProofCache(2, ethdb).GetProofProvider(roothash) // Initiate the proof cache, maximum 2 blocks
	if err != nil {
		t.Fatal(err)
	}

	// Verify Bob's storage for a big int value.
	bobAddr := ethcommon.BytesToAddress([]byte(hexutil.MustDecode(bob)))
	accountResult, err := provider.GetProof(bobAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000001")})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	opProof := opadapter.Convertible(*accountResult).New() // Convert Bob's proof to OP format and verify.
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}

	bobAddr = ethcommon.BytesToAddress([]byte(hexutil.MustDecode(bob)))
	accountResult, err = provider.GetProof(bobAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000002")})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	opProof = opadapter.Convertible(*accountResult).New() // Convert Bob's proof to OP format and verify.
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}

	accountResult, err = provider.GetProof(bobAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000003")})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	opProof = opadapter.Convertible(*accountResult).New() // Convert Bob's proof to OP format and verify.
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}
}

func TestProofCacheBigInt(t *testing.T) {
	store := chooseDataStore()
	// store := hybridStore.(*stgproxy.StorageProxy).EthStore()
	// store := stgcommstorage.NewParallelEthMemDataStore()
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	adaptorcommon.CreateNewAccount(common.SYSTEM, alice, writeCache)

	/* Alice updates */
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001",
		noncommutative.NewBigint(12))

	v := slice.New[byte](5, byte(11))
	v = ethcommon.BytesToHash(v).Bytes()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003",
		noncommutative.NewBytes(v))

	v = slice.New[byte](32, byte(12))
	// v = ethcommon.BytesToHash(v).Bytes()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003",
		noncommutative.NewBytes(v))

	FlushToStore(writeCache, store)

	outv, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003",
		new(noncommutative.Bytes))

	fmt.Println("outv", outv.([]byte))

	if !bytes.Equal(outv.([]byte), v) {
		t.Error("Mismatch", outv, "!=", v)
	}

	roothash := store.(*stgproxy.StorageProxy).EthStore().Root()
	EthDB := store.(*stgproxy.StorageProxy).EthStore().EthDB()
	provider, err := ethstg.NewMerkleProofCache(2, EthDB).GetProofProvider(roothash)
	if err != nil {
		t.Fatal(err)
	}

	aliceAddr := ethcommon.BytesToAddress([]byte(hexutil.MustDecode(alice)))
	accountResult, err := provider.GetProof(aliceAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000003")}) // String
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	// Convert to OP format and verify.
	opProof := opadapter.Convertible(*accountResult).New() // To OP format
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}
}

func TestProofCacheNonNaitve(t *testing.T) {
	store := chooseDataStore()
	// store := stgcommstorage.NewParallelEthMemDataStore()
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	adaptorcommon.CreateNewAccount(common.SYSTEM, alice, writeCache)
	FlushToStore(writeCache, store)

	buf := slice.New[byte](32, 0)
	buf[31] = 1
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewBytes(buf))

	buf = slice.New[byte](32, 1)
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", noncommutative.NewBytes(buf))

	buf = slice.New[byte](33, 0)
	buf[32] = 1
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000002", noncommutative.NewBytes(buf))

	buf = slice.New[byte](33, 1)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003", noncommutative.NewBytes(buf)); err != nil {
		t.Error(err)
	}
	FlushToStore(writeCache, store)

	// Reads
	v, _ := writeCache.ReadOnlyDataStore().Retrive("blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.Bytes))
	buffer := v.(*noncommutative.Bytes).Value().(codec.Bytes)
	if buffer[0] != 1 {
		// t.Error("Mismatch", v)
	}

	v, _ = writeCache.ReadOnlyDataStore().Retrive("blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", new(noncommutative.Bytes))
	buffer = v.(*noncommutative.Bytes).Value().(codec.Bytes)
	if buffer[31] != 1 { // Native encoder will remove the prefix zeros, so the result is 1 bytes.
		t.Error("Mismatch", v)
	}

	// Big int encoder will trim the leading zeros, only keep the last 1, so when decoding, it will be 32 bytes with 31 zeros and 1
	v, _ = writeCache.ReadOnlyDataStore().Retrive("blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000002", new(noncommutative.Bytes))
	buffer = v.(*noncommutative.Bytes).Value().(codec.Bytes)
	if buffer[0] != 1 { // Native encoder will remove the prefix zeros, so the result is 1 bytes.
		// t.Error("Mismatch", v)
	}

	// Big int encoder will trim the leading zeros, only keep the last 1, so when decoding, it will be 32 bytes with 31 zeros and 1
	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003",
		new(noncommutative.Bytes)); v.([]byte)[31] != 1 {
		t.Error("Mismatch", v)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele0", noncommutative.NewString("124"))

	buf = slice.New[byte](33, 1)
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele1", noncommutative.NewBytes(buf))
	FlushToStore(writeCache, store)

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele1", new(noncommutative.Bytes)); !bytes.Equal(v.([]byte), slice.New[byte](33, 1)) {
		t.Error("Mismatch", v)
	}
}

func TestProofCache(t *testing.T) {
	store := chooseDataStore()
	store.(*stgproxy.StorageProxy).DisableCache()
	store.(*stgproxy.StorageProxy).EthStore().DisableAccountCache()

	// store := stgcommstorage.NewParallelEthMemDataStore()
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	adaptorcommon.CreateNewAccount(common.SYSTEM, alice, writeCache)

	bob := BobAccount()
	adaptorcommon.CreateNewAccount(common.SYSTEM, bob, writeCache)

	FlushToStore(writeCache, store)

	/* Alice updates */
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele0", noncommutative.NewString("124"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele1",
		noncommutative.NewString("abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456"))

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000007", noncommutative.NewBigint(2))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000009", noncommutative.NewInt64(19999))

	/* Bob updates */
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124"))
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", noncommutative.NewInt64(1))
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000002", noncommutative.NewBigint(1))

	buf := slice.New[byte](32, 0)
	buf[30] = 1
	buf[31] = 1
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003",
		noncommutative.NewBytes(buf))

	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/ele1", noncommutative.NewString("6789"))
	FlushToStore(writeCache, store)

	if str, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele0", new(noncommutative.String)); str != "124" {
		t.Fatal("String mismatch")
	}

	if str, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele1", new(noncommutative.String)); str != "abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456" {
		t.Fatal("String mismatch")
	}

	v, _ := writeCache.ReadOnlyDataStore().Retrive("blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003", new(noncommutative.Bytes))
	buffer := v.(*noncommutative.Bytes).Value().(codec.Bytes)

	if !bytes.Equal(buffer, []byte{1, 1}) { // Native encoder will remove the prefix zeros, so the result is 2 bytes.
		t.Fatal("String mismatch")
	}

	// Initiate the proof cache, maximum 2 blocks
	// cache := stgcommstorage.NewMerkleProofCache(2, store.EthDB())

	EthDB := store.(*stgproxy.StorageProxy).EthStore().EthDB()
	Root := store.(*stgproxy.StorageProxy).EthStore().Root()
	provider, err := ethstg.NewMerkleProofCache(2, EthDB).GetProofProvider(Root) // Get the proof provider by a root hash.
	if err != nil {
		t.Fatal(err)
	}

	bobAddr := ethcommon.BytesToAddress([]byte(hexutil.MustDecode(bob)))
	accountResult, err := provider.GetProof(bobAddr, []string{})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	// Verify Bob's storage for a string value.
	bobAddr = ethcommon.BytesToAddress([]byte(hexutil.MustDecode(bob)))
	accountResult, err = provider.GetProof(bobAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000000")}) // String
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	// Convert to OP format and verify.
	opProof := opadapter.Convertible(*accountResult).New() // To OP format
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}

	// Verify Bob's storage for an int64 value.
	bobAddr = ethcommon.BytesToAddress([]byte(hexutil.MustDecode(bob)))
	accountResult, err = provider.GetProof(bobAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000001")})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	opProof = opadapter.Convertible(*accountResult).New() // Convert Bob's proof to OP format and verify.
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}

	// Verify Bob's storage for an byte value. This shouldn't be provable because it's a non-native type that is larger than 32 bytes.
	// The original proof algorithm works with a byte array with maxiumn of 32 bytes, but this doesn't hold for non-native types.
	bobAddr = ethcommon.BytesToAddress([]byte(hexutil.MustDecode(bob)))
	accountResult, err = provider.GetProof(bobAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000003")})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	opProof = opadapter.Convertible(*accountResult).New() // Convert Bob's proof to OP format and verify.
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}

	// Verify Bob's storage for a big int value.
	bobAddr = ethcommon.BytesToAddress([]byte(hexutil.MustDecode(bob)))
	accountResult, err = provider.GetProof(bobAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000002")})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	opProof = opadapter.Convertible(*accountResult).New() // Convert Bob's proof to OP format and verify.
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}

	// Verify Alice's storage.
	aliceAddr := ethcommon.BytesToAddress([]byte(hexutil.MustDecode(alice)))
	accountResult, err = provider.GetProof(aliceAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000009")})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	// Convert Alice's proof to OP format and verify.
	opProof = opadapter.Convertible(*accountResult).New() // To OP format
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}

	// Verify Alice's storage on an non-existing account
	aliceAddr = ethcommon.BytesToAddress([]byte(hexutil.MustDecode(alice)))
	accountResult, err = provider.GetProof(aliceAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000019")})
	if err := accountResult.Validate(provider.Root()); err == nil {
		t.Error("This should fail")
	}
}

func TestHistoryProofs(t *testing.T) {
	store := chooseDataStore()
	store.(*stgproxy.StorageProxy).DisableCache()
	store.(*stgproxy.StorageProxy).EthStore().DisableAccountCache()

	writeCache := eucache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	adaptorcommon.CreateNewAccount(common.SYSTEM, alice, writeCache)
	FlushToStore(writeCache, store)

	/* Bob updates */
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", noncommutative.NewBigint(999))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000002", noncommutative.NewBigint(222))

	FlushToStore(writeCache, store)
	roothash0 := store.(*stgproxy.StorageProxy).EthStore().Root()
	verifierEthMerkle(roothash0, alice, "0x0000000000000000000000000000000000000000000000000000000000000001", store, t)

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", noncommutative.NewBigint(1999))
	FlushToStore(writeCache, store)
	roothash1 := store.(*stgproxy.StorageProxy).EthStore().Root()
	verifierEthMerkle(roothash1, alice, "0x0000000000000000000000000000000000000000000000000000000000000001", store, t)

	verifierEthMerkle(roothash0, alice, "0x0000000000000000000000000000000000000000000000000000000000000001", store, t)
	verifierEthMerkle(roothash0, alice, "0x0000000000000000000000000000000000000000000000000000000000000002", store, t)

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003", noncommutative.NewBigint(1999))
	FlushToStore(writeCache, store)

	roothash3 := store.(*stgproxy.StorageProxy).EthStore().Root()
	verifierEthMerkle(roothash3, alice, "0x0000000000000000000000000000000000000000000000000000000000000001", store, t)
}
