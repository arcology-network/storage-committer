package ccurltest

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	"github.com/arcology-network/common-lib/merkle"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	storage "github.com/arcology-network/concurrenturl/storage"
	univalue "github.com/arcology-network/concurrenturl/univalue"
	"github.com/arcology-network/evm/core/rawdb"
	"github.com/arcology-network/evm/core/types"
	"github.com/arcology-network/evm/ethdb"
	"github.com/arcology-network/evm/ethdb/memorydb"
	"github.com/arcology-network/evm/trie"
	ethmpt "github.com/arcology-network/evm/trie"
)

func TestEthTrieBasic(t *testing.T) {
	store := storage.NewParallelEthMemDataStore()
	keys := []string{
		"blcc://eth1.0/account/abbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbc/storage/container/ctrn-0/",
		"blcc://eth1.0/account/abbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbc/storage/native/" + string(codec.Bytes32([32]byte{1}).Encode()),
		"blcc://eth1.0/account/abbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbc/storage/native/" + string(codec.Bytes32([32]byte{2}).Encode()),
	}

	vals := []interface{}{
		univalue.NewUnivalue(0, "", 0, 0, 0, commutative.NewBoundedUint64(1, 111), nil),
		univalue.NewUnivalue(0, "", 0, 0, 0, commutative.InitNewPaths([]string{"ctrn-0"}), nil),
		univalue.NewUnivalue(0, "", 0, 0, 0, noncommutative.NewInt64(199), nil),
	}

	store.Precommit(keys, vals)

	v, err := store.Retrive(keys[0], new(commutative.Uint64))
	if v == nil {
		t.Error(err)
	}

	if v == nil || !vals[0].(interfaces.Univalue).Value().(interfaces.Type).Equal(v) {
		t.Error("Expeced :", vals[0].(interfaces.Univalue).Value().(interfaces.Type))
		t.Error("Actual; :", v)
	}

	v, err = store.Retrive(keys[1], new(commutative.Path))
	if v == nil {
		t.Error(err)
	}

	if v == nil || !vals[1].(interfaces.Univalue).Value().(interfaces.Type).Equal(v) {
		t.Error("Expeced :", vals[0].(interfaces.Univalue).Value().(interfaces.Type))
		t.Error("Actual; :", v)
	}

	v, err = store.Retrive(keys[2], new(noncommutative.Int64))
	if v == nil {
		t.Error(err)
	}

	if v == nil || !vals[2].(interfaces.Univalue).Value().(interfaces.Type).Equal(v) {
		t.Error("Expeced :", vals[0].(interfaces.Univalue).Value().(interfaces.Type))
		t.Error("Actual; :", v)
	}

	store.Commit() // Calculate root hash
}

func TestEthTrieBasicProof(t *testing.T) {
	store := storage.NewParallelEthMemDataStore()

	aliceKeys := []string{
		"blcc://eth1.0/account/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/storage/container/ctrn-0/",
		"blcc://eth1.0/account/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/storage/native/" + string(codec.Bytes32([32]byte{1}).Encode()),
		"blcc://eth1.0/account/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/storage/native/" + string(codec.Bytes32([32]byte{2}).Encode()),
	}

	vals := []interface{}{
		univalue.NewUnivalue(0, "", 0, 0, 0, commutative.NewBoundedUint64(1, 111), nil),
		univalue.NewUnivalue(0, "", 0, 0, 0, commutative.InitNewPaths([]string{"ctrn-0"}), nil),
		univalue.NewUnivalue(0, "", 0, 0, 0, noncommutative.NewInt64(199), nil),
	}

	store.Precommit(aliceKeys, vals)
	store.Commit() // Calculate root hash

	proofs := memorydb.New()
	common.FilterFirst(storage.LoadDataStore(store.EthDB(), store.Root())).Trie().Prove(store.Hash(aliceKeys[0]), 0, proofs)
	if _, err := ethmpt.VerifyProof(store.Root(), store.Hash(aliceKeys[0]), proofs); err != nil {
		t.Error("Actual :", err)
	}

	common.FilterFirst(storage.LoadDataStore(store.EthDB(), store.Root())).Trie().Prove(store.Hash(aliceKeys[1]), 0, proofs)
	if _, err := ethmpt.VerifyProof(store.Root(), store.Hash(aliceKeys[1]), proofs); err != nil {
		t.Error("Actual :", err)
	}

	common.FilterFirst(storage.LoadDataStore(store.EthDB(), store.Root())).Trie().Prove(store.Hash(aliceKeys[2]), 0, proofs)
	if _, err := ethmpt.VerifyProof(store.Root(), store.Hash(aliceKeys[2]), proofs); err != nil {
		t.Error("Actual :", err)
	}
}

func TestEthWorldTrieProof(t *testing.T) {
	url := ccurl.NewConcurrentUrl(storage.NewParallelEthMemDataStore())
	alice := AliceAccount()
	aliceTrans, _ := url.NewAccount(0, alice)
	fmt.Print(aliceTrans)

	bob := BobAccount()
	bobTrans, _ := url.NewAccount(0, bob)
	fmt.Print(bobTrans)

	bobTrans[0].Value().(interfaces.Type).Clone()

	acctTrans := url.Export(indexer.Sorter)

	url.Import(acctTrans)
	url.Sort()
	url.Commit([]uint32{0})

	proofs := memorydb.New() // Proof DB
	store := url.Importer().Store().(*storage.EthDataStore)
	// store.Precommit()

	// Prove the world trie path
	store.Trie().Prove([]byte(*aliceTrans[0].GetPath()), 0, proofs)
	if _, err := ethmpt.VerifyProof(store.Trie().Hash(), []byte(alice), proofs); err != nil {
		t.Error("Actual :", err)
	}

	// aliceAcctFromTrie := store.GetAccountFromTrie(alice, new(ethmpt.AccessListCache))

	// if len(aliceAcctData.([]byte)) == 0 {
	// 	t.Error("Error:")
	// }

	// Get Alice's account
	aliceAcct, _ := store.GetAccountFromTrie(alice, &ethmpt.AccessListCache{})
	fmt.Print(aliceAcct)

	// Prove the storage value
	// proofs = memorydb.New() // Proof DB
	// aliceAcct.Trie().Prove([]byte(*aliceTrans[0].GetPath()), 0, proofs)
	// if _, err := ethmpt.VerifyProof(aliceAcct.StateAccount.Root, []byte(*aliceTrans[0].GetPath()), proofs); err != nil {
	// 	t.Error("Actual :", err)
	// }
}

// need to hash the keys first
func TestEthStorageConnection(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	trans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import(trans)
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	v, err := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path))
	if v == nil {
		t.Error(err)
	}
}

func TestBasicAddRead(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// create a path
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewString("path")); err == nil {
		t.Error(err)
	}

	// Try to read an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1", nil); value != nil {
		t.Error("Error: Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-000", nil); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// try again
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// try to read an nonexistent path
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil); value != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	// Write the entry
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	// Write the entry
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-111", noncommutative.NewInt64(9999)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-111")
	}

	// if v, _ := url.Find(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewInt64(1111)); v != nil {
	// 	t.Error("Error: The path should have been deleted")
	// }

	// Read the entry back
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value.(int64) != 1111 {
		t.Error("Error: Wrong value")
	}

	// Read the path
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error(value)
	} else {
		target := value.(*orderedset.OrderedSet).Keys()
		if !reflect.DeepEqual(target, []string{"elem-000", "elem-111"}) {
			t.Error("Error: Wrong value !!!!")
		}
	}
}

func TestEthDataStoreAddDeleteRead(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	// acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	// url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))

	// url.Sort()
	// url.Commit([]uint32{ccurlcommon.SYSTEM})

	// url.Init(store)
	// create a path

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewString("path")); err == nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err)
	}

	// Write an entry having the the same name of a path, should go through
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil {
		t.Error(err)
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); value != nil {
		t.Error("not found")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value != nil {
		t.Error("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000 not found")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", new(noncommutative.Int64)); value != nil {
		t.Error("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-001 not found")
	}

	// Write an entry having the the same name of a path, should go through
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-888", noncommutative.NewInt64(888)); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-999", noncommutative.NewInt64(999)); err != nil {
		t.Error(err)
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-888", new(noncommutative.Int64)); value == nil {
		t.Error("not found")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-999", new(noncommutative.Int64)); value == nil {
		t.Error("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000 not found")
	}

	meta, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := meta.(*orderedset.OrderedSet).Keys()
	if meta == nil || len(keys) != 2 ||
		keys[0] != "elem-888" ||
		keys[1] != "elem-999" {
		t.Error("not found")
	}
}

func TestAddThenDeletePathInEthTrie(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans := url.Export(indexer.Sorter)
	trans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	acctTrans := (&indexer.Univalues{}).Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues)

	//values := indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).([]interfaces.Univalue)
	ts := indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues)
	url.Import(ts)
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	transitions := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import((&indexer.Univalues{}).Decode(indexer.Univalues(transitions).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{1})

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{})
	if v == nil {
		t.Error("Error: The path should exists")
	}

	url.Init(store)
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", nil); err != nil { // Delete the path
		t.Error(err)
	}

	trans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import((&indexer.Univalues{}).Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path)); v != nil {
		t.Error("Error: The path should have been deleted")
	}
}

func BenchmarkMultipleAccountCommitDataStore(b *testing.B) {
	// store := chooseDataStore() // Eth data store
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode) // Native data store

	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath() // create a path
	if _, err := url.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	// t0 := time.Now()
	for i := 0; i < 100000; i++ {
		acct := fmt.Sprint(rand.Int())
		if _, err := url.NewAccount(ccurlcommon.SYSTEM, acct); err != nil { // NewAccount account structure {
			fmt.Println(err)
		}

		path := commutative.NewPath() // create a path
		if _, err := url.Write(0, "blcc://eth1.0/account/"+acct+"/storage/ctrn-0/", path); err != nil {
			b.Error(err)
		}

		for j := 0; j < 4; j++ {
			if _, err := url.Write(0, "blcc://eth1.0/account/"+acct+"/storage/ctrn-0/elem-0"+fmt.Sprint(j), noncommutative.NewString("fmt.Sprint(i)")); err != nil { /* The first Element */
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
	common.Fill(diskdbs[:], leveldb)
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
	common.Fill(diskdbs[:], leveldb)
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
	common.ParallelWorker(len(keys), 8, func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			trie.Get(keys[i])
		}
	})
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
