package concurrenturl

import (
	"math/big"
	"reflect"
	"testing"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	ccurltype "github.com/arcology-network/concurrenturl/type"
	urltype "github.com/arcology-network/concurrenturl/type"
	commutative "github.com/arcology-network/concurrenturl/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/type/noncommutative"
)

func TestBasic(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}
	// (*url.indexer.Store()).Print()
	// if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
	// 	t.Error(err)
	// }

	(*url.indexer.Store()).Print()

	// create a path
	path, err := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Error: Failed to create the path")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", "path"); err == nil {
		t.Error(err)
	}
	(*url.indexer.Store()).Print()

	// Try to read an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-1"); value != nil {
		t.Error("Error: Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-1/elem-000"); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	//try again
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Error: Shouldn't be not found")
	}
	(*url.indexer.Store()).Print()
	// try to read an nonexistent path
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Error: Shouldn't be not found")
	}

	// Write the entry
	if value := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); value != nil {
		t.Error("Error: Shouldn't be not found")
	}
	(*url.indexer.Store()).Print()

	// Read the entry back
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); (*value.(*noncommutative.Int64)) != 1111 {
		t.Error("Error: Shouldn't be not found")
	}

	// Read the path
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/"); value == nil {
		t.Error(value)
	} else {
		if !reflect.DeepEqual(value.(*commutative.Meta).GetKeys(), []string{"elem-000"}) {
			t.Error("Error: Wrong value ")
		}
	}

	_, transitions := url.Export()
	(*url.indexer.Store()).Print()

	if !reflect.DeepEqual(transitions[6].GetValue().(*commutative.Meta).GetKeys(), []string{"elem-000"}) {
		t.Error("Error: keys don't match")
	}

	value := transitions[7].GetValue()
	if *(value.(*noncommutative.Int64)) != 1111 {
		t.Error("Error: keys don't match")
	}

	// wrong condition, value should still exists
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/"); value == nil {
		t.Error("Error: The variable has been cleared")
	}

	url.indexer.Import(transitions)
	if errs := url.indexer.Commit([]uint32{1}); len(errs) != 0 {
		t.Error(errs)
	}

	/* =========== The second cycle ==============*/

	//try reading an element written in the previous cycle
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); value == nil {
		t.Error("Error: Entry not found")
	}
}

func TestUrl1(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}
	path, err := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Error: Failed to create the path")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Write an entry having the the same name of a path, should go through
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0", noncommutative.NewString("ctrn-0")); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/elem-0", noncommutative.NewString("elem-0")); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000", noncommutative.NewInt64(0)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000", noncommutative.NewInt64(5555)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-001", noncommutative.NewInt64(6666)); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-002", noncommutative.NewInt64(7777)); err != nil {
		t.Error(err)
	}

	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); value == nil || (*value.(*noncommutative.Int64)) != 5555 {
		t.Error("Error: Wrong value")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-001"); value == nil || (*value.(*noncommutative.Int64)) != 6666 {
		t.Error("Error: Wrong value")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-002"); value == nil || (*value.(*noncommutative.Int64)) != 7777 {
		t.Error("Error: Wrong value")
	}

	if meta, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/"); meta == nil {
		t.Error("Error: not found")
	}

	// Export all access records and state transitions
	_, transitions := url.Export()
	if (*transitions[6].GetValue().(*noncommutative.String)) != "ctrn-0" {
		t.Error("Error: keys don't match")
	}

	if !reflect.DeepEqual(transitions[7].GetValue().(*commutative.Meta).GetKeys(), []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	if meta, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/"); meta == nil {
		t.Error("Error: The variable has been cleared")
	}
}

func TestUrl2(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}

	// Create a new container
	path, err := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Error: Failed to create the path")
	}
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", path); err != nil {
		t.Error(err, "Error:  Failed to MakePath: "+"/ctrn-0/")
	}

	// Add a vaiable directly
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/elem-0", noncommutative.NewString("0000")); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/elem-0")
	}

	// Add the first element
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error(err, "Error: Failed to Write: "+"/ctrn-0/elem-000")
	}

	// Add the second element
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-001")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-002", noncommutative.NewInt64(3333)); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-0/elem-002")
	}

	// Write to an nonexistent path, shouldn't succeed, but should leave a couple of access records
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-1/elem-002", noncommutative.NewInt64(3333)); err == nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-1/elem-002")
	}

	// Read an nonexistent path, shouldn't succeed
	if err, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-1/elem-002"); err != nil {
		t.Error(err, "Error:  Failed to Write: "+"/ctrn-1/elem-002")
	}

	// Add the first element
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); value == nil || (*value.(*noncommutative.Int64)) != 1111 {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-000")
	}

	// Try to read an nonexistent element, should leave a access record
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-005"); value != nil {
		t.Error("Error: Failed to Read: " + "/ctrn-0/elem-005")
	}

	// Update then return path meta info
	meta0, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if !reflect.DeepEqual(meta0.(*commutative.Meta).GetKeys(), []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Do again
	meta1, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if !reflect.DeepEqual(meta1.(*commutative.Meta).GetKeys(), []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: Keys don't match")
	}

	// Two queries should match
	if !reflect.DeepEqual(meta0, meta1) {
		t.Error("Error: Wrong meta info")
	}

	// Delete elem-00
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000", nil); err != nil {
		t.Error("Error: Failed to write: " + "/ctrn-0/elem-000")
	}

	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); value != nil {
		t.Error("Error: The element wasn't successfully deleted")
	}

	// The elem-00 has been deleted, only "elem-001", "elem-002" left
	meta, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if !reflect.DeepEqual(meta.(*commutative.Meta).GetKeys(), []string{"elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	// Readd elem-00 back
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000", noncommutative.NewInt64(9999)); err != nil { // delete
		t.Error("Error: Failed to write: " + "/ctrn-0/elem-000")
	}

	// Check elem-00's value
	if value, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000"); (*value.(*noncommutative.Int64)) != 9999 {
		t.Error("Error: The element wasn't successfully deleted")
	}

	// Update then read the path info again
	meta, _ = url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if !reflect.DeepEqual(meta.(*commutative.Meta).GetKeys(), []string{"elem-000", "elem-001", "elem-002"}) {
		t.Error("Error: keys don't match")
	}

	v, _ := url.Read(1, "blcc://eth1.0/Alice/storage/elem-0")
	if v == nil {
		t.Error("Error: keys don't match")
	}

	/* Remove the path and all the elements underneath */
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", nil); err != nil {
		t.Error(err, "Failed to remove path: "+"/ctrn-0/")
	}

	/* The path should be gone by now */
	v, _ = url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if v != nil {
		t.Error("Error: keys don't match")
	}

	/*  */
	v, _ = url.Read(1, "blcc://eth1.0/Alice/storage/elem-0")
	if v == nil {
		t.Error("Error: keys don't match")
	}

	/*  Read the storage path to see what is left*/
	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/")
	if !reflect.DeepEqual(v.(*commutative.Meta).GetKeys(), []string{"containers/", "elem-0", "native/"}) {
		t.Error("Error: keys don't match")
	}

	/*  Export all */
	accessRecords, transitions := url.Export()
	if transitions[7].GetPath() != "blcc://eth1.0/Alice/storage/elem-0" ||
		*(transitions[7].GetValue().(*noncommutative.String)) != "0000" {
		t.Error("Error: Transitions don't match")
	}

	/* ------------- Check access records first  -------------
	blcc://eth1.0/Alice/storage/
	blcc://eth1.0/Alice/storage/ctrn-0/
	blcc://eth1.0/Alice/storage/elem-0
	blcc://eth1.0/Alice/storage/ctrn-0/elem-000
	blcc://eth1.0/Alice/storage/ctrn-0/elem-001
	blcc://eth1.0/Alice/storage/ctrn-0/elem-002
	blcc://eth1.0/Alice/storage/ctrn-1/elem-002
	blcc://eth1.0/Alice/storage/ctrn-0/elem-005
	blcc://eth1.0/Alice/storage/ctrn-1/elem-002
	*/

	condition := ccurltype.NewUnivalue(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-000", 3, 4, nil)
	if !ccurltype.Univalues(accessRecords).IfContains(condition) {
		t.Error("Error: Error: ")
	}

	condition = ccurltype.NewUnivalue(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-001", 0, 2, nil)
	if !ccurltype.Univalues(accessRecords).IfContains(condition) {
		t.Error("Error: Error: ")
	}

	condition = ccurltype.NewUnivalue(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-002", 0, 2, nil)
	if !ccurltype.Univalues(accessRecords).IfContains(condition) {
		t.Error("Error: Error: ")
	}

	condition = ccurltype.NewUnivalue(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-005", 1, 0, nil)
	if !ccurltype.Univalues(accessRecords).IfContains(condition) {
		t.Error("Error: Error: ")
	}

	// Encode then Decode access records
	in := ccurltype.Univalues(transitions).Encode()

	codec := Decoder{}
	out := ccurltype.Univalues{}.Decode(in, codec).(ccurltype.Univalues)
	for i := range transitions {
		if !transitions[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Error: Accesses don't match")
		}
	}

	// Encode then Decode state transitions
	in = ccurltype.Univalues(transitions).Encode()
	out = ccurltype.Univalues{}.Decode(in, codec).(ccurltype.Univalues)
	// if len(out) != 2 {
	// 	t.Error("Error: Wrong transition count")
	// }

	if out[3].GetPath() != "blcc://eth1.0/Alice/storage/" ||
		reflect.DeepEqual(out[3].GetValue(), []string{"containers/", "elem-0", "native/"}) {
		t.Error("Error: Transitions don't match after decoding")
	}

	if out[7].GetPath() != "blcc://eth1.0/Alice/storage/elem-0" ||
		*out[7].GetValue().(*noncommutative.String) != "0000" {
		t.Error("Error: Transitions don't match after decoding")
	}
}

func TestUrl3(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}

	path, err := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Error: Failed to create the path")
	}
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", path); err != nil {
		t.Error(err, " Failed to MakePath: "+"/ctrn-0/")
	}

	inV := noncommutative.NewBigint(123456)
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-0", inV); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	v, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-0")
	value := (*big.Int)(inV.(*noncommutative.Bigint))
	outV := (*big.Int)(v.(*noncommutative.Bigint))
	if outV.Cmp(value) != 0 {
		t.Error("Error: Bigint values don't match")
	}

	accessRecords, _ := url.Export()
	in := ccurltype.Univalues(accessRecords).Encode()
	out := ccurltype.Univalues{}.Decode(in, &Decoder{}).(ccurltype.Univalues)
	for i := range accessRecords {
		if !accessRecords[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Error: Accesses don't match")
		}
	}
}

func TestCommutative(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}

	// create a path
	path, err := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Error: Failed to create the path: blcc://eth1.0/Alice/storage/ctrn-0/")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", path); err != nil {
		t.Error(err, " Failed to MakePath: blcc://eth1.0/Alice/storage/ctrn-0/")
	}

	// create a noncommutative bigint
	inV := noncommutative.NewBigint(100)
	value := (*big.Int)(inV.(*noncommutative.Bigint))
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-0", inV); err != nil {
		t.Error(err, " Failed to Write: blcc://eth1.0/Alice/storage/ctrn-0/elem-0")
	}

	v, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-0")
	outV := (*big.Int)(v.(*noncommutative.Bigint))
	if outV.Cmp(value) != 0 {
		t.Error("Error: Failed to read: blcc://eth1.0/Alice/storage/ctrn-0/elem-0")
	}

	// -------------------Create a commutative bigint ------------------------------
	comtVInit := commutative.NewBalance(big.NewInt(300), big.NewInt(0))
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/comt-0", comtVInit); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/comt-0", commutative.NewBalance(big.NewInt(300), big.NewInt(1))); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/comt-0", commutative.NewBalance(big.NewInt(300), big.NewInt(2))); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	v, _ = url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/comt-0")
	if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(303)) != 0 {
		t.Error("Error: comt-0 has a wrong returned value")
	}

	// ----------------------------Balance ---------------------------------------------------
	if err := url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(0))); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	// Add the first delta
	if err := url.Write(1, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(22))); err != nil {
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	// Add the second delta
	if err := url.Write(1, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(200), big.NewInt(11))); err != nil {
		t.Error(err, "blcc://eth1.0/Alice/balance")
	}

	// Read Alice's balance
	v, _ = url.Read(1, "blcc://eth1.0/Alice/balance")
	if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(233)) != 0 {
		t.Error("Error: blcc://eth1.0/Alice/balance")
	}

	// Export variables
	accessRecords, transitions := url.Export()
	in := ccurltype.Univalues(accessRecords).Encode()
	out := ccurltype.Univalues{}.Decode(in, &Decoder{}).(ccurltype.Univalues)
	for i := range accessRecords {
		if !accessRecords[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Error: Accesses don't match")
		}
	}

	in = ccurltype.Univalues(transitions).Encode()
	out = ccurltype.Univalues{}.Decode(in, &Decoder{}).(ccurltype.Univalues)
	for i := range transitions {
		if !transitions[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Error: Accesses don't match")
		}
	}

	url.indexer.Import(transitions)
	url.indexer.Commit([]uint32{0, 1})
	//url.indexer.Store().Print()
}

func TestNestedPath(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}

	// create a path
	path, err := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/")
	if err != nil {
		t.Error("Error: Failed to create the path")
	}
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// create a sub path
	path, err = commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/ctrn-00/")
	if err != nil {
		t.Error("Error: Failed to create the path")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/ctrn-00/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-00", noncommutative.NewString("elem-00")); err != nil {
		t.Error(err)
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-01", noncommutative.NewInt64(1234)); err != nil {
		t.Error(err)
	}

	// The first element !
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/ctrn-00/elem-00", noncommutative.NewString("elem-00")); err != nil {
		t.Error(err)
	}

	// The second element
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/ctrn-00/elem-01", noncommutative.NewString("elem-01")); err != nil {
		t.Error(err)
	}

	/* Read */
	v, _ := url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/ctrn-00/")
	if !reflect.DeepEqual(v.(*commutative.Meta).GetKeys(), []string{"elem-00", "elem-01"}) {
		t.Error("Error: keys don't match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(*commutative.Meta).GetKeys(), []string{"ctrn-00/", "elem-00", "elem-01"}) {
		t.Error("Error: keys don't match")
	}

	if !reflect.DeepEqual(v.(*commutative.Meta).Removed(), []string{}) {
		t.Error("Error: keys don't match")
	}

	/* Remove the path */
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", nil); err != nil {
		t.Error(err)
	}

	/*Try reading again */
	v, _ = url.Read(1, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if v != nil {
		t.Error("Error: Error:  Path should been deleted already !")
	}

	accessRecords, transitions := url.Export()
	in := ccurltype.Univalues(accessRecords).Encode()
	out := ccurltype.Univalues{}.Decode(in, Decoder{}).(ccurltype.Univalues)
	for i := range accessRecords {
		if !accessRecords[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Error: Accesses don't match")
		}
	}

	in = ccurltype.Univalues(transitions).Encode()
	out = ccurltype.Univalues{}.Decode(in, Decoder{}).(ccurltype.Univalues)
	for i := range transitions {
		if !transitions[i].(*urltype.Univalue).EqualAccess(out[i].(*urltype.Univalue)) {
			t.Error("Error: Accesses don't match")
		}
	}

	//url.indexer.Store().Print()
}
