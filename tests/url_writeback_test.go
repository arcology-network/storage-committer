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

func TestAddAndDelete(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.NewAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true)

	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	url.Init(store)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil, true) // Delete an non-existing entry, should NOT appear in the transitions

	raw := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	if acctTrans := raw; len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionSameBatch(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.NewAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true)
	// _, addPath := url.Export(indexer.Sorter)
	addPath := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(addPath).Encode()).(indexer.Univalues))
	// url.Import(url.Decode(indexer.Univalues(addPath).Encode()))
	url.Sort()
	url.Commit([]uint32{1})

	url.Init(store)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewInt64(1), true)
	// _, addTrans := url.Export(indexer.Sorter)
	addTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	// url.Import(url.Decode(indexer.Univalues(addTrans).Encode()))
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(addTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1"); v == nil {
		t.Error("Error: Failed to read the key !")
	}

	url2 := ccurl.NewConcurrentUrl(store)
	if v, _ := url2.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1"); v.(int64) != 1 {
		t.Error("Error: Failed to read the key !")
	}

	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", nil, true)
	// _, deleteTrans := url2.Export(indexer.Sorter)
	deleteTrans := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	if v, _ := url2.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1"); v != nil {
		t.Error("Error: Failed to read the key !")
	}

	url3 := ccurl.NewConcurrentUrl(store)
	url3.Import(append(addTrans, deleteTrans...))
	url3.Sort()
	url3.Commit([]uint32{1, 2})

	if v, _ := url3.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1"); v != nil {
		t.Error("Error: Failed to delete the entry !")
	}
}

func TestApplyingTransitionsFromMulitpleBatches(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.NewAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(acctTrans)
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	path := commutative.NewPath()
	_, err := url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path, true)

	if err != nil {
		t.Error("error")
	}

	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	url.Init(store)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil, true)

	if acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{}); len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionDifferentBatch(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.NewAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	in := indexer.Univalues(acctTrans).Encode()
	out := indexer.Univalues{}.Decode(in).(indexer.Univalues)
	// url.Import(url.Decode(indexer.Univalues(out).Encode()))
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(out).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("1"), true)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("2"), true)

	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	in = indexer.Univalues(acctTrans).Encode()
	out = indexer.Univalues{}.Decode(in).(indexer.Univalues)
	// url.Import(url.Decode(indexer.Univalues(out).Encode()))
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(out).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	url.Init(store)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("3"), true)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("4"), true)

	path, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	if reflect.DeepEqual(path.([]string), []string{"1", "2", "3", "4"}) {
		t.Error("Error: Not match")
	}

	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil, true) // delete the path
	if acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{}); len(acctTrans) != 3 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestStateUpdate(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.NewAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	// _, initTrans := url.Export(indexer.Sorter)
	initTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	// url.Import(url.Decode(indexer.Univalues(initTrans).Encode()))
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(initTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	tx0bytes, trans, err := Create_Ctrn_0(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx0Out := indexer.Univalues{}.Decode(tx0bytes).(indexer.Univalues)
	tx0Out = trans
	tx1bytes, err := Create_Ctrn_1(alice, store)
	if err != nil {
		t.Error(err)
	}

	tx1Out := indexer.Univalues{}.Decode(tx1bytes).(indexer.Univalues)

	// url.Import(url.Decode(indexer.Univalues(append(tx0Out, tx1Out...)).Encode()))
	url.Import((append(tx0Out, tx1Out...)))
	url.Sort()
	url.Commit([]uint32{0, 1})
	//need to encode delta only now it encodes everything

	if err := CheckPaths(alice, url); err != nil {
		t.Error(err)
	}

	v, _ := url.Read(9, "blcc://eth1.0/account/"+alice+"/storage/") //system doesn't generate sub paths for /storage/
	// if v.(*commutative.Path).CommittedLength() != 2 {
	// 	t.Error("Error: Wrong sub paths")
	// }

	// if !reflect.DeepEqual(v.([]string), []string{"ctrn-0/", "ctrn-1/"}) {
	// 	t.Error("Error: Didn't find the subpath!")
	// }

	v, _ = url.Read(9, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	keys := v.([]string)
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		t.Error("Error: Keys don't match !")
	}

	// Delete the container-0
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil, true); err != nil {
		t.Error("Error: Cann't delete a path twice !")
	}

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); v != nil {
		t.Error("Error: The path should be gone already !")
	}

	transitions := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	out := indexer.Univalues{}.Decode(indexer.Univalues(transitions).Encode()).(indexer.Univalues)

	url.Import(out)
	url.Sort()
	url.Commit([]uint32{1})

	if v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); v != nil {
		t.Error("Error: Should be gone already !")
	}
}

// func TestMultipleTxStateUpdate(t *testing.T) {
// 	store := cachedstorage.NewDataStore()
// 	url := ccurl.NewConcurrentUrl(store)
// 	alice := datacompression.RandomAccount()
// 	if err := url.NewAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	// _, initTrans := url.Export(indexer.Sorter)
// 	initTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

// 	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(initTrans).Encode()).(indexer.Univalues))
// 	url.Sort()
// 	url.Commit([]uint32{ccurlcommon.SYSTEM})

// 	url.Init(store)
// 	tx0bytes, err := Create_Ctrn_0(alice, store)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	tx0Out := indexer.Univalues{}.Decode(tx0bytes).(indexer.Univalues)

// 	tx1bytes, err := Create_Ctrn_1(alice, store)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	tx1Out := indexer.Univalues{}.Decode(tx1bytes).(indexer.Univalues)

// 	// url.Import(url.Decode(indexer.Univalues(append(tx0Out, tx1Out...)).Encode()))
// 	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(append(tx0Out, tx1Out...)).Encode()).(indexer.Univalues))
// 	url.Sort()

// 	url.Commit([]uint32{0, 1})

// 	// url.Init(store)
// 	CheckPaths(alice, url) /* Check Paths */

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-111", noncommutative.NewString("tx0-elem-111")); err != nil {
// 		t.Error("Error: Failed to delete the path !")
// 	}

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-222", noncommutative.NewString("tx0-elem-222")); err != nil {
// 		t.Error("Error: Failed to delete the path !")
// 	}

// 	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-111")
// 	if v.(string) != "tx0-elem-111" {
// 		t.Error("Error: Failed to delete the path !")
// 	}

// 	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-222")
// 	if v.(string) != "tx0-elem-222" {
// 		t.Error("Error: Failed to delete the path !")
// 	}

// 	// 	transitions := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
// 	// url.Import(indexer.Univalues{}.Decode(indexer.Univalues(transitions).Encode()).(indexer.Univalues))

// 	// url.Sort()
// 	// url.Commit([]uint32{1})

// 	// url.Init(store)
// 	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
// 	keys := v.([]string)
// 	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01", "elem-111", "elem-222"}) {
// 		t.Error("Error: Keys don't match !")
// 	}
// 	// url.Importer().Store().Print()
// }

// func TestAccessControl(t *testing.T) {
// 	store := cachedstorage.NewDataStore()
// 	url := ccurl.NewConcurrentUrl(store)
// 	alice := datacompression.RandomAccount()
// 	if err := url.NewAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	// _, initTrans := url.Export(indexer.Sorter)
// 	initTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

// 	// url.Import(url.Decode(indexer.Univalues(initTrans).Encode()))
// 	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(initTrans).Encode()).(indexer.Univalues))

// 	url.Sort()
// 	url.Commit([]uint32{ccurlcommon.SYSTEM})

// 	url.Init(store)
// 	tx0bytes, err := Create_Ctrn_0(alice, store)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	tx0Out := indexer.Univalues{}.Decode(tx0bytes).(indexer.Univalues)

// 	tx1bytes, err := Create_Ctrn_1(alice, store)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	tx1Out := indexer.Univalues{}.Decode(tx1bytes).(indexer.Univalues)

// 	// url.Import(url.Decode(indexer.Univalues(append(tx0Out, tx1Out...)).Encode()))
// 	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(append(tx0Out, tx1Out...)).Encode()).(indexer.Univalues))

// 	url.Sort()
// 	url.Commit([]uint32{0, 1})

// 	url.Init(store)
// 	// Account root Path
// 	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/")
// 	if v == nil {
// 		t.Error(err) // Users shouldn't be able to read any of the system paths
// 	}

// 	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/")
// 	if v == nil {
// 		t.Error(err)
// 	}

// 	/* Code */
// 	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/code")
// 	if v != nil { // Shouldn't be able to read
// 		t.Error(err)
// 	}

// 	// err = url.Write(1, "blcc://eth1.0/account/"+alice+"/code", noncommutative.NewString("New code"))
// 	// if err == nil {
// 	// 	t.Error("Error: Users shouldn't be updated blcc://eth1.0/account/alice/code")
// 	// }

// 	/* U256 */
// 	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/balance")
// 	if v != nil {
// 		t.Error(err)
// 	}

// 	_, err = url.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(200), true))
// 	if err != nil {
// 		t.Error("Error: Failed to write the balance")
// 	}

// 	_, err = url.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(100), true))
// 	if err != nil {
// 		t.Error("Error: Failed to initialize balance")
// 	}

// 	_, err = url.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(100), false))
// 	if err != nil {
// 		t.Error("Error: Failed to initialize balance")
// 	}

// 	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/balance")
// 	if v.(*uint256.Int).Cmp(uint256.NewInt(200)) != 0 {
// 		t.Error("Error: blcc://eth1.0/account/alice/balance, should be 200 not ", v.(*commutative.U256).Value().(*uint256.Int).ToBig().Uint64())
// 	}

// 	/* Nonce */
// 	_, err = url.Write(1, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64(0, 10))
// 	if err != nil {
// 		t.Error("Error: Failed to read the nonce value !")
// 	}

// 	/* Storage */
// 	meta := commutative.NewPath()
// 	// err = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/", meta)
// 	// if err == nil {
// 	// 	t.Error("Error: Users shouldn't be able to change the storage path !")
// 	// }

// 	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/")
// 	if v == nil {
// 		t.Error("Error: Failed to read the data !")
// 	}

// 	v, _ = url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", meta) //this should preexists
// 	if v == nil {
// 		t.Error("Error: The system should be able to change the storage path !")
// 	}

// 	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/")
// 	if v != nil {
// 		t.Error(err)
// 	}
// }

// func TestArray(t *testing.T) {
// 	array := []interface{}{1, 2, 3, 4}

// 	fun := func(arr []interface{}) []interface{} {
// 		arr[0] = nil
// 		arr[1] = nil
// 		return arr
// 	}

// 	fun(array)

// 	fmt.Println(array)

// }
