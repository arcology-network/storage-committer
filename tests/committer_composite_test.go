package committertest

import (
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	cache "github.com/arcology-network/eu/cache"
	stgcommitter "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	importer "github.com/arcology-network/storage-committer/importer"
	"github.com/arcology-network/storage-committer/interfaces"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	platform "github.com/arcology-network/storage-committer/platform"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

func TestAuxTrans(t *testing.T) {
	store := chooseDataStore()
	committer := stgcommitter.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans00 := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0) // Commit

	committer.SetStore(store)
	// create a path
	writeCache.Clear()

	path := commutative.NewPath()

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err == nil {
		t.Error(err)
	}

	// Try to read an nonexistent path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1", new(commutative.Path)); value != nil {
		t.Error("Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-000", new(noncommutative.String)); value != nil {
		t.Error("Shouldn't be not found")
	}

	//try again
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.String)); value != nil {
		t.Error("Shouldn't be not found")
	}

	// Write the entry
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Shouldn't be not found")
	}

	// Read the entry back
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil || value.(int64) != 1111 {
		t.Error("Shouldn't be not found")
	}

	// Read the path
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error(value)
	} else {
		keys := value.(*deltaset.DeltaSet[string]).Elements()
		if !reflect.DeepEqual(keys, []string{"elem-000"}) {
			t.Error("Wrong value ")
		}
	}

	transitions := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	if len(transitions) == 0 || !reflect.DeepEqual(transitions[0].Value().(interfaces.Type).Delta().(*deltaset.DeltaSet[string]).Updated().Elements(), []string{"elem-000"}) {
		t.Error("keys don't match")
	}

	value := transitions[1].Value()
	if *(value.(*noncommutative.Int64)) != 1111 {
		t.Error("keys don't match")
	}

	// wrong condition, value should still exists
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error("The variable has been cleared")
	}

	in := univalue.Univalues(transitions).Encode()
	out := univalue.Univalues{}.Decode(in).(univalue.Univalues)

	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(out)

	committer.Precommit([]uint32{1})
	committer.Commit(0)
}

func TestCheckAccessRecords(t *testing.T) {
	store := chooseDataStore()

	committer := stgcommitter.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans00 := writeCache.Export(importer.Sorter)
	trans00 := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans00).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{1})
	committer.Commit(0) // Commit

	committer.SetStore(store)
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	}

	// _, trans10 := writeCache.Export(importer.Sorter)
	trans10 := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans10).Encode()).(univalue.Univalues))
	committer.Precommit([]uint32{1})
	committer.Commit(0) // Commit

	committer.SetStore(store)
	writeCache.Clear()

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/1") // create a path
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewInt64(2222)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/2") // create a path
	}

	// accesses10, trans11 := writeCache.Export(importer.Sorter)
	// committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans11).Encode()).(univalue.Univalues))

	//
	// committer.Precommit([]uint32{1})
	// committer = stgcommitter.NewStorageCommitter(store)
	// committer.Commit(0) // Commit

	// committer = stgcommitter.NewStorageCommitter(store)
	// if len(trans11) != 3 {
	// 	t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/2") // create a path
	// }

	// if len(trans11) != 3 {
	// 	t.Error("Error: There should be 3 transitions in committer") // create a path
	// }

	// if len(accesses10) != 3 {
	// 	t.Error("Error: There should be 3 accesse records committer") // create a path
	// }

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/3", noncommutative.NewInt64(3333)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/3") // create a path
	}

	// if writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/3", noncommutative.NewInt64(4444)) != nil {
	// 	t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/3") // create a path
	// }

	v1, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path))
	if v1 == nil {
		t.Error("Error: Failed to read blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	}

	keys := v1.(*deltaset.DeltaSet[string]).Elements()
	if len(keys) != 3 {
		t.Error("Error: There should be 3 elements only!!! actual = ", len(keys)) // create a path
	}
}
