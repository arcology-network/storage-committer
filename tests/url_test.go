package ccurltest

import (
	"fmt"
	"reflect"
	"testing"

	datastore "github.com/arcology-network/common-lib/cachedstorage/datastore"
	"github.com/arcology-network/common-lib/cachedstorage/memdb"
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
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

func TestSize(t *testing.T) {
	// store := chooseDataStore()
	store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ele0", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); v == nil {
		t.Error("Error: The path should exists")
	}

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ele0", new(noncommutative.String))
	if v == nil || v.(string) != "124" {
		t.Error("Error: The path should exists")
	}

	raw := url.WriteCache().Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.IPCTransition{})

	indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode())
	url.Import(acctTrans)
}

func TestReadWriteAt(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele0", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele1", noncommutative.NewString("1111")); err != nil {
		t.Error(err)
	}

	_0, _, _ := url.ReadAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 0, new(noncommutative.String))
	if !reflect.DeepEqual(_0, "124") {
		t.Error("Error: Should be empty!!")
	}

	_1, _, _ := url.ReadAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 1, new(noncommutative.String))
	if !reflect.DeepEqual(_1, "1111") {
		t.Error("Error: Should be empty!!")
	}

	url.WriteAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 0, noncommutative.NewString("456"))

	_0, _, _ = url.ReadAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 0, new(noncommutative.String))
	if !reflect.DeepEqual(_0, "456") {
		t.Error("Error: Should be empty!!")
	}

	url.WriteAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 0, nil) // Delete the first one

	_0, _, _ = url.ReadAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 1, new(noncommutative.String))
	if !reflect.DeepEqual(_0, "1111") {
		t.Error("Error: Should be empty!!")
	}
}

func TestAddThenDeletePath(t *testing.T) {
	store := chooseDataStore()
	alice := AliceAccount()

	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, acctTrans := url.WriteCache().Export(indexer.Sorter)

	// acctTrans := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})

	// buffer := indexer.Univalues(acctTrans).Encode()
	// out := indexer.Univalues{}.Decode(buffer).(indexer.Univalues)

	// url.Import(out)
	// url.Sort()
	// url.Commit([]uint32{ccurlcommon.SYSTEM})

	// url = ccurl.NewConcurrentUrl(store)
	// create a path
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// transitions := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	// url.Import(indexer.Univalues{}.Decode(indexer.Univalues(transitions).Encode()).(indexer.Univalues))
	// url.Sort()
	// url.Commit([]uint32{1})

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if v == nil {
		t.Error("Error: The path should exists")
	}

	url.Init(store)
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil { // Delete the path
		t.Error(err)
	}

	// acctTrans = indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	// buffer = indexer.Univalues(acctTrans).Encode()
	// url.Import(indexer.Univalues{}.Decode(buffer).(indexer.Univalues))
	// url.Sort()
	// url.Commit([]uint32{1})

	// if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
	// 	t.Error("Error: The path should have been deleted")
	// }
}

func TestAddThenDeletePath2(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans := url.WriteCache().Export(indexer.Sorter)
	trans := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	acctTrans := (&indexer.Univalues{}).Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues)

	//values := indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).([]interfaces.Univalue)
	ts := indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues)
	url.Import(ts)
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	transitions := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import((&indexer.Univalues{}).Decode(indexer.Univalues(transitions).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{1})

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if v == nil {
		t.Error("Error: The path should exists")
	}

	url.Init(store)
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil { // Delete the path
		t.Error(err)
	}

	trans = indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import((&indexer.Univalues{}).Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); v != nil {
		t.Error("Error: The path should have been deleted")
	}
}

func TestBasic(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})
	// indexer.Univalues(acctTrans).Print()

	url.Init(store)
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

	trans := common.Clone(url.WriteCache().Export(indexer.Sorter))
	transitions := indexer.Univalues(trans).To(indexer.ITCTransition{})

	// indexer.Univalues(transitions).Print()

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

	bob := BobAccount()
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+bob+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value != nil {
		t.Error("Error: Wrong value")
	}

}

func TestPathAddThenDelete(t *testing.T) {
	store := chooseDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

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

func TestUrl1(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	raw := url.WriteCache().Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.ITCTransition{})
	// accesses := indexer.Univalues(common.Clone(this.buffer)).To(indexer.ITCAccess{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	// url.Init(store)
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	// Write an entry having the the same name of a path, should go through
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0", noncommutative.NewString("ctrn-0")); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("elem-0")); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(5555)); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(6666)); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(7777)); err != nil {
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
	transitions := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})
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
	store := chooseDataStore()
	// store := datastore.NewDataStore(nil, nil, nil, encoder, decoder)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// Create a new container
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err, "Error:  Failed to MakePath: "+"/ctrn-0/")
	}

	// Add a vaiable directly
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("0000")); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/elem-0")
	}

	// Add the first element
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error(err, "Error: Failed to Write: "+"/ctrn-0/elem-000")
	}

	// Add the second element
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-001")
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-002")
	}

	// Write to an nonexistent path, will fail, but leave a couple of access records
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-002", noncommutative.NewInt64(3333)); err == nil {
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
	keys := meta0.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Do again
	meta1, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys = meta1.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Two queries should match
	if !reflect.DeepEqual(meta0, meta1) {
		t.Error("Error: Wrong meta info")
	}

	/////////////////=====================================================================/////////////////////////////
	// _0, _, _ := url.ReadAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", 0, new(noncommutative.Int64))
	// if !reflect.DeepEqual(_0, 1111) {
	// 	t.Error("Error: Should be empty!!")
	// }

	// _1, _, _ := url.ReadAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", 1, new(noncommutative.Int64))
	// if !reflect.DeepEqual(_1, 2222) {
	// 	t.Error("Error: Should be empty!!")
	// }

	// _2, _, _ := url.ReadAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", 2, new(noncommutative.Int64))
	// if !reflect.DeepEqual(_2, 3333) {
	// 	t.Error("Error: Should be empty!!")
	// }

	// url.WriteAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", 0, new(noncommutative.String))
	// if !reflect.DeepEqual(_0, "elem-000") {
	// 	t.Error("Error: Should be empty!!")
	// }

	// url.WriteAt(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", 1, new(noncommutative.String))
	// if !reflect.DeepEqual(_1, "elem-001") {
	// 	t.Error("Error: Should be empty!!")
	// }
	/////////////////=====================================================================/////////////////////////////

	// Delete elem-00
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil); err != nil {
		t.Error("Error: Failed to delete: " + "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil); err == nil {
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
	keys = meta1.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	// Readd elem-00 back
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(9999)); err != nil { // delete
		t.Error("Error: Failed to write: " + "/ctrn-0/elem-000")
	}

	// Check elem-00's value
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value.(int64) != 9999 {
		t.Error("Error: The element wasn't successfully deleted")
	}

	// Update then read the path info again
	meta, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys = meta.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-001", "elem-002", "elem-000"}) {
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
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil { // Delete the path and its sub paths
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
	keys = v.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{}) {
		t.Error("Error: Should be empty!!")
	}

	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{})
	keys = v.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{}) {
		t.Error("Error: Should be empty!!")
	}

	/*  Export all */
	// accessRecords, transitions := url.WriteCache().Export(indexer.Sorter)
	accessRecords := indexer.Univalues(common.Clone(url.WriteCache().Export())).To(indexer.ITCAccess{})
	transitions := indexer.Univalues(common.Clone(url.WriteCache().Export())).To(indexer.ITCTransition{})

	// 3 writes + 1 affiliated write
	value := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, nil, nil)
	if !indexer.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", 1, 1, 0, nil, nil)
	if !indexer.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", 0, 1, 0, nil, nil)
	if !indexer.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-005", 1, 0, 0, nil, nil)
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

func TestTransientDBv2(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := url.WriteCache().Export(indexer.Sorter)
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

func TestCustomCodec(t *testing.T) {
	// fileDB, err := datastore.NewFileDB(ROOT_PATH, 8, 2)
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }

	policy := datastore.NewCachePolicy(0, 1)
	store := datastore.NewDataStore(nil, policy, memdb.NewMemDB(), storage.Rlp{}.Encode, storage.Rlp{}.Decode)
	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := url.WriteCache().Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.IPCTransition{})

	buffer := indexer.Univalues(acctTrans).Encode()
	indexer.Univalues{}.Decode(buffer)
	url.Import(acctTrans)

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)

	// commutative.NewU256Delta(100)
	// url.Write(1, "blcc://eth1.0/account/"+alice+"/balance", &commutative.U256{value: 100})

	value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/balance", &commutative.U256{})
	valueAdd := value.(uint256.Int)
	if value == nil || (&valueAdd).ToBig().Uint64() != 0 {
		t.Error("Error: Wrong value", value.(*uint256.Int).ToBig().Uint64())
	}
}
