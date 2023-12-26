package ccurltest

import (
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	"github.com/arcology-network/concurrenturl"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
)

func TestAuxTrans(t *testing.T) {
	store := chooseDataStore()
	url := ccurl.NewConcurrentUrl(store)
	writeCache := url.WriteCache()

	alice := AliceAccount()
	if _, err := concurrenturl.CreateNewAccount(ccurlcommon.SYSTEM, alice, ccurlcommon.NewPlatform(), writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans00 := url.WriteCache().Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM}) // Commit

	url.Init(store)
	// create a path
	path := commutative.NewPath()

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err == nil {
		t.Error(err)
	}

	// Try to read an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1", new(commutative.Path)); value != nil {
		t.Error("Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-000", new(noncommutative.String)); value != nil {
		t.Error("Shouldn't be not found")
	}

	//try again
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.String)); value != nil {
		t.Error("Shouldn't be not found")
	}

	// Write the entry
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Shouldn't be not found")
	}

	// Read the entry back
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value.(int64) != 1111 {
		t.Error("Shouldn't be not found")
	}

	// Read the path
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error(value)
	} else {
		keys := value.(*orderedset.OrderedSet).Keys()
		if !reflect.DeepEqual(keys, []string{"elem-000"}) {
			t.Error("Wrong value ")
		}
	}

	transitions := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})
	if !reflect.DeepEqual(transitions[0].Value().(interfaces.Type).Delta().(*commutative.PathDelta).Added(), []string{"elem-000"}) {
		t.Error("keys don't match")
	}

	value := transitions[1].Value()
	if *(value.(*noncommutative.Int64)) != 1111 {
		t.Error("keys don't match")
	}

	// wrong condition, value should still exists
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error("The variable has been cleared")
	}

	in := indexer.Univalues(transitions).Encode()
	out := indexer.Univalues{}.Decode(in).(indexer.Univalues)

	url.Import(out)
	url.Sort()
	url.Commit([]uint32{1})
}

func TestCheckAccessRecords(t *testing.T) {
	store := chooseDataStore()

	url := ccurl.NewConcurrentUrl(store)
	writeCache := url.WriteCache()
	alice := AliceAccount()
	if _, err := concurrenturl.CreateNewAccount(ccurlcommon.SYSTEM, alice, ccurlcommon.NewPlatform(), writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans00 := url.WriteCache().Export(indexer.Sorter)
	trans00 := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans00).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{1}) // Commit

	url.Init(store)
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	}

	// _, trans10 := url.WriteCache().Export(indexer.Sorter)
	trans10 := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans10).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{1}) // Commit

	url.Init(store)
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/1") // create a path
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewInt64(2222)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/2") // create a path
	}

	// accesses10, trans11 := url.WriteCache().Export(indexer.Sorter)
	// url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans11).Encode()).(indexer.Univalues))

	// url.Sort()
	// url.Commit([]uint32{1}) // Commit

	// url = ccurl.NewConcurrentUrl(store)
	// if len(trans11) != 3 {
	// 	t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/2") // create a path
	// }

	// if len(trans11) != 3 {
	// 	t.Error("Error: There should be 3 transitions in url") // create a path
	// }

	// if len(accesses10) != 3 {
	// 	t.Error("Error: There should be 3 accesse records url") // create a path
	// }

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/3", noncommutative.NewInt64(3333)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/3") // create a path
	}

	// if url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/3", noncommutative.NewInt64(4444)) != nil {
	// 	t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/3") // create a path
	// }

	v1, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path))
	keys := v1.(*orderedset.OrderedSet).Keys()
	if len(keys) != 3 {
		t.Error("Error: There should be 3 elements only!!! actual = ", len(keys)) // create a path
	}
}
