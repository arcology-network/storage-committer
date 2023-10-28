package ccurltest

import (
	"fmt"
	"reflect"
	"testing"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/concurrenturl/noncommutative"
	storage "github.com/arcology-network/concurrenturl/storage"
	"github.com/arcology-network/concurrenturl/univalue"
	"github.com/holiman/uint256"
)

const (
	ROOT_PATH   = "./tmp/filedb/"
	BACKUP_PATH = "./tmp/filedb-back/"
)

var (
	// encoder = storage.Codec{}.Encode
	// decoder = storage.Codec{}.Decode

	encoder = storage.Rlp{}.Encode
	decoder = storage.Rlp{}.Decode
)

func TestSize(t *testing.T) {
	fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	if err != nil {
		t.Error(err)
		return
	}

	store := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), fileDB, encoder, decoder)
	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := url.Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.IPCTransition{})

	indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode())
	url.Import(acctTrans)

	original := []int{1, 2, 3, 4}
	original = append([]int{}, (original)...)
	fmt.Println(original)
	original[0] = 99
	fmt.Println(original)
	fmt.Println(original, "!!!")
}

func TestAddThenDeletePath(t *testing.T) {
	fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	if err != nil {
		t.Error(err)
		return
	}

	// compressionLut := datacompression.NewCompressionLut()
	store := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), fileDB, encoder, decoder)
	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, acctTrans := url.Export(indexer.Sorter)

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})

	buffer := indexer.Univalues(acctTrans).Encode()
	out := indexer.Univalues{}.Decode(buffer).(indexer.Univalues)

	url.Import(out)
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true); err != nil {
		t.Error(err)
	}

	transitions := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(transitions).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if v == nil {
		t.Error("Error: The path should exists")
	}

	url.Init(store)
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil, true); err != nil { // Delete the path
		t.Error(err)
	}

	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	buffer = indexer.Univalues(acctTrans).Encode()
	url.Import(indexer.Univalues{}.Decode(buffer).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
		t.Error("Error: The path should have been deleted")
	}
}

func TestAddThenDeletePath2(t *testing.T) {
	fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	if err != nil {
		t.Error(err)
		return
	}

	// compressionLut := datacompression.NewCompressionLut()
	store := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), fileDB, encoder, decoder)

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
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
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true); err != nil {
		t.Error(err)
	}

	transitions := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import((&indexer.Univalues{}).Decode(indexer.Univalues(transitions).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{1})

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if v == nil {
		t.Error("Error: The path should exists")
	}

	url.Init(store)
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil, true); err != nil { // Delete the path
		t.Error(err)
	}

	trans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import((&indexer.Univalues{}).Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); v != nil {
		t.Error("Error: The path should have been deleted")
	}
}

func TestBasic(t *testing.T) {
	fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	if err != nil {
		t.Error(err)
		return
	}

	// compressionLut := datacompression.NewCompressionLut()
	store := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(1000000, 1), fileDB, encoder, decoder)

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewString("path"), true); err == nil {
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
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111), true); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	// Write the entry
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-111", noncommutative.NewInt64(9999), true); err != nil {
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
		target := value.([]string)
		if !reflect.DeepEqual(target, []string{"elem-000", "elem-111"}) {
			t.Error("Error: Wrong value !!!!")
		}
	}

	trans := common.Clone(url.Export(indexer.Sorter))
	transitions := indexer.Univalues(trans).To(indexer.ITCTransition{})

	if !reflect.DeepEqual(transitions[0].Value().(interfaces.Type).Delta().(*commutative.PathDelta).Added(), []string{"elem-000", "elem-111"}) {
		t.Error("Error: keys are missing from the added buffer!")
	}

	value := transitions[1].Value()
	if *value.(*noncommutative.Int64) != 1111 {
		t.Error("Error: keys don't match")
	}

	// wrong condition, value should still exists
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); value == nil {
		t.Error("Error: The variable has been cleared !")
	}

	// data := indexer.Univalues(transitions).To(indexer.IPCTransition{})

	buffer := indexer.Univalues(indexer.Univalues(transitions).To(indexer.IPCTransition{})).Encode()
	url.Import(indexer.Univalues{}.Decode(buffer).(indexer.Univalues))
	// url.Import(url.Decode(indexer.Univalues(transitions).Encode()))
	url.Sort()
	url.Commit([]uint32{1})

	/* =========== The second cycle ==============*/
	//try reading an element written in the previous cycle
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil {
		t.Error("Error: Entry not found")
	}
}

func TestPathAddThenDelete(t *testing.T) {
	fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	if err != nil {
		t.Error(err)
		return
	}
	store := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), fileDB, encoder, decoder)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	// url.Init(store)
	// create a path

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath(), true); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewString("path"), true); err == nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222), true); err != nil {
		t.Error(err)
	}

	// Write an entry having the the same name of a path, should go through
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil, true); err != nil {
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
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath(), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-888", noncommutative.NewInt64(888), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-999", noncommutative.NewInt64(999), true); err != nil {
		t.Error(err)
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-888", new(noncommutative.Int64)); value == nil {
		t.Error("not found")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-999", new(noncommutative.Int64)); value == nil {
		t.Error("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000 not found")
	}

	meta, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if meta == nil || len(meta.([]string)) != 2 ||
		meta.([]string)[0] != "elem-888" ||
		meta.([]string)[1] != "elem-999" {
		t.Error("not found")
	}
}

func TestUrl1(t *testing.T) {
	fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	if err != nil {
		t.Error(err)
		return
	}
	store := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), fileDB, encoder, decoder)

	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	raw := url.Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.ITCTransition{})
	// accesses := indexer.Univalues(common.Clone(this.buffer)).To(indexer.ITCAccess{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	// url.Init(store)
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath(), true); err != nil {
		t.Error(err)
	}

	// Write an entry having the the same name of a path, should go through
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0", noncommutative.NewString("ctrn-0"), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("elem-0"), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(5555), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(6666), true); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(7777), true); err != nil {
		t.Error(err)
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil || value.(int64) != 5555 {
		t.Error("Error: Wrong value")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", new(noncommutative.Int64)); value == nil || value.(int64) != 6666 {
		t.Error("Error: Wrong value")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", new(noncommutative.Int64)); value == nil || value.(int64) != 7777 {
		t.Error("Error: Wrong value")
	}

	if meta, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); meta == nil {
		t.Error("Error: not found")
	}

	// Export all access records and state transitions
	transitions := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	// v, _, _ := transitions[0].Value().(interfaces.Type).Get()
	if (*transitions[0].Value().(*noncommutative.String)) != "ctrn-0" {
		t.Error("Error: keys don't match")
	}

	addedkeys := codec.Strings(transitions[1].Value().(interfaces.Type).Delta().(*commutative.PathDelta).Added()).Sort()
	if !reflect.DeepEqual([]string(addedkeys), []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	if meta, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{}); meta == nil {
		t.Error("Error: The variable has been cleared")
	}
}

func TestUrl2(t *testing.T) {
	fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	if err != nil {
		t.Error(err)
		return
	}
	store := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), fileDB, encoder, decoder)
	// store := cachedstorage.NewDataStore(nil, nil, nil, encoder, decoder)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// Create a new container
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true); err != nil {
		t.Error(err, "Error:  Failed to MakePath: "+"/ctrn-0/")
	}

	// Add a vaiable directly
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("0000"), true); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/elem-0")
	}

	// Add the first element
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111), true); err != nil {
		t.Error(err, "Error: Failed to Write: "+"/ctrn-0/elem-000")
	}

	// Add the second element
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222), true); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-001")
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333), true); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-002")
	}

	// Write to an nonexistent path, will fail, but leave a couple of access records
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-002", noncommutative.NewInt64(3333), true); err == nil {
		t.Error(err, "Error:    /ctrn-1/ does not exist, the Write should fail!!")
	}

	// Read an nonexistent path, shouldn't succeed
	if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-002", new(noncommutative.Int64)); v != nil {
		t.Error("Error:  /ctrn-1/ does not exist, the read should fail!!")
	}

	// Add the first element
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil || value.(int64) != 1111 {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-000")
	}

	// Try to read an nonexistent element, should leave a access record
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-005", nil); value != nil {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-005")
	}

	// Update then return path meta info
	meta0, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if !reflect.DeepEqual(meta0.([]string), []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Do again
	meta1, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if !reflect.DeepEqual(meta1.([]string), []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Two queries should match
	if !reflect.DeepEqual(meta0, meta1) {
		t.Error("Error: Wrong meta info")
	}

	// Delete elem-00
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil, true); err != nil {
		t.Error("Error: Failed to delete: " + "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil, true); err == nil {
		t.Error("Error: Failed to delete: " + "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value != nil {
		t.Error("Error: The element wasn't successfully deleted")
	}

	// Check elem-00's value
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", new(noncommutative.Int64)); value.(int64) != 2222 {
		t.Error("Error: The element wasn't found")
	}

	// The elem-00 has been deleted, only "elem-001", "elem-002" left
	meta, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if !reflect.DeepEqual(meta.([]string), []string{"elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	// Readd elem-00 back
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(9999), true); err != nil { // delete
		t.Error("Error: Failed to write: " + "/ctrn-0/elem-000")
	}

	// Check elem-00's value
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value.(int64) != 9999 {
		t.Error("Error: The element wasn't successfully deleted")
	}

	// Update then read the path info again
	meta, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if !reflect.DeepEqual(meta.([]string), []string{"elem-001", "elem-002", "elem-000"}) {
		t.Error("Error: keys don't match")
	}

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", new(noncommutative.Int64))
	if v == nil {
		t.Error("Error: keys don't match")
	}

	// if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); (*value.(*noncommutative.Int64)) != 9999 {
	// 	t.Error("Error: The element wasn't successfully deleted")
	// }

	/* Remove the path and all the elements underneath */
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil, true); err != nil { // Delete the path and its sub paths
		t.Error(err, "Failed to remove path: "+"/ctrn-0/")
	}

	if v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil { /* The path should be gone by now */
		t.Error("Error: The key should not exist!")
	}

	if v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", new(noncommutative.Int64)); v != nil { /* all the sub paths should be gone by now*/
		t.Error("Error: The key should not exist!")
	}

	/*  Read the storage path to see what is left*/
	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{})
	if !reflect.DeepEqual(v.([]string), []string{}) {
		t.Error("Error: Should be empty!!")
	}

	/*  Export all */
	// accessRecords, transitions := url.Export(indexer.Sorter)
	accessRecords := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCAccess{})
	transitions := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCTransition{})

	// 3 writes + 1 affiliated write
	value := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, nil)
	if !indexer.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", 1, 1, 0, nil)
	if !indexer.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", 0, 1, 0, nil)
	if !indexer.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-005", 1, 0, 0, nil)
	if !indexer.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	// Encode then Decode access records
	buffer := indexer.Univalues(transitions).Encode()
	out := indexer.Univalues{}.Decode(buffer).(indexer.Univalues)

	for i := range transitions {
		if !transitions[i].(*univalue.Univalue).Equal(out[i].(*univalue.Univalue)) {
			t.Error("Error: transitions don't match")
		}
	}
}

// // func TestUnivaluesBatchCodec(t *testing.T) {
// // 	store := cachedstorage.NewDataStore(nil, nil, nil, encoder, decoder)
// // 	url := ccurl.NewConcurrentUrl(store)
// // 	alice := AliceAccount()
// // 	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
// // 		t.Error(err)
// // 	}

// // 	// path := commutative.NewPath()
// // 	// if err != nil {
// // 	// 	t.Error("Error: Failed to create the path")
// // 	// }
// // 	// if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
// // 	// 	t.Error(err, " Failed to MakePath: "+"/ctrn-0/")
// // 	// }

// // 	// inV := noncommutative.NewBigint(123456)
// // 	// if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", inV); err != nil {
// // 	// 	t.Error(err, " Failed to Write: "+"/elem-0")
// // 	// }

// // 	// v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0")
// // 	// value := (*big.Int)(inV.(*noncommutative.Bigint))
// // 	// outV := (*big.Int)(v.(*noncommutative.Bigint))
// // 	// if outV.Cmp(value) != 0 {
// // 	// 	t.Error("Error: Bigint values don't match")
// // 	// }

// // 	// accessRecords, _ := url.Export(indexer.Sorter)
// // 	accessRecords := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCAccess{})

// // 	in := indexer.Univalues(accessRecords).Encode()

// // 	// uint256delta isn't inthe encoder !!!

// // 	fmt.Println(len(in))
// // 	out := indexer.Univalues{}.Decode(in).(indexer.Univalues)
// // 	for i := range accessRecords {
// // 		if !accessRecords[i].(*univalue.Univalue).Equal(out[i].(*univalue.Univalue)) {
// // 			t.Error("Error: Accesses don't match")
// // 		}
// // 	}
// // }

// func TestCommutative(t *testing.T) {
// 	store := cachedstorage.NewDataStore(nil, nil, nil, encoder, decoder)
// 	url := ccurl.NewConcurrentUrl(store)
// 	alice := AliceAccount()
// 	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	// create a path
// 	path := commutative.NewPath()
// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true); err != nil {
// 		t.Error(err, " Failed to MakePath: blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
// 	}

// 	// create a noncommutative bigint
// 	inV := noncommutative.NewBigint(100)
// 	value := (*big.Int)(inV.(*noncommutative.Bigint))
// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", inV, true); err != nil {
// 		t.Error(err, " Failed to Write: blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0")
// 	}

// 	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0")
// 	outV := (*big.Int)(v.(*big.Int))
// 	if outV.Cmp(value) != 0 {
// 		t.Error("Error: Failed to read: blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-0")
// 	}

// 	// -------------------Create a commutative UINT256 ------------------------------
// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewBoundedU256(commutative.U256_MIN, commutative.U256_MAX), true); err != nil { // 0
// 		t.Error(err, " Failed to Write: "+"/elem-0")
// 	}

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(300), true), true); err != nil { // 300
// 		t.Error(err, " Failed to Write: "+"/elem-0")
// 	}

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewBoundedU256(commutative.U256_MIN, commutative.U256_MAX), true); err != nil { // still 300
// 		t.Error(err, " Failed to Write: "+"/elem-0")
// 	}

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(300), true), true); err != nil { //  600
// 		t.Error(err, " Failed to Write: "+"/elem-0")
// 	}

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewBoundedU256(commutative.U256_MIN, commutative.U256_MAX), true); err != nil { // still 300
// 		t.Error(err, " Failed to Write: "+"/elem-0")
// 	}

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(300), false), true); err != nil { // 600 - 300 = 300
// 		t.Error(err, " Failed to Write: "+"/elem-0")
// 	}

// 	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0")
// 	if v.(*uint256.Int).Cmp(uint256.NewInt(300)) != 0 {
// 		t.Error("Error: comt-0 has a wrong returned value")
// 	}

// 	// ----------------------------U256 ---------------------------------------------------
// 	if _, err := url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/balance",
// 		commutative.NewBoundedU256(), true); err != nil { //initialization
// 		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
// 	}

// 	// Add the first delta
// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(22), true), true); err != nil {
// 		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
// 	}

// 	// Add the second delta
// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(11), true), true); err != nil {
// 		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
// 	}

// 	// Read alice's balance
// 	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/balance")
// 	if v.(*uint256.Int).Cmp(uint256.NewInt(33)) != 0 {
// 		t.Error("Error: blcc://eth1.0/account/alice/balance")
// 	}

// 	// Export variables
// 	// accessRecords, _ := url.Export(indexer.Sorter)
// 	// bf := accessRecords[6].Encode()
// 	// bfout := (&univalue.Univalue{}).Decode(bf).(*univalue.Univalue)
// 	// if accessRecords[6].Equal(bfout) {
// 	// 	t.Error("Error: Accesses don't match")
// 	// }

// 	// in := indexer.Univalues(accessRecords).Encode()
// 	// out := indexer.Univalues{}.Decode(in).(indexer.Univalues)
// 	// for i := range accessRecords {
// 	// 	if !accessRecords[i].(*univalue.Univalue).Equal(out[i].(*univalue.Univalue)) {
// 	// 		t.Error("Error: Accesses don't match")
// 	// 	}
// 	// }

// 	// in = indexer.Univalues(transitions).Encode()
// 	// out = indexer.Univalues{}.Decode(in).(indexer.Univalues)
// 	// for i := range transitions {
// 	// 	if !transitions[i].(*univalue.Univalue).Equal(out[i].(*univalue.Univalue)) {
// 	// 		t.Error("Error: Transitions don't match !!!")
// 	// 		fmt.Println(transitions[i])
// 	// 		fmt.Println(out[i])
// 	// 	}
// 	// }

// 	// url.Import(out)
// 	// url.Commit([]uint32{0, 1})
// }

// func TestNestedPath(t *testing.T) {
// 	store := cachedstorage.NewDataStore(nil, nil, nil, encoder, decoder)
// 	url := ccurl.NewConcurrentUrl(store)
// 	alice := AliceAccount()
// 	if err := url.NewAccount(ccurlcommon.SYSTEM, AliceAccount()); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	if err := url.NewAccount(ccurlcommon.SYSTEM, BobAccount()); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	// create a path
// 	path := commutative.NewPath()

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true); err != nil {
// 		t.Error(err)
// 	}

// 	// create a sub path
// 	path = commutative.NewPath()

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ctrn-00/", path, true); err != nil {
// 		t.Error(err)
// 	}

// 	// Try to rewrite a path, should fail !
// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-00", noncommutative.NewString("elem-00"), true); err != nil {
// 		t.Error(err)
// 	}

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-01", noncommutative.NewInt64(1234), true); err != nil {
// 		t.Error(err)
// 	}

// 	// The first element !
// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ctrn-00/elem-00", noncommutative.NewString("elem-00"), true); err != nil {
// 		t.Error(err)
// 	}

// 	// The second element
// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ctrn-00/elem-01", noncommutative.NewString("elem-01"), true); err != nil {
// 		t.Error(err)
// 	}

// 	/* Read */
// 	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ctrn-00/")
// 	if !reflect.DeepEqual(v.([]string), []string{"elem-00", "elem-01"}) {
// 		t.Error("Error: keys don't match")
// 	}

// 	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
// 	if !reflect.DeepEqual(v.([]string), []string{"ctrn-00/", "elem-00", "elem-01"}) {
// 		t.Error("Error: keys don't match")
// 	}

// 	/* Remove the path */
// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil, true); err != nil {
// 		t.Error(err)
// 	}

// 	/*Try reading again */
// 	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
// 	if v != nil {
// 		t.Error("Error: Error:  Path should been deleted already !")
// 	}

// 	// accessRecords, transitions := url.Export(indexer.Sorter)
// 	accessRecords := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
// 	transitions := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

// 	in := indexer.Univalues(accessRecords).Encode()
// 	out := indexer.Univalues{}.Decode(in).(indexer.Univalues)
// 	for i := range accessRecords {
// 		if !accessRecords[i].(*univalue.Univalue).Equal(out[i].(*univalue.Univalue)) {
// 			t.Error("Error: Accesses don't match  before and after encoding")
// 		}
// 	}

// 	in = indexer.Univalues(transitions).Encode()
// 	out = indexer.Univalues{}.Decode(in).(indexer.Univalues)
// 	// for i := range transitions {
// 	// 	if !transitions[i].(*univalue.Univalue).Equal(out[i].(*univalue.Univalue)) {
// 	// 		t.Error("Error: Transitions don't match before and after encoding")
// 	// 	}
// 	// }

// 	url.Import(out)
// 	url.Sort()
// 	url.Commit([]uint32{1})
// }

func TestCustomCodec(t *testing.T) {
	fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	if err != nil {
		t.Error(err)
		return
	}

	policy := cachedstorage.NewCachePolicy(0, 1)
	store := cachedstorage.NewDataStore(nil, policy, fileDB, storage.Rlp{}.Encode, storage.Rlp{}.Decode)
	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := url.Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.IPCTransition{})

	buffer := indexer.Univalues(acctTrans).Encode()
	indexer.Univalues{}.Decode(buffer)
	url.Import(acctTrans)

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)

	// commutative.NewU256Delta(100, true)
	// url.Write(1, "blcc://eth1.0/account/"+alice+"/balance", &commutative.U256{value: 100})

	value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/balance", &commutative.U256{})
	if value == nil || value.(*uint256.Int).ToBig().Uint64() != 0 {
		t.Error("Error: Wrong value", value.(*uint256.Int).ToBig().Uint64())
	}
}
