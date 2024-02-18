package ccurltest

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/codec"
	orderedset "github.com/arcology-network/common-lib/container/set"
	"github.com/arcology-network/common-lib/exp/array"
	datastore "github.com/arcology-network/common-lib/storage/datastore"
	"github.com/arcology-network/common-lib/storage/memdb"
	ccurl "github.com/arcology-network/concurrenturl"
	committercommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/commutative"
	importer "github.com/arcology-network/concurrenturl/importer"
	"github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/concurrenturl/noncommutative"
	platform "github.com/arcology-network/concurrenturl/platform"
	storage "github.com/arcology-network/concurrenturl/storage"
	"github.com/arcology-network/concurrenturl/univalue"
	cache "github.com/arcology-network/eu/cache"
	"github.com/holiman/uint256"
)

func TestSize(t *testing.T) {
	// store := chooseDataStore()
	store := chooseDataStore()

	alice := AliceAccount()
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ele0", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	buffer := array.New[byte](320, 11)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ele1", noncommutative.NewBytes(buffer)); err != nil {
		t.Error(err)
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); v == nil {
		t.Error("Error: The path should exists")
	}

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ele0", new(noncommutative.String))
	if v == nil || v.(string) != "124" {
		t.Error("Error: The path should exists")
	}

	raw := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(array.Clone(raw)).To(importer.IPTransition{})

	committer.Import(acctTrans)
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()

	committer.Init(store)
	writeCache.Clear()

	outV, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ele1", new(noncommutative.Bytes))
	if !bytes.Equal(outV.([]byte), array.New[byte](320, 11)) {
		t.Error("Error: The path should exists")
	}
}

func TestReadWriteAt(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	// committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele0", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele1", noncommutative.NewString("1111")); err != nil {
		t.Error(err)
	}

	_0, _, _ := writeCache.ReadAt(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 0, new(noncommutative.String))
	if !reflect.DeepEqual(_0, "124") {
		t.Error("Error: Should be empty!!")
	}

	_1, _, _ := writeCache.ReadAt(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 1, new(noncommutative.String))
	if !reflect.DeepEqual(_1, "1111") {
		t.Error("Error: Should be empty!!")
	}

	writeCache.WriteAt(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 0, noncommutative.NewString("456"))

	_0, _, _ = writeCache.ReadAt(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 0, new(noncommutative.String))
	if !reflect.DeepEqual(_0, "456") {
		t.Error("Error: Should be empty!!")
	}

	writeCache.WriteAt(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 0, nil) // Delete the first one

	_0, _, _ = writeCache.ReadAt(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", 1, new(noncommutative.String))
	if !reflect.DeepEqual(_0, "1111") {
		t.Error("Error: Should be empty!!")
	}
}

func TestAddThenDeletePath(t *testing.T) {
	store := chooseDataStore()
	alice := AliceAccount()

	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())

	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
	// 	t.Error(err)
	// }

	// _, acctTrans := writeCache.Export(importer.Sorter)

	// acctTrans := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	// buffer := univalue.Univalues(acctTrans).Encode()
	// out := univalue.Univalues{}.Decode(buffer).(univalue.Univalues)

	// committer.Import(out)
	// committer.Sort()
	// committer.Precommit([]uint32{committercommon.SYSTEM})
	committer.Commit()

	// committer = ccurl.NewStorageCommitter(store)
	// create a path
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// transitions := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	// committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(transitions).Encode()).(univalue.Univalues))
	// committer.Sort()
	// committer.Precommit([]uint32{1})
	committer.Commit()

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if v == nil {
		t.Error("Error: The path should exists")
	}

	committer.Init(store)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil { // Delete the path
		t.Error(err)
	}

	// acctTrans = univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	// buffer = univalue.Univalues(acctTrans).Encode()
	// committer.Import(univalue.Univalues{}.Decode(buffer).(univalue.Univalues))
	// committer.Sort()
	// committer.Precommit([]uint32{1})
	committer.Commit()

	// if v, _ := committer.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
	// 	t.Error("Error: The path should have been deleted")
	// }
}

func TestAddThenDeletePath2(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()

	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	// _, trans := writeCache.Export(importer.Sorter)
	trans := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	acctTrans := (&univalue.Univalues{}).Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues)

	//values := univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).([]*univalue.Univalue)
	ts := univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues)

	committer := ccurl.NewStorageCommitter(store)
	committer.Import(ts)
	committer.Sort()
	committer.Precommit([]uint32{committercommon.SYSTEM})
	committer.Commit()

	committer.Init(store)
	writeCache.Clear()

	// create a path
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	transitions := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer.Import((&univalue.Univalues{}).Decode(univalue.Univalues(transitions).Encode()).(univalue.Univalues))

	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()
	writeCache.Clear()

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if v == nil {
		t.Error("Error: The path should exists")
	}

	committer.Init(store)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil { // Delete the path
		t.Error(err)
	}

	trans = univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer.Import((&univalue.Univalues{}).Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()
	committer.Init(store)

	writeCache.Clear()
	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); v != nil {
		t.Error("Error: The path should have been deleted")
	}
}

func TestBasic(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{committercommon.SYSTEM})
	committer.Commit()
	// univalue.Univalues(acctTrans).Print()

	committer.Init(store)
	writeCache.Clear()
	// create a path
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewString("path")); err == nil {
		t.Error(err)
	}

	// Write a long string
	str := string(array.New[byte](320, 11))
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewString(str)); err == nil {
		t.Error(err)
	}

	// Try to read an NONEXISTENT path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1", nil); value != nil {
		t.Error("Error: Path shouldn't be not found")
	}

	// Try to read an NONEXIST nonexistent entry from an nonexistent path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-000", nil); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// try again
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// try to read an nonexistent path
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil); value != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	// Write the entry
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	// Write the entry
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-111", noncommutative.NewInt64(9999)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-111")
	}

	// if v, _ := committer.Find(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewInt64(1111)); v != nil {
	// 	t.Error("Error: The path should have been deleted")
	// }

	// Read the entry back
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value.(int64) != 1111 {
		t.Error("Error: Wrong value")
	}

	// Read the path
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error(value)
	} else {
		target := value.(*orderedset.OrderedSet).Keys()
		if !reflect.DeepEqual(target, []string{"elem-000", "elem-111"}) {
			t.Error("Error: Wrong value !!!!")
		}
	}

	trans := array.Clone(writeCache.Export(importer.Sorter))
	transitions := univalue.Univalues(trans).To(importer.ITTransition{})

	// failed to filter a read on noneexist element

	// univalue.Univalues(transitions).Print()

	if !reflect.DeepEqual(transitions[0].Value().(interfaces.Type).Delta().(*commutative.PathDelta).Added(), []string{"elem-000", "elem-111"}) {
		t.Error("Error: keys are missing from the added buffer!")
	}

	value := transitions[1].Value()
	if *value.(*noncommutative.Int64) != 1111 {
		t.Error("Error: keys don't match")
	}

	// wrong condition, value should still exists
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); value == nil {
		t.Error("Error: The variable has been cleared !")
	}

	// data := univalue.Univalues(transitions).To(importer.IPTransition{})

	buffer := univalue.Univalues(univalue.Univalues(transitions).To(importer.IPTransition{})).Encode()
	committer.Import(univalue.Univalues{}.Decode(buffer).(univalue.Univalues))
	// committer.Import(committer.Decode(univalue.Univalues(transitions).Encode()))
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()

	/* =========== The second cycle ==============*/
	//try reading an element written in the previous cycle
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil {
		t.Error("Error: Entry not found")
	}

	bob := BobAccount()
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+bob+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value != nil {
		t.Error("Error: Wrong value")
	}

}

func TestPathAddThenDelete(t *testing.T) {
	store := chooseDataStore()
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	acctTrans := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Sort()
	committer.Precommit([]uint32{committercommon.SYSTEM})
	committer.Commit()

	// committer.Init(store)
	// create a path

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewString("path")); err == nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err)
	}

	// Write an entry having the the same name of a path, should go through
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil {
		t.Error(err)
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); value != nil {
		t.Error("not found")
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value != nil {
		t.Error("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000 not found")
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", new(noncommutative.Int64)); value != nil {
		t.Error("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-001 not found")
	}

	// Write an entry having the the same name of a path, should go through
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-888", noncommutative.NewInt64(888)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-999", noncommutative.NewInt64(999)); err != nil {
		t.Error(err)
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-888", new(noncommutative.Int64)); value == nil {
		t.Error("not found")
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-999", new(noncommutative.Int64)); value == nil {
		t.Error("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000 not found")
	}

	meta, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := meta.(*orderedset.OrderedSet).Keys()
	if meta == nil || len(keys) != 2 ||
		keys[0] != "elem-888" ||
		keys[1] != "elem-999" {
		t.Error("not found")
	}
}

func TestCommitter(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	raw := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(array.Clone(raw)).To(importer.ITTransition{})
	// accesses := univalue.Univalues(array.Clone(this.buffer)).To(importer.ITAccess{})

	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Sort()
	committer.Precommit([]uint32{committercommon.SYSTEM})
	committer.Commit()

	writeCache.Clear()
	// committer.Init(store)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	// Write an entry having the the same name of a path, should go through
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0", noncommutative.NewString("ctrn-0")); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("elem-0")); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(5555)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(6666)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(7777)); err != nil {
		t.Error(err)
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil || value.(int64) != 5555 {
		t.Error("Error: Wrong value")
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", new(noncommutative.Int64)); value == nil || value.(int64) != 6666 {
		t.Error("Error: Wrong value")
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", new(noncommutative.Int64)); value == nil || value.(int64) != 7777 {
		t.Error("Error: Wrong value")
	}

	if meta, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); meta == nil {
		t.Error("Error: not found")
	}

	// Export all access records and state transitions
	transitions := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	// v, _, _ := transitions[0].Value().(interfaces.Type).Get()
	if (*transitions[0].Value().(*noncommutative.String)) != "ctrn-0" {
		t.Error("Error: keys don't match")
	}

	addedkeys := codec.Strings(transitions[1].Value().(interfaces.Type).Delta().(*commutative.PathDelta).Added()).Sort()
	if !reflect.DeepEqual([]string(addedkeys), []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	if meta, _, _ := writeCache.Read(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{}); meta == nil {
		t.Error("Error: The variable has been cleared")
	}
}

func TestCommitter2(t *testing.T) {
	store := chooseDataStore()
	// store := datastore.NewDataStore(nil, nil, nil, encoder, decoder)
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{committercommon.SYSTEM})
	committer.Commit()

	committer.Init(store)
	// Create a new container
	path := commutative.NewPath()
	writeCache.Clear()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err, "Error:  Failed to MakePath: "+"/ctrn-0/")
	}

	// Add a vaiable directly
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("0000")); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/elem-0")
	}

	// Add the first element
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error(err, "Error: Failed to Write: "+"/ctrn-0/elem-000")
	}

	// Add the second element
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-001")
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-002")
	}

	// Write to an nonexistent path, will fail, but leave a couple of access records
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-002", noncommutative.NewInt64(3333)); err == nil {
		t.Error(err, "Error:    /ctrn-1/ does not exist, the Write should fail!!")
	}

	// Read an nonexistent path, shouldn't succeed
	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-002", new(noncommutative.Int64)); v != nil {
		t.Error("Error:  /ctrn-1/ does not exist, the read should fail!!")
	}

	// Add the first element
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil || value.(int64) != 1111 {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-000")
	}

	// Try to read an nonexistent element, should leave a access record
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-005", nil); value != nil {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-005")
	}

	// Update then return path meta info
	meta0, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := meta0.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Do again
	meta1, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys = meta1.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Delete elem-00
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil); err != nil {
		t.Error("Error: Failed to delete: " + "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	// The elem-00 has been deleted, only "elem-001", "elem-002" left
	meta0, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys = meta0.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(meta0.(*orderedset.OrderedSet).Keys(), []string{"elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	// Readd elem-00 back
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(9999)); err != nil { // delete
		t.Error("Error: Failed to write: " + "/ctrn-0/elem-000")
	}

	// Check elem-00's value
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value.(int64) != 9999 {
		t.Error("Error: The element wasn't successfully deleted")
	}

	// Update then read the path info again
	meta, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys = meta.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-001", "elem-002", "elem-000"}) {
		t.Error("Error: keys don't match")
	}

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", new(noncommutative.Int64))
	if v == nil {
		t.Error("Error: keys don't match")
	}

	// if value, _ := committer.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); (*value.(*noncommutative.Int64)) != 9999 {
	// 	t.Error("Error: The element wasn't successfully deleted")
	// }

	/* Remove the path and all the elements underneath */
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil { // Delete the path and its sub paths
		t.Error(err, "Failed to remove path: "+"/ctrn-0/")
	}

	if v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil { /* The path should be gone by now */
		t.Error("Error: The key should not exist!")
	}

	if v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", new(noncommutative.Int64)); v != nil { /* all the sub paths should be gone by now*/
		t.Error("Error: The key should not exist!")
	}

	/*  Read the storage path to see what is left*/
	v, _, _ = writeCache.Read(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{})
	keys = v.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{}) {
		t.Error("Error: Should be empty!!")
	}

	v, _, _ = writeCache.Read(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{})
	keys = v.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{}) {
		t.Error("Error: Should be empty!!")
	}

	/*  Export all */
	// accessRecords, transitions := writeCache.Export(importer.Sorter)
	accessRecords := univalue.Univalues(array.Clone(writeCache.Export())).To(importer.ITAccess{})
	transitions := univalue.Univalues(array.Clone(writeCache.Export())).To(importer.ITTransition{})

	// 3 writes + 1 affiliated write
	value := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, nil, nil)
	if !univalue.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", 1, 1, 0, nil, nil)
	if !univalue.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", 0, 1, 0, nil, nil)
	if !univalue.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-005", 1, 0, 0, nil, nil)
	if !univalue.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	// Encode then Decode access records
	buffer := univalue.Univalues(transitions).Encode()
	out := univalue.Univalues{}.Decode(buffer).(univalue.Univalues)

	for i := range transitions {
		if !transitions[i].Equal(out[i]) {
			t.Error("Error: transitions don't match")
		}
	}
}

// This test is taken from the unit test above, but it is modified to use it simpler.
// It will trigger an unknown error, which is not expected. The meta1 seems to be altered by
// an delete operation, which is not expected. Not sure if it is a bug or not, certainly, this
// is not how we ususally use the system, when there is a delete operation, we should already reread
// the data in stread of using the old data object, but it is still wired.
// The test above never trigger this error before, so is either a issue in the new version of golang or a bug in the
// code.
func TestError(t *testing.T) {
	store := chooseDataStore()
	// store := datastore.NewDataStore(nil, nil, nil, encoder, decoder)
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(array.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{committercommon.SYSTEM})
	committer.Commit()

	committer.Init(store)
	// Create a new container
	path := commutative.NewPath()
	writeCache.Clear()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err, "Error:  Failed to MakePath: "+"/ctrn-0/")
	}

	// Add a vaiable directly
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("0000")); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/elem-0")
	}

	// Add the first element
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error(err, "Error: Failed to Write: "+"/ctrn-0/elem-000")
	}

	// Add the second element
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-001")
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-002")
	}

	meta1, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := meta1.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Delete elem-00
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil); err != nil {
		t.Error("Error: Failed to delete: " + "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	meta1, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	if !reflect.DeepEqual(meta1.(*orderedset.OrderedSet).Keys(), []string{"elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}
}

func TestTransientDBv2(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(array.Clone(raw)).To(importer.IPTransition{})

	univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode())
	committer.Import(acctTrans)

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
	store := datastore.NewDataStore(nil, policy, memdb.NewMemoryDB(), storage.Rlp{}.Encode, storage.Rlp{}.Decode)
	alice := AliceAccount()
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(array.Clone(raw)).To(importer.IPTransition{})

	buffer := univalue.Univalues(acctTrans).Encode()
	univalue.Univalues{}.Decode(buffer)
	committer.Import(acctTrans)

	committer.Sort()
	committer.Precommit([]uint32{committercommon.SYSTEM})
	committer.Commit()

	committer.Init(store)
	writeCache.Clear()
	// commutative.NewU256Delta(100)
	// writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", &commutative.U256{value: 100})

	value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/balance", &commutative.U256{})
	valueAdd := value.(uint256.Int)
	if value == nil || (&valueAdd).ToBig().Uint64() != 0 {
		t.Error("Error: Wrong value", value.(*uint256.Int).ToBig().Uint64())
	}
}