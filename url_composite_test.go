package concurrenturl

import (
	"reflect"
	"testing"

	ccurlcommon "github.com/HPISTechnologies/concurrenturl/common"
	commutative "github.com/HPISTechnologies/concurrenturl/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/type/noncommutative"
	assert "github.com/magiconair/properties/assert"
)

func TestComp(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}

	_, transitions := url.Export()
	for _, v := range transitions {
		v.Print()
	}
}

func TestAuxTrans(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	// create a path
	path, err := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Failed to create the path")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", "path"); err == nil {
		t.Error(err)
	}

	// Try to read an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-1"); value != nil {
		t.Error("Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-1/elem-000"); value != nil {
		t.Error("Shouldn't be not found")
	}

	//try again
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Shouldn't be not found")
	}

	// try to read an nonexistent path
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Shouldn't be not found")
	}

	// Write the entry
	if value := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); value != nil {
		t.Error("Shouldn't be not found")
	}

	// Read the entry back
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); (*value.(*noncommutative.Int64)) != 1111 {
		t.Error("Shouldn't be not found")
	}

	// Read the path
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/"); value == nil {
		t.Error(value)
	} else {
		if !reflect.DeepEqual(value.(*commutative.Meta).GetKeys(), []string{"elem-000"}) {
			t.Error("Wrong value ")
		}
	}

	_, transitions := url.Export()
	if !reflect.DeepEqual(transitions[6].GetValue().(*commutative.Meta).GetKeys(), []string{"elem-000"}) {
		t.Error("keys don't match")
	}

	value := transitions[7].GetValue()
	if *(value.(*noncommutative.Int64)) != 1111 {
		t.Error("keys don't match")
	}

	// wrong condition, value should still exists
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/"); value == nil {
		t.Error("The variable has been cleared")
	}

	url.indexer.Import(transitions)
	if errs := url.indexer.Commit([]uint32{1}); len(errs) != 0 {
		t.Error(errs)
	}

	/* =========== The second cycle ==============*/
	//try reading an element written in the previous cycle
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); value == nil {
		t.Error("Entry not found")
	}
}

func TestCheckAccessRecords(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url1 := NewConcurrentUrl(store)
	if err := url1.Preload(ccurlcommon.SYSTEM, url1.Platform.Eth10(), "Alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	url1.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", path) // create a path

	_, trans10 := url1.Export()
	url1.Commit(trans10, []uint32{1}) // Commit

	url1.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/1", noncommutative.NewInt64(1111))
	url1.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/2", noncommutative.NewInt64(2222))

	(*url1.indexer.Store()).Print()

	accesses10, trans11 := url1.Export()
	assert.Equal(t, len(trans11), 3, "Error: There should be 3 transitions")
	assert.Equal(t, len(accesses10), 3, "Error: There should be 3 accesse records")

	/* A new url*/
	url2 := NewConcurrentUrl(store)
	if err := url1.Preload(ccurlcommon.SYSTEM, url2.Platform.Eth10(), "Alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	url2.Commit(trans10, []uint32{1}) // blcc://eth1.0/Alice/storage/ctrn-0/
	url2.Commit(trans11, []uint32{1}) // 1, 2

	url2.Write(2, "blcc://eth1.0/Alice/storage/ctrn-0/3", noncommutative.NewInt64(3333))
	url2.Write(2, "blcc://eth1.0/Alice/storage/ctrn-0/4", noncommutative.NewInt64(4444))

	_, trans20 := url2.Export()
	assert.Equal(t, len(trans20), 3, "Error: There should be 3 transitions")
	url2.Commit(trans20, []uint32{2}) // Commit
	(*url2.indexer.Store()).Print()

	//url1.Commit(trans11, []uint32{1}) // Commit.
	v1, _ := url1.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/")
	assert.Equal(t, len(v1.(ccurlcommon.TypeInterface).Value().([]string)), 4, "Error: There should be 4 keys in url1")

	v2, _ := url2.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/")
	assert.Equal(t, len(v2.(ccurlcommon.TypeInterface).Value().([]string)), 4, "Error: There should be 4 keys in url2")
}

func TestNonce(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url1 := NewConcurrentUrl(store)
	if err := url1.Preload(ccurlcommon.SYSTEM, url1.Platform.Eth10(), "Alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	if err := url1.Write(0, "blcc://eth1.0/Alice/nonce", commutative.NewInt64(10, 100)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	if err := url1.Write(0, "blcc://eth1.0/Alice/nonce", commutative.NewInt64(10, 9)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	nonce, _ := url1.Read(0, "blcc://eth1.0/Alice/nonce")
	v := nonce.(ccurlcommon.TypeInterface)
	assert.Equal(t, v.Value(), int64(119), "")
	assert.Equal(t, v.Transitional(nil), int64(0), "Error: delta != 0")

	//	nonce doesn't seem to increment properly, need to find out
}
