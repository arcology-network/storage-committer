package ccurltest

import (
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	ccurl "github.com/arcology-network/concurrenturl"
	committercommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/importer"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	cache "github.com/arcology-network/eu/cache"
)

func TestAuxTrans(t *testing.T) {
	store := chooseDataStore()
	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans00 := writeCache.Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	committer.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))

	committer.Sort()
	committer.Commit([]uint32{committercommon.SYSTEM}) // Commit

	committer.Init(store)
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
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value.(int64) != 1111 {
		t.Error("Shouldn't be not found")
	}

	// Read the path
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error(value)
	} else {
		keys := value.(*orderedset.OrderedSet).Keys()
		if !reflect.DeepEqual(keys, []string{"elem-000"}) {
			t.Error("Wrong value ")
		}
	}

	transitions := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	if !reflect.DeepEqual(transitions[0].Value().(interfaces.Type).Delta().(*commutative.PathDelta).Added(), []string{"elem-000"}) {
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

	in := indexer.Univalues(transitions).Encode()
	out := indexer.Univalues{}.Decode(in).(indexer.Univalues)

	committer.Import(out)
	committer.Sort()
	committer.Commit([]uint32{1})
}

func TestCheckAccessRecords(t *testing.T) {
	store := chooseDataStore()

	committer := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans00 := writeCache.Export(indexer.Sorter)
	trans00 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	committer.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans00).Encode()).(indexer.Univalues))

	committer.Sort()
	committer.Commit([]uint32{1}) // Commit

	committer.Init(store)
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	}

	// _, trans10 := writeCache.Export(indexer.Sorter)
	trans10 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	committer.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans10).Encode()).(indexer.Univalues))

	committer.Sort()
	committer.Commit([]uint32{1}) // Commit

	committer.Init(store)
	writeCache.Clear()

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/1") // create a path
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewInt64(2222)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/2") // create a path
	}

	// accesses10, trans11 := writeCache.Export(indexer.Sorter)
	// committer.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans11).Encode()).(indexer.Univalues))

	// committer.Sort()
	// committer.Commit([]uint32{1}) // Commit

	// committer = ccurl.NewStorageCommitter(store)
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
	keys := v1.(*orderedset.OrderedSet).Keys()
	if len(keys) != 3 {
		t.Error("Error: There should be 3 elements only!!! actual = ", len(keys)) // create a path
	}
}
