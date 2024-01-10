package ccurltest

import (
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	committercommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	importer "github.com/arcology-network/concurrenturl/importer"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	ccurlstorage "github.com/arcology-network/concurrenturl/storage"
	storage "github.com/arcology-network/concurrenturl/storage"
	univalue "github.com/arcology-network/concurrenturl/univalue"
	cache "github.com/arcology-network/eu/cache"
	ethcommon "github.com/ethereum/go-ethereum/common"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	ethmpt "github.com/ethereum/go-ethereum/trie"
)

func TestConcurrentDB(t *testing.T) {
	store := chooseDataStore()

	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	bob := BobAccount()

	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, bob); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// if _, err := committer.NewAccount(committercommon.SYSTEM, bob); err != nil { // NewAccount account structure {
	// 	t.Error(err)
	// }

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		hash := ethcommon.BytesToHash(codec.Uint64(uint64(i)).Encode())
		if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), noncommutative.NewString(string(codec.Uint64(i).Encode()))); err != nil {
			t.Error(err)
		}

		if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), noncommutative.NewString(string(codec.Uint64(i).Encode()))); err != nil {
			t.Error(err)
		}
	}

	common.ParallelExecute(
		func() {
			for i := 1000; i < 2000; i++ {
				hash := ethcommon.BytesToHash(codec.Uint64(uint64(i)).Encode())
				if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), noncommutative.NewString("124")); err != nil {
					t.Error(err)
				}

				if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), noncommutative.NewString("124")); err != nil {
					t.Error(err)
				}
				time.Sleep(5 * time.Millisecond)
			}
			// },
			// func() {
			for i := 0; i < 1000; i++ {
				hash := ethcommon.BytesToHash(codec.Uint64(uint64(i)).Encode())
				if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), new(noncommutative.String)); v != string(codec.Uint64(i).Encode()) {
					t.Error("Mismatch")
				}

				if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), new(noncommutative.String)); v != string(codec.Uint64(i).Encode()) {
					t.Error("Mismatch")
				}
			}

		})
}

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

	raw := univalue.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
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
	if _, err := dstore.IsProvable(alice); err != nil {
		t.Error(err)
	}

	if d, err := dstore.IsProvable(bob); err != nil || len(d) == 0 {
		t.Error(err)
	}

	if _, err := dstore.IsProvable(CarolAccount()); err == nil {
		t.Error("Error: Should've flagged an error")
	}

	bobCache, _ := dstore.GetAccount(bob, &ethmpt.AccessListCache{})
	if _, err := bobCache.IsProvable("0x0000000000000000000000000000000000000000000000000000000000000000"); err != nil {
		t.Error(err)
	}

	aliceCache, _ := dstore.GetAccount(alice, &ethmpt.AccessListCache{})
	if _, err := aliceCache.IsProvable("0x0000000000000000000000000000000000000000000000000000000000000009"); err != nil {
		t.Error(err)
	}
}

func TestGetProofAPI(t *testing.T) {
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
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000009", noncommutative.NewInt64(1111))

	/* Bob updates */
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124"))
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", noncommutative.NewInt64(9999))
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/ele1", noncommutative.NewString("6789"))
	writeCache.FlushToDataSource(store)

	bobAcct, _ := store.GetAccount(bob, &ethmpt.AccessListCache{})

	if _, err := bobAcct.IsProvable("0x0000000000000000000000000000000000000000000000000000000000000000"); err != nil {
		t.Error(err)
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.String)); v != "124" {
		t.Error("Error: Wrong return value")
	}

	roothash := store.Root()

	proof, err := storage.NewMerkleProof(store.EthDB(), roothash)
	if err != nil {
		t.Error(err)
	}

	accountResult, err := proof.GetProof(bob, []string{string("0x0000000000000000000000000000000000000000000000000000000000000000")})
	if accountResult.StorageProof[0].Value.ToInt().Cmp(big.NewInt(0)) == 0 {
		t.Error(err)
	}

	manager := storage.NewMerkleProofManager(10, store.EthDB())
	accountResult, err = manager.GetProof(roothash, bob, []string{string("0x0000000000000000000000000000000000000000000000000000000000000000")})
	if accountResult.StorageProof[0].Value.ToInt().Cmp(big.NewInt(0)) == 0 {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/ele1", noncommutative.NewString("8976"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000009", noncommutative.NewInt64(1111))

	writeCache.FlushToDataSource(store)

	roothash2 := store.Root()
	if roothash == roothash2 {
		t.Error(errors.New("Error: Root hash should've changed"))
	}

	// dstore := committer.Importer().Store().(*storage.EthDataStore)
	if _, err := store.IsProvable(bob); err != nil {
		t.Error(err)
	}

	// Initiate a new proof provider
	proofProvider, err := storage.NewMerkleProof(store.EthDB(), roothash2)
	if err != nil {
		t.Error(err)
	}

	// Get the merkle proof for a storage k in Bob's account
	accountResult, err = proofProvider.GetProof(bob, []string{string("0x0000000000000000000000000000000000000000000000000000000000000000")})
	if err != nil {
		t.Error(err)
	}

	if accountResult.StorageProof[0].Value.ToInt().Cmp(big.NewInt(0)) == 0 {
		t.Error(err)
	}

	// Get the merkle proof for a storage k in Bob's account
	accountResult, err = proofProvider.GetProof(alice, []string{string("0x0000000000000000000000000000000000000000000000000000000000000009")})
	if err != nil {
		t.Error(err)
	}

	if err := accountResult.Verify(roothash2); err != nil {
		t.Error(err)
	}

	// accountResult.OpVerify(roothash2)

	// if accountResult.StorageProof[0].Value.ToInt().Cmp(big.NewInt(0)) == 0 {
	// 	t.Error(err)
	// }

	// t.Error(accountResult.OpVerify(roothash2))
}
