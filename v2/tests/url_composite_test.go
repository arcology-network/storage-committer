package ccurltest

import (
	"reflect"
	"testing"

	ccurl "github.com/arcology/concurrenturl/v2"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
	commutative "github.com/arcology/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology/concurrenturl/v2/type/noncommutative"
	assert "github.com/magiconair/properties/assert"
)

func TestAuxTrans(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	_, trans00 := url.Export(true)
	url.Commit(trans00, []uint32{ccurlcommon.SYSTEM}) // Commit

	// create a path
	path, err := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Failed to create the path")
	}

	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", "path"); err == nil {
		t.Error(err)
	}

	// Try to read an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-1"); value != nil {
		t.Error("Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-1/elem-000"); value != nil {
		t.Error("Shouldn't be not found")
	}

	//try again
	if value, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Shouldn't be not found")
	}

	// try to read an nonexistent path
	if value, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Shouldn't be not found")
	}

	// Write the entry
	if value := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); value != nil {
		t.Error("Shouldn't be not found")
	}

	// Read the entry back
	if value, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-000"); (*value.(*noncommutative.Int64)) != 1111 {
		t.Error("Shouldn't be not found")
	}

	// Read the path
	if value, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/"); value == nil {
		t.Error(value)
	} else {
		if !reflect.DeepEqual(value.(*commutative.Meta).GetKeys(), []string{"elem-000"}) {
			t.Error("Wrong value ")
		}
	}

	_, transitions := url.Export(true)
	if !reflect.DeepEqual(transitions[1].Value().(*commutative.Meta).GetAdded(), []string{"elem-000"}) {
		t.Error("keys don't match")
	}

	value := transitions[2].Value()
	if *(value.(*noncommutative.Int64)) != 1111 {
		t.Error("keys don't match")
	}

	// wrong condition, value should still exists
	if value, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/"); value == nil {
		t.Error("The variable has been cleared")
	}

	url.Indexer().Import(transitions)
	if _, _, errs := url.Indexer().Commit([]uint32{1}); len(errs) != 0 {
		t.Error(errs)
	}
}

func TestCheckAccessRecords(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url1 := ccurl.NewConcurrentUrl(store)
	if err := url1.Preload(ccurlcommon.SYSTEM, url1.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	_, trans00 := url1.Export(true)
	url1.Commit(trans00, []uint32{1}) // Commit

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	if url1.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path) != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	}

	_, trans10 := url1.Export(true)
	url1.Commit(trans10, []uint32{1}) // Commit

	if url1.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/1", noncommutative.NewInt64(1111)) != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/1") // create a path
	}

	if url1.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/2", noncommutative.NewInt64(2222)) != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/2") // create a path
	}

	accesses10, trans11 := url1.Export(true)
	url1.Commit(trans11, []uint32{1}) // Commit

	if len(trans11) != 3 {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/2") // create a path
	}

	assert.Equal(t, len(trans11), 3, "Error: There should be 3 transitions in url1")
	assert.Equal(t, len(accesses10), 3, "Error: There should be 3 accesse records url1")

	if url1.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/3", noncommutative.NewInt64(3333)) != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/3") // create a path
	}

	if url1.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/3", noncommutative.NewInt64(4444)) != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/3") // create a path
	}

	v1, _ := url1.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	keys := v1.(*commutative.Meta).GetKeys()
	if len(keys) != 3 {
		t.Error("Error: There should be 3 elements only") // create a path
	}
}
