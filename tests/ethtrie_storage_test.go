package committertest

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/common-lib/merkle"
	datastore "github.com/arcology-network/common-lib/storage/datastore"
	cache "github.com/arcology-network/eu/cache"
	stgcommitter "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	importer "github.com/arcology-network/storage-committer/importer"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	platform "github.com/arcology-network/storage-committer/platform"
	storage "github.com/arcology-network/storage-committer/storage"
	univalue "github.com/arcology-network/storage-committer/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	"github.com/holiman/uint256"
)

func TestConcurrentDB(t *testing.T) {
	store := chooseDataStore()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, bob); err != nil { // NewAccount account structure {
		t.Error(err)
	}

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

// TestTrieUpdates tests the updates to the trie data structure.
// It creates multiple accounts and performs write operations on their storage.
// It checks the correctness of the storage updates and cache management.
func TestTrieUpdates(t *testing.T) {
	store := chooseDataStore()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, bob); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	carol := CarolAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, carol); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	trans := writeCache.Export(importer.Sorter)
	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues(slice.Clone(trans)).To(importer.IPTransition{}))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})

	ds := committer.Store().(*storage.StoreRouter).EthStore()
	if (len(ds.AccountDict())) != 3 {
		t.Error("Error: Cache() should be 3", len(ds.AccountDict()))
	}
	committer.Commit(0)

	committer.Init(store)
	writeCache.Reset(writeCache)

	if _, err := writeCache.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele-0", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if len(ds.DirtyAccounts()) != 0 {
		t.Error("Error: DirtyAccounts() should be 0, actual", len(ds.DirtyAccounts()))
	}

	if (len(ds.AccountDict())) != 3 {
		t.Error("Error: Cache() should be 3, actual", len(ds.AccountDict()))
	}

	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{}))
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})

	// aliceAddr := ethcommon.BytesToAddress(hexutil.MustDecode(alice))
	// if len(ds.Dirties()) != 1 || ds.Dirties()[0].Address() != aliceAddr || !ds.Dirties()[0].StorageDirty {
	// 	t.Error("Error: Dirties() should be 1, actual", len(ds.Dirties()))
	// }
	committer.Commit(0)

	committer.Init(store)
	writeCache.Reset(writeCache)

	if _, err := writeCache.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(100), true)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(uint64(11))); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/code", noncommutative.NewBytes([]byte{1, 2, 3, 4})); err != nil {
		t.Error(err)
	}

	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{}))
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})

	// if len(ds.Dirties()) != 1 || ds.Dirties()[0].Address() != aliceAddr || ds.Dirties()[0].StorageDirty {
	// 	t.Error("Error: Dirties() should be 1, actual", len(ds.Dirties()))
	// }

	if (len(ds.AccountDict())) != 3 {
		t.Error("Error: Cache() should be 3, actual", len(ds.AccountDict()))
	}
}

// need to hash the keys first
func TestEthStorageConnection(t *testing.T) {
	store := chooseDataStore()
	// store = chooseDataStore()

	alice := AliceAccount()

	// writeCache := committer.WriteCache()
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := writeCache.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	u256 := new(commutative.U256).NewBoundedU256FromUint64(111, 0, 0, 999, true)
	if _, err := writeCache.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/1", u256); err != nil {
		t.Error(err)
	}

	trans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(trans)

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)

	writeCache = cache.NewWriteCache(store, 1, 1, platform.NewPlatform()) // Reset the write cache
	v, _, err := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path))
	if v == nil {
		t.Error(err)
	}

	v, _, _ = writeCache.Read(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/1", new(commutative.U256))
	typedv := v.(uint256.Int)
	if v == nil || typedv.Cmp(uint256.NewInt(111)) != 0 {
		t.Error(err, v)
	}
}

func TestBasicAddRead(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()

	alice := AliceAccount()

	// writeCache := committer.WriteCache()
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// create a path
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", noncommutative.NewString("path")); err == nil {
		t.Error(err)
	}

	// Try to read an nonexistent path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1", nil); value != nil {
		t.Error("Error: Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-000", nil); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// try again
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", nil); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// try to read an nonexistent path
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", nil); value != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/elem-000")
	}

	// Write the entry
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/elem-000")
	}

	// Write the entry
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-111", noncommutative.NewInt64(9999)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/elem-111")
	}

	// if v, _ := committer.Find(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", noncommutative.NewInt64(1111)); v != nil {
	// 	t.Error("Error: The path should have been deleted")
	// }

	// Read the entry back
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", new(noncommutative.Int64)); value.(int64) != 1111 {
		t.Error("Error: Wrong value")
	}

	// Read the path
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error(value)
	} else {
		target := value.(*deltaset.DeltaSet[string]).Elements()
		if !reflect.DeepEqual(target, []string{"elem-000", "elem-111"}) {
			t.Error("Error: Wrong value !!!!")
		}
	}
}

func TestAddThenDeletePathInEthTrie(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()

	alice := AliceAccount()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans := writeCache.Export(importer.Sorter)
	trans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	acctTrans := (&univalue.Univalues{}).Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues)

	//values := univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).([]*univalue.Univalue)
	ts := univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues)

	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(ts)

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)
	committer.Init(store)

	writeCache.Reset(writeCache)
	// create a path
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	transitions := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import((&univalue.Univalues{}).Decode(univalue.Univalues(transitions).Encode()).(univalue.Univalues))
	committer.Precommit([]uint32{1})
	committer.Commit(0)

	writeCache.Reset(writeCache)
	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{})
	if v == nil {
		t.Error("Error: The path should exist")
	}

	committer.Init(store)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", nil); err != nil { // Delete the path
		t.Error(err)
	}

	trans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import((&univalue.Univalues{}).Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))
	committer.Precommit([]uint32{1})
	committer.Commit(0)

	writeCache.Reset(writeCache)
	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path)); v != nil {
		t.Error("Error: The path should have been deleted")
	}
}

func BenchmarkMultipleAccountCommitDataStore(b *testing.B) {
	// store := chooseDataStore() // Eth data store
	store := datastore.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode) // Native data store

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	// t0 := time.Now()
	for i := 0; i < 100000; i++ {
		acct := fmt.Sprint(rand.Int())
		if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, acct); err != nil { // NewAccount account structure {
			fmt.Println(err)
		}

		// if _, err := committer.NewAccount(stgcommcommon.SYSTEM, acct); err != nil { // NewAccount account structure {
		// 	fmt.Println(err)
		// }

		path := commutative.NewPath() // create a path
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+acct+"/storage/container/ctrn-0/", path); err != nil {
			b.Error(err)
		}

		for j := 0; j < 4; j++ {
			if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+acct+"/storage/container/ctrn-0/elem-0"+fmt.Sprint(j), noncommutative.NewString("fmt.Sprint(i)")); err != nil { /* The first Element */
				b.Error(err)
			}
		}
	}
}

func TestLevelDBBasic(t *testing.T) {
	leveldb, err := rawdb.NewLevelDBDatabase("./leveldb", 256, 16, "temp", false)
	if err != nil {
		return
	}

	diskdbs := [16]ethdb.Database{}
	slice.Fill(diskdbs[:], leveldb)
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	// db := trie.NewDatabase(leveldb)
	trie := trie.NewEmptyParallel(db)
	res := trie.Hash()
	exp := types.EmptyRootHash
	if res != exp {
		t.Errorf("expected %x got %x", exp, res)
	}

	keys := make([][]byte, 10)
	data := make([][]byte, len(keys))
	for i := 0; i < len(data); i++ {
		keys[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i)))
		data[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i)))
	}

	trie.ParallelUpdate(keys, data)

	for i, k := range keys {
		v, err := trie.Get(k)
		if err != nil || !bytes.Equal(v, data[i]) {
			t.Errorf("expected %x got %x", exp, res)
		}
	}

	if err := os.RemoveAll("./leveldb"); err != nil {
		t.Error(err)
	}
}

func BenchmarkLevelDBPerformance1M(t *testing.B) {
	leveldb, err := rawdb.NewLevelDBDatabase("./leveldb", 0, 16, "temp", false)
	if err != nil {
		return
	}

	diskdbs := [16]ethdb.Database{}
	slice.Fill(diskdbs[:], leveldb)
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	trie := trie.NewEmptyParallel(db)
	res := trie.Hash()
	exp := types.EmptyRootHash
	if res != exp {
		t.Errorf("expected %x got %x", exp, res)
	}

	keys := make([][]byte, 2000000)
	data := make([][]byte, len(keys))
	for i := 0; i < len(data); i++ {
		keys[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i)))
		data[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i)))
	}

	t0 := time.Now()
	trie.ParallelUpdate(keys, data)
	fmt.Println("Parallel Update ", len(keys), " entries in ", time.Since(t0))

	offset := len(keys)
	for i := 0; i < len(data); i++ {
		keys[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i + offset)))
		data[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i + offset)))
	}

	t0 = time.Now()
	slice.ParallelForeach(keys, 8, func(i int, _ *[]byte) { trie.Get(keys[i]) })
	fmt.Println("Parallel Get ", len(keys), " entries in ", time.Since(t0))

	t0 = time.Now()
	for i, k := range keys {
		trie.Update(k, data[i])
	}
	fmt.Println("Sequential Update ", len(keys), " entries in ", time.Since(t0))

	t0 = time.Now()
	for _, k := range keys {
		trie.Get(k)
	}
	fmt.Println("Sequential Get ", len(keys), " entries in ", time.Since(t0))

	if err := os.RemoveAll("./leveldb"); err != nil {
		t.Error(err)
	}
}
