package ccurltest

import (
	"reflect"
	"testing"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
)

func TestAuxTrans(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)

	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	// _, trans00 := url.Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)

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
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", noncommutative.NewString("path")); err == nil {
		t.Error(err)
	}

	// Try to read an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1"); value != nil {
		t.Error("Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-000"); value != nil {
		t.Error("Shouldn't be not found")
	}

	//try again
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Shouldn't be not found")
	}

	// try to read an nonexistent path
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Shouldn't be not found")
	}

	// Write the entry
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Shouldn't be not found")
	}

	// Read the entry back
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); value.(int64) != 1111 {
		t.Error("Shouldn't be not found")
	}

	// Read the path
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); value == nil {
		t.Error(value)
	} else {
		if !reflect.DeepEqual(value.([]string), []string{"elem-000"}) {
			t.Error("Wrong value ")
		}
	}

	transitions := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)
	if !reflect.DeepEqual(transitions[0].Value().(ccurlcommon.TypeInterface).Delta().(*commutative.PathDelta).Added(), []string{"elem-000"}) {
		t.Error("keys don't match")
	}

	value := transitions[1].Value()
	if *(value.(*noncommutative.Int64)) != 1111 {
		t.Error("keys don't match")
	}

	// wrong condition, value should still exists
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); value == nil {
		t.Error("The variable has been cleared")
	}

	in := indexer.Univalues(transitions).Encode()
	out := indexer.Univalues{}.Decode(in).(indexer.Univalues)

	url.Import(out)
	url.Sort()
	url.Commit([]uint32{1})
}

func TestCheckAccessRecords(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	// _, trans00 := url.Export(indexer.Sorter)
	trans00 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans00).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{1}) // Commit

	url.Init(store)
	path := commutative.NewPath()
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	}

	// _, trans10 := url.Export(indexer.Sorter)
	trans10 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)

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

	// accesses10, trans11 := url.Export(indexer.Sorter)
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

	v1, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	keys := v1.([]string)
	if len(keys) != 3 {
		t.Error("Error: There should be 3 elements only!!! actual = ", len(keys)) // create a path
	}
}
