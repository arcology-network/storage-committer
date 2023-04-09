package ccurltest

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"unsafe"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl/v2"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/type/noncommutative"
	"github.com/holiman/uint256"
)

func TestSize(t *testing.T) {
	// compressionLut := datacompression.NewCompressionLut()
	store := cachedstorage.NewDataStore()
	alice := datacompression.RandomAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	_, acctTrans := url.Export(true)
	buf := ccurltype.Univalues(acctTrans).Encode()
	url.Import(ccurltype.Univalues{}.Decode(buf).(ccurltype.Univalues))

	fmt.Println(ccurltype.Univalues(acctTrans).GetEncodedSize())
	fmt.Println(unsafe.Sizeof(map[string]int{}))

	original := []int{1, 2, 3, 4}
	original = append([]int{}, (original)...)
	fmt.Println(original)
	original[0] = 99
	fmt.Println(original)
	fmt.Println(original, "!!!")
}

func TestAddThenDeletePath(t *testing.T) {
	// compressionLut := datacompression.NewCompressionLut()
	store := cachedstorage.NewDataStore()
	alice := datacompression.RandomAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	_, acctTrans := url.Export(true)

	//values := ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).([]ccurlcommon.UnivalueInterface)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	_, transitions := url.Export(true)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(transitions).Encode()).(ccurltype.Univalues))
	url.PostImport()
	url.Commit([]uint32{1})

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	if v == nil {
		t.Error("Error: The path should have exists")
	}

	url.Init(store)
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil { // Delete the path
		t.Error(err)
	}

	_, acctTrans = url.Export(true)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues))
	url.PostImport()
	url.Commit([]uint32{1})

	// if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); v != nil {
	// 	t.Error("Error: The path should have been deleted")
	// }
}

func TestBasic(t *testing.T) {
	store := cachedstorage.NewDataStore()
	alice := datacompression.RandomAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	_, acctTrans := url.Export(true)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues))

	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", "path"); err == nil {
		t.Error(err)
	}

	// Try to read an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1"); value != nil {
		t.Error("Error: Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-000"); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	//try again
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// try to read an nonexistent path
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	// Write the entry
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	// Write the entry
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-111", noncommutative.NewInt64(9999)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-111")
	}

	// Read the entry back
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); (*value.(*noncommutative.Int64)) != 1111 {
		t.Error("Error: Shouldn't be not found")
	}

	// Read the path
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); value == nil {
		t.Error(value)
	} else {
		if !reflect.DeepEqual(value.(*commutative.Meta).Value().([]interface{}), []interface{}{"elem-000", "elem-111"}) {
			t.Error("Error: Wrong value !!!!")
		}
	}

	_, transitions := url.Export(true)

	if !reflect.DeepEqual(transitions[0].Value().(*commutative.Meta).Added(), []string{"elem-000", "elem-111"}) {
		t.Error("Error: keys are missing from the added buffer!")
	}

	value := transitions[1].Value()
	if *(value.(*noncommutative.Int64)) != 1111 {
		t.Error("Error: keys don't match")
	}

	// wrong condition, value should still exists
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); value == nil {
		t.Error("Error: The variable has been cleared !")
	}

	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(transitions).Encode()).(ccurltype.Univalues))
	// url.Import(url.Decode(ccurltype.Univalues(transitions).Encode()))
	url.PostImport()
	url.Commit([]uint32{1})

	/* =========== The second cycle ==============*/
	//try reading an element written in the previous cycle
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); value == nil {
		t.Error("Error: Entry not found")
	}
}

func TestUrl1(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		fmt.Println(err)
	}

	_, acctTrans := url.Export(true)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues))

	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	path, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Write an entry having the the same name of a path, should go through
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0", noncommutative.NewString("ctrn-0")); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("elem-0")); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(0)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(5555)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(6666)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(7777)); err != nil {
		t.Error(err)
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); value == nil || (*value.(*noncommutative.Int64)) != 5555 {
		t.Error("Error: Wrong value")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001"); value == nil || (*value.(*noncommutative.Int64)) != 6666 {
		t.Error("Error: Wrong value")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002"); value == nil || (*value.(*noncommutative.Int64)) != 7777 {
		t.Error("Error: Wrong value")
	}

	if meta, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); meta == nil {
		t.Error("Error: not found")
	}

	// Export all access records and state transitions
	_, transitions := url.Export(true)
	if (*transitions[0].Value().(*noncommutative.String)) != "ctrn-0" {
		t.Error("Error: keys don't match")
	}

	if !reflect.DeepEqual(ccurlcommon.SortString(transitions[1].Value().(*commutative.Meta).Added()), []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	if meta, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/"); meta == nil {
		t.Error("Error: The variable has been cleared")
	}
}

func TestUrl2(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	_, acctTrans := url.Export(true)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// Create a new container
	path, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err, "Error:  Failed to MakePath: "+"/ctrn-0/")
	}

	// Add a vaiable directly
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0", noncommutative.NewString("0000")); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/elem-0")
	}

	// Add the first element
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error(err, "Error: Failed to Write: "+"/ctrn-0/elem-000")
	}

	// Add the second element
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-001")
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-002")
	}

	// Write to an nonexistent path, will fail, but leave a couple of access records
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-002", noncommutative.NewInt64(3333)); err == nil {
		t.Error(err, "Error:    /ctrn-1/ does not exist, the Write should fail!!")
	}

	// Read an nonexistent path, shouldn't succeed
	if err, v := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-002"); v != nil {
		t.Error(err, "Error:  /ctrn-1/ does not exist, the read should fail!!")
	}

	// Add the first element
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); value == nil || (*value.(*noncommutative.Int64)) != 1111 {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-000")
	}

	// Try to read an nonexistent element, should leave a access record
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-005"); value != nil {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-005")
	}

	// Update then return path meta info
	meta0, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	if !reflect.DeepEqual(meta0.(*commutative.Meta).Value().([]interface{}), []interface{}{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Do again
	meta1, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	if !reflect.DeepEqual(meta1.(*commutative.Meta).Value().([]interface{}), []interface{}{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Two queries should match
	if !reflect.DeepEqual(meta0, meta1) {
		t.Error("Error: Wrong meta info")
	}

	// Delete elem-00
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", nil); err != nil {
		t.Error("Error: Failed to delete: " + "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-000")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Error: The element wasn't successfully deleted")
	}

	// Check elem-00's value
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001"); (*value.(*noncommutative.Int64)) != 2222 {
		t.Error("Error: The element wasn't found")
	}

	// The elem-00 has been deleted, only "elem-001", "elem-002" left
	meta, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	if !reflect.DeepEqual(meta.(*commutative.Meta).Value().([]interface{}), []interface{}{"elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	// Readd elem-00 back
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(9999)); err != nil { // delete
		t.Error("Error: Failed to write: " + "/ctrn-0/elem-000")
	}

	// Check elem-00's value
	if value, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"); (*value.(*noncommutative.Int64)) != 9999 {
		t.Error("Error: The element wasn't successfully deleted")
	}

	// Update then read the path info again
	meta, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	if !reflect.DeepEqual(meta.(*commutative.Meta).Value().([]interface{}), []interface{}{"elem-001", "elem-002", "elem-000"}) {
		t.Error("Error: keys don't match")
	}

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/elem-0")
	if v == nil {
		t.Error("Error: keys don't match")
	}

	/* Remove the path and all the elements underneath */
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil {
		t.Error(err, "Failed to remove path: "+"/ctrn-0/")
	}

	if v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); v != nil { /* The path should be gone by now */
		t.Error("Error: keys don't match")
	}

	if v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"); v != nil { /* all the sub paths should be gone as well*/
		t.Error("Error: keys don't match")
	}

	/*  Read the storage path to see what is left*/
	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/")
	if !reflect.DeepEqual(v.(*commutative.Meta).Value().([]interface{}), []interface{}{}) {
		t.Error("Error: Should be empty!!")
	}

	/*  Export all */
	accessRecords, transitions := url.Export(true)

	// 3 writes + 1 affiliated write
	value := ccurltype.NewUnivalue(ccurlcommon.VARIATE_TRANSITIONS, 1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, nil)
	if !ccurltype.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = ccurltype.NewUnivalue(ccurlcommon.VARIATE_TRANSITIONS, 1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-001", 1, 1, nil)
	if !ccurltype.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = ccurltype.NewUnivalue(ccurlcommon.VARIATE_TRANSITIONS, 1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-002", 0, 1, nil)
	if !ccurltype.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	value = ccurltype.NewUnivalue(ccurlcommon.VARIATE_TRANSITIONS, 1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-005", 1, 0, nil)
	if !ccurltype.Univalues(accessRecords).IfContains(value) {
		t.Error("Error: Error: ")
	}

	// Encode then Decode access records
	in := ccurltype.Univalues(transitions).Encode()

	out := ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	for i := range transitions {
		if !transitions[i].(*ccurltype.Univalue).Equal(out[i].(*ccurltype.Univalue)) {
			t.Error("Error: transitions don't match")
		}
	}

	// Encode then Decode state transitions
	in = ccurltype.Univalues(transitions).Encode()
	out = ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	// if len(out) != 2 {
	// 	t.Error("Error: Wrong transition count")
	// }

	// if out[0].GetPath() != "blcc://eth1.0/account/" + alice + "/storage/" ||
	// 	reflect.DeepEqual(out[0].Value(), []string{"elem-0"}) {
	// 	t.Error("Error: Transitions don't match after decoding")
	// }

	if *out[3].GetPath() != "blcc://eth1.0/account/"+alice+"/storage/elem-0" ||
		*(out[3].Value()).(*noncommutative.String) != "0000" {
		t.Error("Error: Wrong value!!!")
	}
}

func TestUrl3(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	// path, err := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	// if err != nil {
	// 	t.Error("Error: Failed to create the path")
	// }
	// if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
	// 	t.Error(err, " Failed to MakePath: "+"/ctrn-0/")
	// }

	// inV := noncommutative.NewBigint(123456)
	// if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", inV); err != nil {
	// 	t.Error(err, " Failed to Write: "+"/elem-0")
	// }

	// v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0")
	// value := (*big.Int)(inV.(*noncommutative.Bigint))
	// outV := (*big.Int)(v.(*noncommutative.Bigint))
	// if outV.Cmp(value) != 0 {
	// 	t.Error("Error: Bigint values don't match")
	// }

	accessRecords, _ := url.Export(true)
	in := ccurltype.Univalues(accessRecords).Encode()

	fmt.Println(len(in))
	out := ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	for i := range accessRecords {
		if !accessRecords[i].(*ccurltype.Univalue).Equal(out[i].(*ccurltype.Univalue)) {
			t.Error("Error: Accesses don't match")
		}
	}
}

func TestCommutative(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	// create a path
	path, err := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	if err != nil {
		t.Error("Error: Failed to create the path: blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err, " Failed to MakePath: blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	}

	// create a noncommutative bigint
	inV := noncommutative.NewBigint(100)
	value := (*big.Int)(inV.(*noncommutative.Bigint))
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", inV); err != nil {
		t.Error(err, " Failed to Write: blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0")
	}

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0")
	outV := (*big.Int)(v.(*noncommutative.Bigint))
	if outV.Cmp(value) != 0 {
		t.Error("Error: Failed to read: blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-0")
	}

	// -------------------Create a commutative UINT256 ------------------------------
	comtVInit := commutative.NewU256(uint256.NewInt(300), uint256.NewInt(0), commutative.U256MIN, commutative.U256MAX, commutative.ADDITION)
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", comtVInit); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0",
		commutative.NewU256(uint256.NewInt(300), uint256.NewInt(1), commutative.U256MIN, commutative.U256MAX, commutative.ADDITION)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0",
		commutative.NewU256(uint256.NewInt(300), uint256.NewInt(2), commutative.U256MIN, commutative.U256MAX, commutative.ADDITION)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0")
	if v.(*commutative.U256).Value().(*uint256.Int).Cmp(uint256.NewInt(303)) != 0 {
		t.Error("Error: comt-0 has a wrong returned value")
	}

	// ----------------------------U256 ---------------------------------------------------
	if err := url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/balance",
		commutative.NewU256(uint256.NewInt(0), uint256.NewInt(0), commutative.U256MIN, commutative.U256MAX, commutative.ADDITION)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the first delta
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/balance",
		commutative.NewU256(uint256.NewInt(0), uint256.NewInt(22), commutative.U256MIN, commutative.U256MAX, commutative.ADDITION)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the second delta
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/balance",
		commutative.NewU256(uint256.NewInt(0), uint256.NewInt(11), commutative.U256MIN, commutative.U256MAX, commutative.ADDITION)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Read alice's balance
	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/balance")
	if v.(*commutative.U256).Value().(*uint256.Int).Cmp(uint256.NewInt(33)) != 0 {
		t.Error("Error: blcc://eth1.0/account/alice/balance")
	}

	// Export variables
	accessRecords, transitions := url.Export(true)
	in := ccurltype.Univalues(accessRecords).Encode()
	out := ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	for i := range accessRecords {
		if !accessRecords[i].(*ccurltype.Univalue).Equal(out[i].(*ccurltype.Univalue)) {
			t.Error("Error: Accesses don't match")
		}
	}

	in = ccurltype.Univalues(transitions).Encode()
	out = ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	for i := range transitions {
		if !transitions[i].(*ccurltype.Univalue).Equal(out[i].(*ccurltype.Univalue)) {
			t.Error("Error: Transitions don't match !!!")
			fmt.Println(transitions[i])
			fmt.Println(out[i])
		}
	}

	url.Import(out)
	url.Commit([]uint32{0, 1})
}

func TestNestedPath(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	// create a path
	path, err := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	if err != nil {
		t.Error("Error: Failed to create the path")
	}
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// create a sub path
	path, err = commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/ctrn-00/")
	if err != nil {
		t.Error("Error: Failed to create the path")
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ctrn-00/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-00", noncommutative.NewString("elem-00")); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-01", noncommutative.NewInt64(1234)); err != nil {
		t.Error(err)
	}

	// The first element !
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ctrn-00/elem-00", noncommutative.NewString("elem-00")); err != nil {
		t.Error(err)
	}

	// The second element
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ctrn-00/elem-01", noncommutative.NewString("elem-01")); err != nil {
		t.Error(err)
	}

	/* Read */
	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/ctrn-00/")
	if !reflect.DeepEqual(v.(*commutative.Meta).Value().([]interface{}), []interface{}{"elem-00", "elem-01"}) {
		t.Error("Error: keys don't match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(*commutative.Meta).Value().([]interface{}), []interface{}{"ctrn-00/", "elem-00", "elem-01"}) {
		t.Error("Error: keys don't match")
	}

	if !reflect.DeepEqual(v.(*commutative.Meta).Removed(), []string{}) {
		t.Error("Error: keys don't match")
	}

	/* Remove the path */
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil {
		t.Error(err)
	}

	/*Try reading again */
	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	if v != nil {
		t.Error("Error: Error:  Path should been deleted already !")
	}

	accessRecords, transitions := url.Export(true)
	in := ccurltype.Univalues(accessRecords).Encode()
	out := ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	for i := range accessRecords {
		if !accessRecords[i].(*ccurltype.Univalue).Equal(out[i].(*ccurltype.Univalue)) {
			t.Error("Error: Accesses don't match  before and after encoding")
		}
	}

	in = ccurltype.Univalues(transitions).Encode()
	out = ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	for i := range transitions {
		if !transitions[i].(*ccurltype.Univalue).Equal(out[i].(*ccurltype.Univalue)) {
			t.Error("Error: Transitions don't match before and after encoding")
		}
	}

	url.Import(out)
	url.PostImport()
	url.Commit([]uint32{1})
}
