package ccurltest

import (
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	ccurl "github.com/arcology-network/concurrenturl"
	committercommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	importer "github.com/arcology-network/concurrenturl/importer"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	cache "github.com/arcology-network/eu/cache"
)

func TestAddAndDelete(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
	committer.Import(importer.Univalues{}.Decode(importer.Univalues(acctTrans).Encode()).(importer.Univalues))
	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM})

	// acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
	committer.Import(importer.Univalues{})
	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM})

	committer.Init(store)
	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	acctTrans = importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
	committer.Import(importer.Univalues{}.Decode(importer.Univalues(acctTrans).Encode()).(importer.Univalues))
	committer.Sort()
	committer.Commit([]uint32{1})

	committer.Init(store)
	writeCache.Clear()

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil); err == nil {
		t.Error("Deleting an non-existing entry should've flaged an error", err)
	}

	raw := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
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

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.String)); v != "124" {
		t.Error("Error: Wrong return value")
	}
}

func TestEmptyNodeSet(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
	committer.Import(importer.Univalues{}.Decode(importer.Univalues(acctTrans).Encode()).(importer.Univalues))
	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM})

	// acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
	committer.Import(importer.Univalues{})
	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM})

	committer.Import(importer.Univalues{})
	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM})
}

func TestRecursiveDeletionSameBatch(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})

	committer.Import(importer.Univalues{}.Decode(importer.Univalues(acctTrans).Encode()).(importer.Univalues))
	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM})

	committer.Init(store)
	// create a path
	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)
	// _, addPath := writeCache.Export(importer.Sorter)
	addPath := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})

	committer.Import(importer.Univalues{}.Decode(importer.Univalues(addPath).Encode()).(importer.Univalues))
	// committer.Import(committer.Decode(importer.Univalues(addPath).Encode()))
	committer.Sort()
	committer.Commit([]uint32{1})
	committer.Init(store)

	writeCache.Clear()

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewInt64(1))
	// _, addTrans := writeCache.Export(importer.Sorter)
	addTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
	// committer.Import(committer.Decode(importer.Univalues(addTrans).Encode()))
	committer.Import(importer.Univalues{}.Decode(importer.Univalues(addTrans).Encode()).(importer.Univalues))
	committer.Sort()
	committer.Commit([]uint32{1})
	writeCache.Clear()

	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v == nil {
		t.Error("Error: Failed to read the key !")
	}

	// url2 := ccurl.NewStorageCommitter(store)
	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v.(int64) != 1 {
		t.Error("Error: Failed to read the key !")
	}

	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", nil)
	// _, deleteTrans := url2.WriteCache().Export(importer.Sorter)
	deleteTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})

	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v != nil {
		t.Error("Error: Failed to read the key !")
	}

	url3 := ccurl.NewStorageCommitter(store)
	url3.Import(append(addTrans, deleteTrans...))
	url3.Sort()
	url3.Commit([]uint32{1, 2})
	writeCache.Clear()

	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v != nil {
		t.Error("Error: Failed to delete the entry !")
	}
}

func TestApplyingTransitionsFromMulitpleBatches(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)

	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})

	committer := ccurl.NewStorageCommitter(store)
	committer.Import(acctTrans)
	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM})

	committer.Init(store)
	path := commutative.NewPath()
	_, err := writeCache.Write(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path)

	if err != nil {
		t.Error("error")
	}

	acctTrans = importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
	committer.Import(importer.Univalues{}.Decode(importer.Univalues(acctTrans).Encode()).(importer.Univalues))
	committer.Sort()
	committer.Commit([]uint32{1})

	committer.Init(store)

	writeCache.Clear()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil)

	if acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{}); len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionDifferentBatch(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)

	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})

	in := importer.Univalues(acctTrans).Encode()
	out := importer.Univalues{}.Decode(in).(importer.Univalues)
	// committer.Import(committer.Decode(importer.Univalues(out).Encode()))

	committer := ccurl.NewStorageCommitter(store)
	committer.Import(importer.Univalues{}.Decode(importer.Univalues(out).Encode()).(importer.Univalues))
	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM})

	committer.Init(store)
	// create a path
	path := commutative.NewPath()
	writeCache.Clear()

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("1"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("2"))

	acctTrans = importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
	in = importer.Univalues(acctTrans).Encode()
	out = importer.Univalues{}.Decode(in).(importer.Univalues)
	// committer.Import(committer.Decode(importer.Univalues(out).Encode()))
	committer.Import(importer.Univalues{}.Decode(importer.Univalues(out).Encode()).(importer.Univalues))
	committer.Sort()
	committer.Commit([]uint32{1})

	_1, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.String))
	if _1 != "1" {
		t.Error("Error: Not match")
	}

	committer.Init(store)
	writeCache.Clear()

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("3"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("4"))

	outpath, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := outpath.(*orderedset.OrderedSet).Keys()
	if reflect.DeepEqual(keys, []string{"1", "2", "3", "4"}) {
		t.Error("Error: Not match")
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil) // delete the path
	if acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{}); len(acctTrans) != 3 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestStateUpdate(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)

	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	// _, initTrans := writeCache.Export(importer.Sorter)
	initTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})

	// committer.Import(committer.Decode(importer.Univalues(initTrans).Encode()))
	committer := ccurl.NewStorageCommitter(store)
	committer.Import(importer.Univalues{}.Decode(importer.Univalues(initTrans).Encode()).(importer.Univalues))
	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM})
	committer.Init(store)

	writeCache.Clear()
	tx0bytes, trans, err := Create_Ctrn_0(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx0Out := importer.Univalues{}.Decode(tx0bytes).(importer.Univalues)
	tx0Out = trans
	tx1bytes, err := Create_Ctrn_1(alice, store)
	if err != nil {
		t.Error(err)
	}

	tx1Out := importer.Univalues{}.Decode(tx1bytes).(importer.Univalues)

	// committer.Import(committer.Decode(importer.Univalues(append(tx0Out, tx1Out...)).Encode()))
	committer.Import((append(tx0Out, tx1Out...)))
	committer.Sort()
	committer.Commit([]uint32{0, 1})
	//need to encode delta only now it encodes everything

	writeCache.Clear()
	if err := CheckPaths(alice, writeCache); err != nil {
		t.Error(err)
	}

	v, _, _ := writeCache.Read(9, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{}) //system doesn't generate sub paths for /storage/
	// if v.(*commutative.Path).CommittedLength() != 2 {
	// 	t.Error("Error: Wrong sub paths")
	// }

	// if !reflect.DeepEqual(v.([]string), []string{"ctrn-0/", "ctrn-1/"}) {
	// 	t.Error("Error: Didn't find the subpath!")
	// }

	v, _, _ = writeCache.Read(9, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := v.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		t.Error("Error: Keys don't match !")
	}

	// Delete the container-0
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil {
		t.Error("Error: Cann't delete a path twice !")
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
		t.Error("Error: The path should be gone already !")
	}

	transitions := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITCTransition{})
	out := importer.Univalues{}.Decode(importer.Univalues(transitions).Encode()).(importer.Univalues)

	committer.Import(out)
	committer.Sort()
	committer.Commit([]uint32{1})

	writeCache.Clear()
	if v, _, _ := writeCache.Read(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
		t.Error("Error: Should be gone already !")
	}
}
