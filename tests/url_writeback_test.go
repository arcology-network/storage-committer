package ccurltest

import (
	"fmt"
	"reflect"
	"testing"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/univalue"
	"github.com/holiman/uint256"
)

func TestAddAndDelete(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	path := commutative.NewPath()
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	acctTrans = univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{1})

	url.Init(store)
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil) // Delete an non-existing entry, should NOT appear in the transitions

	if acctTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...); len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionSameBatch(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)
	// _, addPath := url.Export(ccurlcommon.Sorter)
	addPath := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(addPath).Encode()).(univalue.Univalues))
	// url.Import(url.Decode(univalue.Univalues(addPath).Encode()))
	url.PostImport()
	url.Commit([]uint32{1})

	url.Init(store)
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewInt64(1))
	// _, addTrans := url.Export(ccurlcommon.Sorter)
	addTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	// url.Import(url.Decode(univalue.Univalues(addTrans).Encode()))
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(addTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{1})

	if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1"); v == nil {
		t.Error("Error: Failed to read the key !")
	}

	url2 := ccurl.NewConcurrentUrl(store)
	if v, _ := url2.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1"); v.(int64) != 1 {
		t.Error("Error: Failed to read the key !")
	}

	_ = url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", nil)
	// _, deleteTrans := url2.Export(ccurlcommon.Sorter)
	deleteTrans := univalue.Univalues(common.Clone(url2.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	if v, _ := url2.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1"); v != nil {
		t.Error("Error: Failed to read the key !")
	}

	url3 := ccurl.NewConcurrentUrl(store)
	url3.Import(append(addTrans, deleteTrans...))
	url3.PostImport()
	url3.Commit([]uint32{1, 2})

	if v, _ := url3.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1"); v != nil {
		t.Error("Error: Failed to delete the entry !")
	}
}

func TestApplyingTransitionsFromMulitpleBatches(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	url.Import(acctTrans)
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	path := commutative.NewPath()
	err := url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/containers/ctrn-0/", path)

	if err != nil {
		t.Error("error")
	}

	acctTrans = univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{1})

	url.Init(store)
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil)

	if acctTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...); len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionDifferentBatch(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	in := univalue.Univalues(acctTrans).Encode()
	out := univalue.Univalues{}.Decode(in).(univalue.Univalues)
	// url.Import(url.Decode(univalue.Univalues(out).Encode()))
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(out).Encode()).(univalue.Univalues))

	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("1"))
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("2"))

	acctTrans = univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	in = univalue.Univalues(acctTrans).Encode()
	out = univalue.Univalues{}.Decode(in).(univalue.Univalues)
	// url.Import(url.Decode(univalue.Univalues(out).Encode()))
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(out).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{1})

	url.Init(store)
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("3"))
	_ = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("4"))

	path, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	if reflect.DeepEqual(path.([]string), []string{"1", "2", "3", "4"}) {
		t.Error("Error: Not match")
	}

	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil) // delete the path
	if acctTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...); len(acctTrans) != 3 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestStateUpdate(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}
	// _, initTrans := url.Export(ccurlcommon.Sorter)
	initTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	// url.Import(url.Decode(univalue.Univalues(initTrans).Encode()))
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(initTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	tx0bytes, err := SimulatedTx0(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx0Out := univalue.Univalues{}.Decode(tx0bytes).(univalue.Univalues)

	tx1bytes, err := SimulatedTx1(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx1Out := univalue.Univalues{}.Decode(tx1bytes).(univalue.Univalues)

	// url.Import(url.Decode(univalue.Univalues(append(tx0Out, tx1Out...)).Encode()))
	url.Import((append(tx0Out, tx1Out...)))
	url.PostImport()
	errs := url.Commit([]uint32{0, 1})
	if len(errs) != 0 {
		t.Error(errs)
	}

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
	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil {
		t.Error("Error: Cann't delete a path twice !")
	}

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); v != nil {
		t.Error("Error: The path should be gone already !")
	}

	transitions := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	out := univalue.Univalues{}.Decode(univalue.Univalues(transitions).Encode()).(univalue.Univalues)

	url.Import(out)
	url.PostImport()
	errs = url.Commit([]uint32{1})
	for _, err := range errs {
		t.Error(err)
	}

	if v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"); v != nil {
		t.Error("Error: Should be gone already !")
	}
}

func TestMultipleTxStateUpdate(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	// _, initTrans := url.Export(ccurlcommon.Sorter)
	initTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(initTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	tx0bytes, err := SimulatedTx0(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx0Out := univalue.Univalues{}.Decode(tx0bytes).(univalue.Univalues)

	tx1bytes, err := SimulatedTx1(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx1Out := univalue.Univalues{}.Decode(tx1bytes).(univalue.Univalues)

	// url.Import(url.Decode(univalue.Univalues(append(tx0Out, tx1Out...)).Encode()))
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(append(tx0Out, tx1Out...)).Encode()).(univalue.Univalues))
	url.PostImport()

	errs := url.Commit([]uint32{0, 1})
	if len(errs) != 0 {
		t.Error(errs)
	}

	// url.Init(store)
	CheckPaths(alice, url) /* Check Paths */

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-111", noncommutative.NewString("tx0-elem-111")); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-222", noncommutative.NewString("tx0-elem-222")); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-111")
	if v.(string) != "tx0-elem-111" {
		t.Error("Error: Failed to delete the path !")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-222")
	if v.(string) != "tx0-elem-222" {
		t.Error("Error: Failed to delete the path !")
	}

	// 	transitions := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	// url.Import(univalue.Univalues{}.Decode(univalue.Univalues(transitions).Encode()).(univalue.Univalues))

	// url.PostImport()
	// url.Commit([]uint32{1})

	// url.Init(store)
	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	keys := v.([]string)
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01", "elem-111", "elem-222"}) {
		t.Error("Error: Keys don't match !")
	}
	// url.Importer().Store().Print()
}

func TestAccessControl(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	// _, initTrans := url.Export(ccurlcommon.Sorter)
	initTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	// url.Import(url.Decode(univalue.Univalues(initTrans).Encode()))
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(initTrans).Encode()).(univalue.Univalues))

	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	tx0bytes, err := SimulatedTx0(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx0Out := univalue.Univalues{}.Decode(tx0bytes).(univalue.Univalues)

	tx1bytes, err := SimulatedTx1(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx1Out := univalue.Univalues{}.Decode(tx1bytes).(univalue.Univalues)

	// url.Import(url.Decode(univalue.Univalues(append(tx0Out, tx1Out...)).Encode()))
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(append(tx0Out, tx1Out...)).Encode()).(univalue.Univalues))

	url.PostImport()
	url.Commit([]uint32{0, 1})

	url.Init(store)
	// Account root Path
	v, err := url.Read(1, "blcc://eth1.0/account/"+alice+"/")
	if v == nil {
		t.Error(err) // Users shouldn't be able to read any of the system paths
	}

	v, err = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/")
	if v == nil {
		t.Error(err)
	}

	/* Code */
	v, err = url.Read(1, "blcc://eth1.0/account/"+alice+"/code")
	if err != nil { // Shouldn't be able to read
		t.Error(err)
	}

	// err = url.Write(1, "blcc://eth1.0/account/"+alice+"/code", noncommutative.NewString("New code"))
	// if err == nil {
	// 	t.Error("Error: Users shouldn't be updated blcc://eth1.0/account/alice/code")
	// }

	/* U256 */
	v, err = url.Read(1, "blcc://eth1.0/account/"+alice+"/balance")
	if err != nil {
		t.Error(err)
	}

	err = url.Write(1, "blcc://eth1.0/account/"+alice+"/balance",
		commutative.NewU256Delta(uint256.NewInt(200), true))
	if err != nil {
		t.Error("Error: Failed to write the balance")
	}

	err = url.Write(1, "blcc://eth1.0/account/"+alice+"/balance",
		commutative.NewU256Delta(uint256.NewInt(100), true))
	if err != nil {
		t.Error("Error: Failed to initialize balance")
	}

	err = url.Write(1, "blcc://eth1.0/account/"+alice+"/balance",
		commutative.NewU256Delta(uint256.NewInt(100), false))
	if err != nil {
		t.Error("Error: Failed to initialize balance")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+alice+"/balance")
	if v.(*uint256.Int).Cmp(uint256.NewInt(200)) != 0 {
		t.Error("Error: blcc://eth1.0/account/alice/balance, should be 200 not ", v.(*commutative.U256).Value().(*uint256.Int).ToBig().Uint64())
	}

	/* Nonce */
	err = url.Write(1, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64(0, 10))
	if err != nil {
		t.Error("Error: Failed to read the nonce value !")
	}

	/* Storage */
	meta := commutative.NewPath()
	// err = url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/", meta)
	// if err == nil {
	// 	t.Error("Error: Users shouldn't be able to change the storage path !")
	// }

	_, err = url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/")
	if err != nil {
		t.Error("Error: Failed to read the storage !")
	}

	err = url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/", meta) //this should preexists
	if err == nil {
		t.Error("Error: The system should be able to change the storage path !")
	}

	_, err = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/")
	if err != nil {
		t.Error(err)
	}
}

func TestArray(t *testing.T) {
	array := []interface{}{1, 2, 3, 4}

	fun := func(arr []interface{}) []interface{} {
		arr[0] = nil
		arr[1] = nil
		return arr
	}

	fun(array)

	fmt.Println(array)

}
