package ccurltest

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/arcology-network/common-lib/exp/array"
	committercommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	importer "github.com/arcology-network/concurrenturl/importer"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	opadapter "github.com/arcology-network/concurrenturl/op"
	ccurlstorage "github.com/arcology-network/concurrenturl/storage"
	storage "github.com/arcology-network/concurrenturl/storage"
	univalue "github.com/arcology-network/concurrenturl/univalue"
	cache "github.com/arcology-network/eu/cache"
	ethcommon "github.com/ethereum/go-ethereum/common"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	ethmpt "github.com/ethereum/go-ethereum/trie"
)

func TestEthWorldTrieProof(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore(nil, nil, nil, committercommon.Codec{}.Encode, committercommon.Codec{}.Decode)

	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, bob); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	writeCache.FlushToDataSource(store)

	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000009", noncommutative.NewString("435")); err != nil {
		t.Error(err)
	}
	writeCache.FlushToDataSource(store)

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil); err == nil {
		t.Error("Deleting an non-existing entry should've flaged an error", err)
	}

	raw := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
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

	writeCache.FlushToDataSource(store)

	/* Get Account Proofs */
	dstore := store.(*storage.EthDataStore)
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
}

func TestGetProofAPI(t *testing.T) {
	store := ccurlstorage.NewParallelEthMemDataStore()
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	bob := BobAccount()
	writeCache.CreateNewAccount(committercommon.SYSTEM, bob)
	writeCache.FlushToDataSource(store)

	/* Bob updates */
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", noncommutative.NewBigint(1999))
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000002", noncommutative.NewBigint(1))
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003", noncommutative.NewBytes(ethcommon.BytesToHash([]byte{1}).Bytes()))
	writeCache.FlushToDataSource(store)

	roothash := store.Root()                                                                       // Get the proof provider by a root hash.
	provider, err := ccurlstorage.NewMerkleProofCache(2, store.EthDB()).GetProofProvider(roothash) // Initiate the proof cache, maximum 2 blocks
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
	store := ccurlstorage.NewParallelEthMemDataStore()
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	alice := AliceAccount()
	writeCache.CreateNewAccount(committercommon.SYSTEM, alice)

	/* Alice updates */
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001",
		noncommutative.NewBigint(12))

	v := array.New[byte](5, byte(11))
	v = ethcommon.BytesToHash(v).Bytes()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003",
		noncommutative.NewBytes(v))

	v = array.New[byte](32, byte(12))
	// v = ethcommon.BytesToHash(v).Bytes()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003",
		noncommutative.NewBytes(v))

	writeCache.FlushToDataSource(store)

	outv, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003",
		new(noncommutative.Bytes))

	fmt.Println("outv", outv.([]byte))

	if !bytes.Equal(outv.([]byte), v) {
		t.Error("Mismatch", outv, "!=", v)
	}

	roothash := store.Root()
	provider, err := ccurlstorage.NewMerkleProofCache(2, store.EthDB()).GetProofProvider(roothash)
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
	store := ccurlstorage.NewParallelEthMemDataStore()
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	alice := AliceAccount()
	writeCache.CreateNewAccount(committercommon.SYSTEM, alice)
	writeCache.FlushToDataSource(store)

	buf := array.New[byte](32, 0)
	buf[31] = 1
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewBytes(buf))

	buf = array.New[byte](32, 1)
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", noncommutative.NewBytes(buf))

	buf = array.New[byte](33, 0)
	buf[32] = 1
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000002", noncommutative.NewBytes(buf))

	buf = array.New[byte](33, 1)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003", noncommutative.NewBytes(buf)); err != nil {
		t.Error(err)
	}
	writeCache.FlushToDataSource(store)

	// Reads
	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.Bytes)); v.([]byte)[31] != 1 {
		t.Error("Mismatch", v)
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", new(noncommutative.Bytes)); v.([]byte)[31] != 1 {
		t.Error("Mismatch", v)
	}

	// Big int encoder will trim the leading zeros, only keep the last 1, so when decoding, it will be 32 bytes with 31 zeros and 1
	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000002", new(noncommutative.Bytes)); v.([]byte)[31] != 1 {
		t.Error("Mismatch", v)
	}

	// Big int encoder will trim the leading zeros, only keep the last 1, so when decoding, it will be 32 bytes with 31 zeros and 1
	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003", new(noncommutative.Bytes)); v.([]byte)[31] != 1 {
		t.Error("Mismatch", v)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele0", noncommutative.NewString("124"))

	buf = array.New[byte](33, 1)
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele1", noncommutative.NewBytes(buf))
	writeCache.FlushToDataSource(store)

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele1", new(noncommutative.Bytes)); !bytes.Equal(v.([]byte), array.New[byte](33, 1)) {
		t.Error("Mismatch", v)
	}
}

func TestProofCache(t *testing.T) {
	store := ccurlstorage.NewParallelEthMemDataStore()
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	alice := AliceAccount()
	writeCache.CreateNewAccount(committercommon.SYSTEM, alice)

	bob := BobAccount()
	writeCache.CreateNewAccount(committercommon.SYSTEM, bob)

	writeCache.FlushToDataSource(store)

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

	buf := array.New[byte](32, 0)
	buf[31] = 1
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003",
		noncommutative.NewBytes(buf))

	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/ele1", noncommutative.NewString("6789"))
	writeCache.FlushToDataSource(store)

	if str, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele0", new(noncommutative.String)); str != "124" {
		t.Fatal("String mismatch")
	}

	if str, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele1", new(noncommutative.String)); str != "abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456abcef123456" {
		t.Fatal("String mismatch")
	}

	if str, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000003", new(noncommutative.Bytes)); !bytes.Equal(str.([]byte), buf) {
		t.Fatal("String mismatch")
	}

	// Initiate the proof cache, maximum 2 blocks
	cache := ccurlstorage.NewMerkleProofCache(2, store.EthDB())

	roothash := store.Root()
	provider, err := ccurlstorage.NewMerkleProofCache(2, store.EthDB()).GetProofProvider(roothash) // Get the proof provider by a root hash.
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

	// Simulate 5 consecutive blocks, record the root hashes and the keys.
	historyRoots := []ethcommon.Hash{}
	keys := []string{}
	for i := 5; i < 10; i++ {
		k := "0x000000000000000000000000000000000000000000000000000000000000000" + fmt.Sprint(i)

		writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+k, noncommutative.NewInt64(int64(i)))
		writeCache.FlushToDataSource(store)

		keys = append(keys, k)
		historyRoots = append(historyRoots, store.Root())
	}

	// Get the proof provider by a root hash from the history.
	for i := 0; i < len(historyRoots); i++ {
		provider, err := cache.GetProofProvider(historyRoots[i])
		if err != nil {
			t.Fatal(err)
		}

		accountResult, err := provider.GetProof(aliceAddr, []string{keys[i]})
		if err := accountResult.Validate(provider.Root()); err != nil {
			t.Error(err)
		}
	}
}
