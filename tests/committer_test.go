package committertest

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	importer "github.com/arcology-network/storage-committer/committer"
	stgcommitter "github.com/arcology-network/storage-committer/committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/commutative"
	"github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/noncommutative"
	platform "github.com/arcology-network/storage-committer/platform"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	cache "github.com/arcology-network/storage-committer/storage/writecache"
	eucache "github.com/arcology-network/storage-committer/storage/writecache"
	"github.com/arcology-network/storage-committer/univalue"
	"github.com/holiman/uint256"
)

func CommitterCache(store interfaces.Datastore, t *testing.T) {
	alice := AliceAccount()
	committer := stgcommitter.NewStateCommitter(store)

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	committer.Import(acctTrans).Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(stgcommcommon.SYSTEM)
	committer.Clear()

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), noncommutative.NewBytes([]byte{1, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), noncommutative.NewBytes([]byte{2, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}
	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	// committer.Import(acctTrans)
	committer = stgcommitter.NewStateCommitter(store)
	committer.Import(acctTrans).Precommit([]uint32{1})
	committer.Commit(2).Clear()
	writeCache.Clear()

	outV, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{1, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{2, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 45, 67}) {
		t.Error("Error: The path should exist", outV)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 199, 199})); err != nil {
		t.Error(err)
	}

	outV, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 199, 199}) {
		t.Error("Error: The path should exist", outV)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), nil); err != nil {
		t.Error(err)
	}

	outV, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV != nil {
		t.Error("Error: The path should not exist", outV)
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer = stgcommitter.NewStateCommitter(store)
	committer.Import(acctTrans).Precommit([]uint32{1})
	committer.Commit(2).Clear()
	writeCache.Clear()
}

func TestNewCommitterWithoutCache(t *testing.T) {
	// CommitterCache(stgproxy.NewStoreProxy().EnableCache(), t) // Use cache
	CommitterCache(stgproxy.NewStoreProxy().DisableCache(), t) // Don't use cache
}

func TestSize(t *testing.T) {
	// store := chooseDataStore()
	store := chooseDataStore()

	key := RandomKey(0)
	alice := AliceAccount()
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer := stgcommitter.NewStateCommitter(store)
	committer.Import(acctTrans)
	committer.Precommit([]uint32{1})
	committer.Commit(0)
	writeCache.Clear()

	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele0", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	buffer := slice.New[byte](320, 11)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+key, noncommutative.NewBytes(buffer)); err != nil {
		t.Error(err)
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path)); v == nil {
		t.Error("Error: The path should exist")
	}

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele0", new(noncommutative.String))
	if v == nil || v.(string) != "124" {
		t.Error("Error: The path should exist")
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer = stgcommitter.NewStateCommitter(store)
	committer.Import(acctTrans)
	committer.Precommit([]uint32{1})
	committer.Commit(0)
	writeCache.Clear()

	outV, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+key, new(noncommutative.Bytes))
	if !bytes.Equal(outV.([]byte), slice.New[byte](320, 11)) {
		t.Error("Error: The path should exist")
	}
}

func TestNativeStorageReadWrite(t *testing.T) {
	store := chooseDataStore()
	store.(*stgproxy.StorageProxy).DisableCache()
	store.(*stgproxy.StorageProxy).EthStore().DisableAccountCache()

	alice := AliceAccount()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	// _, trans := writeCache.Export(importer.Sorter)
	trans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	acctTrans := (&univalue.Univalues{}).Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues)

	//values := univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).([]*univalue.Univalue)
	ts := univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues)

	committer := stgcommitter.NewStateCommitter(store)
	committer.Import(ts)
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)
	committer.SetStore(store)
	writeCache.Clear()

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000011", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000022", noncommutative.NewString("1111")); err != nil {
		t.Error(err)
	}

	committer = stgcommitter.NewStateCommitter(store)
	transitions := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer.Import(transitions)
	committer.Precommit([]uint32{1})
	committer.Commit(0)
	writeCache.Clear()

	_0, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000011", new(noncommutative.String))
	if !reflect.DeepEqual(_0, "124") {
		t.Error("Error: Should be empty!!", _0)
	}

	_1, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000022", new(noncommutative.String))
	if !reflect.DeepEqual(_1, "1111") {
		t.Error("Error: Should be empty!!", _1)
	}
}

func TestReadWriteAt(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele0", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele1", noncommutative.NewString("1111")); err != nil {
		t.Error(err)
	}

	_0, _, _ := writeCache.ReadAt(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/", 0, new(noncommutative.String))
	if !reflect.DeepEqual(_0, "124") {
		t.Error("Error: Should be empty!!")
	}

	_1, _, _ := writeCache.ReadAt(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/", 1, new(noncommutative.String))
	if !reflect.DeepEqual(_1, "1111") {
		t.Error("Error: Should be empty!!")
	}

	writeCache.WriteAt(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/", 0, noncommutative.NewString("456"))

	_0, _, _ = writeCache.ReadAt(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/", 0, new(noncommutative.String))
	if !reflect.DeepEqual(_0, "456") {
		t.Error("Error: Should be empty!!")
	}

	writeCache.WriteAt(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/", 0, nil) // Delete the first one

	_0, _, _ = writeCache.ReadAt(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/", 1, new(noncommutative.String))
	if !reflect.DeepEqual(_0, "1111") {
		t.Error("Error: Should be empty!!")
	}
}

func TestAddThenDeletePath2(t *testing.T) {
	store := chooseDataStore()
	store.(*stgproxy.StorageProxy).DisableCache()

	alice := AliceAccount()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	// _, trans := writeCache.Export(importer.Sorter)
	trans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	acctTrans := (&univalue.Univalues{}).Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues)

	//values := univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).([]*univalue.Univalue)
	ts := univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues)

	committer := stgcommitter.NewStateCommitter(store)
	committer.Import(ts)
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)

	committer.SetStore(store)
	writeCache.Clear()

	// create a path
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000011", noncommutative.NewString("124")); err != nil {
	// 	t.Error(err)
	// }

	transitions := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	newTrans := (&univalue.Univalues{}).Decode(univalue.Univalues(transitions).Encode()).(univalue.Univalues)
	committer = stgcommitter.NewStateCommitter(store)
	committer.Import(newTrans)

	// committer.Precommit([]uint32{1})
	// committer.Commit(0)
	// writeCache.Clear()

	// v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{})
	// if v == nil {
	// 	t.Error("Error: The path should exist")
	// }

	// committer.SetStore(store)
	// if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", nil); err != nil { // Delete the path
	// 	t.Error(err)
	// }

	// trans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	// committer = stgcommitter.NewStateCommitter(store)
	// committer.Import((&univalue.Univalues{}).Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))
	// committer.Precommit([]uint32{1})
	// committer.Commit(0)
	// committer.SetStore(store)

	// writeCache.Clear()
	// if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path)); v != nil {
	// 	t.Error("Error: The path should have been deleted")
	// }

	// if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000011", new(noncommutative.String)); v == nil {
	// 	t.Error("Error: The path should exist")
	// }
}

func TestBasic(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	committer := stgcommitter.NewStateCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)
	// univalue.Univalues(acctTrans).Print()

	committer.SetStore(store)
	writeCache.Clear()
	// create a path
	// path := commutative.NewPath()
	// if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", path); err != nil {
	// 	t.Error(err)
	// }

	// Try to rewrite a path, should fail !
	// if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", noncommutative.NewString("path")); err == nil {
	// 	t.Error(err)
	// }

	// Write a long string
	// str := string(slice.New[byte](320, 11))
	// if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", noncommutative.NewString(str)); err == nil {
	// 	t.Error(err)
	// }

	// Try to read an NONEXISTENT path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container", nil); value != nil {
		t.Error("Error: Path shouldn't be not found")
	}

	// Try to read an NONEXIST nonexistent entry from an nonexistent path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem-000", nil); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// try again
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem-000", nil); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// try to read an nonexistent path
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem-000", nil); value != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/container/elem-000")
	}

	// Write the entry
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/container/elem-000")
	}

	// Write the entry
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem-111", noncommutative.NewInt64(9999)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/container/elem-111")
	}

	// if v, _ := committer.Find(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", noncommutative.NewInt64(1111)); v != nil {
	// 	t.Error("Error: The path should have been deleted")
	// }

	// Read the entry back
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem-000", new(noncommutative.Int64)); value.(int64) != 1111 {
		t.Error("Error: Wrong value")
	}

	// Read the path
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path)); value == nil {
		t.Error(value)
	} else {
		target := value.(*deltaset.DeltaSet[string])
		k0, _ := target.GetByIndex(0)
		k1, _ := target.GetByIndex(1)
		if !reflect.DeepEqual([]string{k0, k1}, []string{"elem-000", "elem-111"}) {
			t.Error("Error: Wrong value !!!!")
		}
	}

	trans := slice.Clone(writeCache.Export(importer.Sorter))
	transitions := univalue.Univalues(trans).To(importer.ITTransition{})

	if !reflect.DeepEqual(transitions[0].Value().(interfaces.Type).Delta().(*deltaset.DeltaSet[string]).Updated().Elements(), []string{"elem-000", "elem-111"}) {
		t.Error("Error: keys are missing from the Updated buffer!", transitions[0].Value().(interfaces.Type).Delta().(*deltaset.DeltaSet[string]).Updated())
	}

	value := transitions[1].Value()
	if *value.(*noncommutative.Int64) != 1111 {
		t.Error("Error: keys don't match")
	}

	// wrong condition, value should still exists
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", &commutative.Path{}); value == nil {
		t.Error("Error: The variable has been cleared !")
	}

	// data := univalue.Univalues(transitions).To(importer.IPTransition{})

	buffer := univalue.Univalues(univalue.Univalues(transitions).To(importer.IPTransition{})).Encode()
	committer = stgcommitter.NewStateCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(buffer).(univalue.Univalues))
	// committer.Import(committer.Decode(univalue.Univalues(transitions).Encode()))

	committer.Precommit([]uint32{1})
	committer.Commit(0)

	/* =========== The second cycle ==============*/
	//try reading an element written in the previous cycle
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem-000", new(noncommutative.Int64)); value == nil {
		t.Error("Error: Entry not found")
	}

	bob := BobAccount()
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+bob+"/storage/container/elem-000", new(noncommutative.Int64)); value != nil {
		t.Error("Error: Wrong value")
	}
}

func TestCommitter(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()
	committer := stgcommitter.NewStateCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	raw := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(slice.Clone(raw)).To(importer.ITTransition{})
	// accesses := univalue.Univalues(slice.Clone(this.buffer)).To(importer.ITAccess{})

	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)

	writeCache.Clear()
	// committer.SetStore(store)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	// Write an entry having the the same name of a path, should go through
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0", noncommutative.NewString("ctrn-0")); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-0", noncommutative.NewString("elem-0")); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", new(noncommutative.Int64)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", noncommutative.NewInt64(5555)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-001", noncommutative.NewInt64(6666)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-002", noncommutative.NewInt64(7777)); err != nil {
		t.Error(err)
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil || value.(int64) != 5555 {
		t.Error("Error: Wrong value")
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-001", new(noncommutative.Int64)); value == nil || value.(int64) != 6666 {
		t.Error("Error: Wrong value")
	}

	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-002", new(noncommutative.Int64)); value == nil || value.(int64) != 7777 {
		t.Error("Error: Wrong value")
	}

	if meta, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{}); meta == nil {
		t.Error("Error: not found")
	}

	// Export all access records and state transitions
	transitions := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	// v, _, _ := transitions[0].Value().(interfaces.Type).Get()
	// if (*transitions[0].Value().(*noncommutative.String)) != "ctrn-0" {
	// 	t.Error("Error: keys don't match")
	// }

	addedkeys := codec.Strings(transitions[2].Value().(interfaces.Type).Delta().(*deltaset.DeltaSet[string]).Updated().Elements()).Sort()
	if !reflect.DeepEqual([]string(addedkeys), []string{"elem-0", "elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: keys don't match", addedkeys)
	}

	if meta, _, _ := writeCache.Read(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{}); meta == nil {
		t.Error("Error: The variable has been cleared")
	}
}

func TestCommitter2(t *testing.T) {
	store := chooseDataStore()
	// store := datastore.NewDataStore(nil, nil, nil, encoder, decoder)
	committer := stgcommitter.NewStateCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)

	committer.SetStore(store)
	// Create a new container
	path := commutative.NewPath()
	writeCache.Clear()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path); err != nil {
		t.Error(err, "Error:  Failed to MakePath: "+"/ctrn-0/")
	}

	// Add a vaiable directly
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("0000")); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/elem-0")
	}

	// Add the first element
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error(err, "Error: Failed to Write: "+"/ctrn-0/elem-000")
	}

	// Add the second element
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-001")
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
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
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil || value.(int64) != 1111 {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-000")
	}

	// Try to read an nonexistent element, should leave a access record
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-005", nil); value != nil {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-005")
	}

	// Update then return path meta info
	meta0, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{})
	keys := meta0.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Do again
	meta1, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{})
	keys = meta1.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Delete elem-00
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", nil); err != nil {
		t.Error("Error: Failed to delete: " + "blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/elem-000")
	}

	// The elem-00 has been deleted, only "elem-001", "elem-002" left
	meta0, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{})
	if !reflect.DeepEqual(meta0.(*deltaset.DeltaSet[string]).Elements(), []string{"elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	// Readd elem-00 back
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", noncommutative.NewInt64(9999)); err != nil { // delete
		t.Error("Error: Failed to write: " + "/ctrn-0/elem-000")
	}

	// Check elem-00's value
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", new(noncommutative.Int64)); value.(int64) != 9999 {
		t.Error("Error: The element wasn't successfully deleted")
	}

	// Update then read the path info again
	meta, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{})
	keys = meta.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: keys don't match", keys, "Expecting", []string{"elem-000", "elem-001", "elem-002"})
	}

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", new(noncommutative.Int64))
	if v == nil {
		t.Error("Error: keys don't match")
	}

	// if value, _ := committer.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000"); (*value.(*noncommutative.Int64)) != 9999 {
	// 	t.Error("Error: The element wasn't successfully deleted")
	// }

	/* Remove the path and all the elements underneath */
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", nil); err != nil { // Delete the path and its sub paths
		t.Error(err, "Failed to remove path: "+"/ctrn-0/")
	}

	if v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{}); v != nil { /* The path should be gone by now */
		t.Error("Error: The key should not exist!")
	}

	if v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-0", new(noncommutative.Int64)); v != nil { /* all the sub paths should be gone by now*/
		t.Error("Error: The key should not exist!")
	}

	/*  Read the storage path to see what is left*/
	v, _, _ = writeCache.Read(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", new(commutative.Path))
	keys = v.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{}) {
		t.Error("Error: Should be empty!!")
	}

	v, _, _ = writeCache.Read(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", new(commutative.Path))
	keys = v.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{}) {
		t.Error("Error: Should be empty!!")
	}

	/*  Export all */
	// accessRecords, transitions := writeCache.Export(importer.Sorter)
	accessRecords := univalue.Univalues(slice.Clone(writeCache.Export())).To(importer.ITAccess{})
	transitions := univalue.Univalues(slice.Clone(writeCache.Export())).To(importer.ITTransition{})

	// 3 writes + 1 affiliated write
	value := univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", 3, 4, 0, nil, nil)
	if !univalue.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-001", 1, 1, 0, nil, nil)
	if !univalue.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-002", 0, 1, 0, nil, nil)
	if !univalue.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-005", 1, 0, 0, nil, nil)
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

func TestTransientDBv2(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()

	committer := stgcommitter.NewStateCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(slice.Clone(raw)).To(importer.IPTransition{})

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
	// policy := datastore.NewCachePolicy(0, 1)
	// store := datastore.NewDataStore(nil, policy, memdb.NewMemoryDB(), stgproxy.Rlp{}.Encode, stgproxy.Rlp{}.Decode)
	store := chooseDataStore()
	alice := AliceAccount()
	committer := stgcommitter.NewStateCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(slice.Clone(raw)).To(importer.IPTransition{})

	buffer := univalue.Univalues(acctTrans).Encode()
	univalue.Univalues{}.Decode(buffer)
	committer.Import(acctTrans)

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)

	committer.SetStore(store)
	writeCache.Clear()
	// commutative.NewU256Delta(100)
	// writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", &commutative.U256{value: 100})

	value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/balance", &commutative.U256{})
	valueAdd := value.(uint256.Int)
	if value == nil || (&valueAdd).ToBig().Uint64() != 0 {
		t.Error("Error: Wrong value", value.(*uint256.Int).ToBig().Uint64())
	}
}

func TestPathReadAndWriteCommutatives(b *testing.T) {
	store := chooseDataStore()
	writeCache := NewWriteCacheWithAcounts(store, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	basev := commutative.NewBoundedU256FromU64(0, 999)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), basev); err != nil {
		b.Error(err)
	}

	trans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer := stgcommitter.NewStateCommitter(store).Import(trans)
	committer.Precommit([]uint32{0})
	committer.Commit(0)

	writeCache = NewWriteCacheWithAcounts(store)
	delta := commutative.NewU256DeltaFromU64(uint64(11), true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), delta); err != nil {
		b.Error(err)
	}

	trans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer = stgcommitter.NewStateCommitter(store).Import(trans)
	committer.Precommit([]uint32{0})
	committer.Commit(0)

	writeCache = NewWriteCacheWithAcounts(store)
	if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), new(commutative.U256)); v == nil ||
		v.(uint256.Int) != *uint256.NewInt(11) {
		b.Error(err)
	}

	delta = commutative.NewU256DeltaFromU64(uint64(11), true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), delta); err != nil {
		b.Error(err)
	}

	delta2 := commutative.NewU256DeltaFromU64(uint64(11), true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), delta2); err != nil {
		b.Error(err)
	}

	trans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer = stgcommitter.NewStateCommitter(store).Import(trans)
	committer.Precommit([]uint32{0})
	committer.Commit(0)

	writeCache = NewWriteCacheWithAcounts(store)
	if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), new(commutative.U256)); v == nil ||
		v.(uint256.Int) != *uint256.NewInt(33) {
		b.Error(err)
	}
}

func TestPathReadAndWriteBatchCache2(b *testing.T) {
	store := chooseDataStore()
	writeCache := NewWriteCacheWithAcounts(store, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), noncommutative.NewInt64(int64(11))); err != nil {
		b.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(1), noncommutative.NewInt64(int64(22))); err != nil {
		b.Error(err)
	}

	committer := stgcommitter.NewStateCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(0)

	writeCache = NewWriteCacheWithAcounts(store)
	if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), new(noncommutative.Int64)); v == nil ||
		v.(int64) != int64(11) {
		b.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), noncommutative.NewInt64(int64(911))); err != nil {
		b.Error(err)
	}

	committer = stgcommitter.NewStateCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(0)

	if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), new(noncommutative.Int64)); v == nil ||
		v.(int64) != int64(911) {
		b.Error(err)
	}
}

func TestPathReadAndWriteBatch(b *testing.T) {
	store := chooseDataStore()
	writeCache := NewWriteCacheWithAcounts(store, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys := RandomKeys(0, 100)
	t0 := time.Now()
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("First Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}

		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i+len(keys)) {
			b.Error(err)
		}
	}
	fmt.Println("First Read time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("1. New Path Write time:", time.Since(t0))

	t0 = time.Now()
	committer := stgcommitter.NewStateCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(0)
	fmt.Println("Commit time:", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}

		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i+len(keys)) {
			b.Error(err)
		}
	}
	fmt.Println("2. First read time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(len(keys), len(keys)+10)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("Second Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(100000, 100000*2)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("2. New Path Write time:", len(keys), "keys in", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}
	}
	fmt.Println("2. Second read time:", len(keys), "keys in", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(100000*2, 100000*2+20)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println(" New Path Write time:", len(keys), "keys in", time.Since(t0))
}

func TestPathReadAndWrites(b *testing.T) {
	store := chooseDataStore()
	writeCache := NewWriteCacheWithAcounts(store, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys := RandomKeys(0, 100)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}

	committer := stgcommitter.NewStateCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(0)

	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}

		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i+len(keys)) {
			b.Error(err)
		}
	}

	keys = RandomKeys(101, 200)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}

	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}

		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i+len(keys)) {
			b.Error(err)
		}
	}
}

func TestPathReadAndWritesPath(b *testing.T) {
	store := chooseDataStore()

	writeCache := eucache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	bob := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, bob, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys1, key2 := RandomKey(1), RandomKey(2)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+keys1, commutative.NewUnboundedU256()); err != nil {
		b.Error(err)
	}

	u256 := new(commutative.U256).NewBoundedU256FromUint64(111, 0, 0, 999, true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+key2, u256); err != nil {
		b.Error(err)
	}

	committer := stgcommitter.NewStateCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(0)

	writeCache = eucache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if typedv, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path)); typedv == nil {
		b.Error("Error: Failed to read the Path !")
	}

	if typedv, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+keys1, new(commutative.U256)); typedv == nil {
		b.Error("Error: Failed to read the key !")
	}

	u256 = new(commutative.U256).NewBoundedU256FromUint64(333, 0, 0, 999, true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+"8975", u256); err != nil {
		b.Error(err)
	}

	typedv, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+"8975", new(commutative.U256))
	tv := typedv.(uint256.Int)
	if tv.Cmp(uint256.NewInt(333)) != 0 {
		b.Error("Error: Failed to read the key !")
	}

	typedv, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path))
	if typedv == nil || typedv.(*deltaset.DeltaSet[string]).Length() != 3 {
		b.Error("Error: Failed to read the key !", typedv.(*deltaset.DeltaSet[string]).Length())
	}
}

func TestEthDataStoreAddDeleteRead(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()

	// writeCache := committer.WriteCache()
	writeCache := eucache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	committer := stgcommitter.NewStateCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)

	committer.SetStore(store)
	// create a path
	writeCache.Clear()

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", noncommutative.NewString("path")); err == nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", new(noncommutative.Int64)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err)
	}

	meta, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path))
	keys := meta.(*deltaset.DeltaSet[string]).Elements()
	if meta == nil || len(keys) != 2 ||
		keys[0] != "elem-000" ||
		keys[1] != "elem-001" {
		t.Error("not found")
	}
}

func TestPathMultiBatch(b *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	// bob := BobAccount()

	writeCache := NewWriteCacheWithAcounts(store, AliceAccount(), BobAccount())
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	committer := stgcommitter.NewStateCommitter(store)
	committer.Import(acctTrans)
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)
	writeCache.Clear()

	keys := RandomKeys(0, 5)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer = stgcommitter.NewStateCommitter(store)
	committer.Import(acctTrans)
	committer.Precommit([]uint32{0})
	committer.Commit(0)
	writeCache.Clear()

	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}
	}

	v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path))
	if v == nil || (v.(*deltaset.DeltaSet[string]).Length()) != uint64(len(keys)) {
		b.Error(err)
	}

	keys2 := RandomKeys(6, 8)
	for i := 0; i < len(keys2); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+keys2[i], noncommutative.NewInt64(int64(11))); err != nil {
			b.Error(err)
		}
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer = stgcommitter.NewStateCommitter(store)
	committer.Import(acctTrans)
	committer.Precommit([]uint32{0})
	committer.Commit(0)
	writeCache.Clear()

	for i, k := range keys {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+k, new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}
	}

	for _, k := range keys2 {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+k, new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(11) {
			b.Error(err)
		}
	}

	v, _, err = writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path))
	if v == nil || (v.(*deltaset.DeltaSet[string]).Length()) != 7 {
		b.Error(err, v.(*deltaset.DeltaSet[string]).Length())
	}

}
