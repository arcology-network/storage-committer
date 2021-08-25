package ccurltest

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	ccurl "github.com/arcology/concurrenturl/v2"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
	ccurltype "github.com/arcology/concurrenturl/v2/type"
	commutative "github.com/arcology/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology/concurrenturl/v2/type/noncommutative"
)

func TestAddAndDelete(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	_, acctTrans := url.Export(false)
	url.Commit(acctTrans, []uint32{ccurlcommon.SYSTEM})

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path)

	_, acctTrans = url.Export(false)
	url.Commit(acctTrans, []uint32{1})

	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/4", nil)

	if _, acctTrans := url.Export(false); len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionSameBatch(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	_, acctTrans := url.Export(false)
	url.Commit(acctTrans, []uint32{ccurlcommon.SYSTEM})

	// create a path
	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path)
	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/1", noncommutative.NewString("1"))
	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/2", noncommutative.NewString("2"))

	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", nil) // delete the path
	if _, acctTrans := url.Export(false); len(acctTrans) != 1 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionDifferentBatch(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	_, acctTrans := url.Export(false)
	url.Commit(acctTrans, []uint32{ccurlcommon.SYSTEM})

	// create a path
	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path)
	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/1", noncommutative.NewString("1"))
	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/2", noncommutative.NewString("2"))

	_, acctTrans = url.Export(false)
	url.Commit(acctTrans, []uint32{1})

	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/1", noncommutative.NewString("3"))
	_ = url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/2", noncommutative.NewString("4"))

	path, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	if reflect.DeepEqual(path.(*commutative.Meta).GetKeys(), []string{"1", "2", "3", "4"}) {
		t.Error("Error: Not match")
	}

	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", nil) // delete the path
	if _, acctTrans := url.Export(false); len(acctTrans) != 4 {
		t.Error("Error: Wrong number of transitions")
	}
}

func SimulatedTx0() []byte {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/", path)
	url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-00", noncommutative.NewString("tx0-elem-00")) /* The first Element */
	url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-01", noncommutative.NewString("tx0-elem-01")) /* The second Element */

	_, transitions := url.Export(true)
	return ccurltype.Univalues(transitions).Encode()
}

func SimulatedTx1() []byte {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-1/") // create a path
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-1/", path)
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-1/elem-00", noncommutative.NewString("tx1-elem-00")) /* The first Element */
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-1/elem-01", noncommutative.NewString("tx1-elem-00")) /* The second Element */

	_, transitions := url.Export(true)
	return ccurltype.Univalues(transitions).Encode()
}

func CheckPaths(url *ccurl.ConcurrentUrl) error {
	v, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-00")
	if v.(ccurlcommon.TypeInterface).Value() == "tx0-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-01")
	if v.(ccurlcommon.TypeInterface).Value() == "tx0-elem-01" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-1/elem-00")
	if v.(ccurlcommon.TypeInterface).Value() == "tx1-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-1/elem-01")
	if v.(ccurlcommon.TypeInterface).Value() == "tx1-elem-01" {
		return errors.New("Error: Not match")
	}

	//Read the path
	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Keys don't match !")
	}

	// Read the path again
	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Keys don't match !")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-1/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Keys don't match !")
	}
	return nil
}

func TestStateUpdate(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	tx0bytes := SimulatedTx0()
	tx0Out := ccurltype.Univalues{}.Decode(tx0bytes).(ccurltype.Univalues)

	tx1bytes := SimulatedTx1()
	tx1Out := ccurltype.Univalues{}.Decode(tx1bytes).(ccurltype.Univalues)

	url.Indexer().Import(tx0Out)
	url.Indexer().Import(tx1Out)

	_, acctTrans := url.Export(false)
	url.Commit(acctTrans, []uint32{0, 1})

	// Delete an nonexistent entry, should fail !
	if err := url.Write(9, "blcc://eth1.0/account/alice/storage/ctrn-0", nil); err == nil {
		t.Error("Error: Writing Should fail !")
	}

	v, _ := url.Read(9, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		t.Error("Error: Keys don't match !")
	}

	// Delete the container-0
	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", nil); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	if v, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/"); v != nil {
		t.Error("Error: The path should be gone already !")
	}

	_, transitions := url.Export(true)
	out := ccurltype.Univalues{}.Decode(ccurltype.Univalues(transitions).Encode()).(ccurltype.Univalues)
	errs := url.Commit(out, []uint32{1})

	for _, err := range errs {
		t.Error(err)
	}

	if v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/storage/ctrn-0/"); v != nil {
		t.Error("Error: Should be gone already !")
	}
}

func TestMultipleTxStateUpdate(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	tx0bytes := SimulatedTx0()
	tx0Out := ccurltype.Univalues{}.Decode(tx0bytes).(ccurltype.Univalues)

	tx1bytes := SimulatedTx1()
	tx1Out := ccurltype.Univalues{}.Decode(tx1bytes).(ccurltype.Univalues)

	errs := url.Commit(append(tx0Out, tx1Out...), []uint32{0, 1})
	if len(errs) != 0 {
		t.Error(errs)
	}

	/* Check Paths */
	CheckPaths(url)

	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-111", noncommutative.NewString("tx0-elem-111")); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-222", noncommutative.NewString("tx0-elem-222")); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	_, transitions := url.Export(true)
	url.Commit(transitions, []uint32{1})

	v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	keys := v.(ccurlcommon.TypeInterface).Value()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01", "elem-111", "elem-222"}) {
		t.Error("Error: Keys don't match !")
	}
	//url.Indexer().Store().Print()
}

func TestAccessControl(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	tx0bytes := SimulatedTx0()
	tx0Out := ccurltype.Univalues{}.Decode(tx0bytes).(ccurltype.Univalues)

	tx1bytes := SimulatedTx1()
	tx1Out := ccurltype.Univalues{}.Decode(tx1bytes).(ccurltype.Univalues)

	url.Indexer().Import(tx0Out)
	url.Indexer().Import(tx1Out)

	url.Commit(append(tx0Out, tx1Out...), []uint32{0, 1})

	// Account root Path
	v, err := url.Read(1, "blcc://eth1.0/account/alice/")
	if v == nil {
		t.Error(err) // Users shouldn't be able to read any of the system paths
	}

	v, err = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/")
	if v == nil {
		t.Error(err)
	}

	/* Code */
	v, err = url.Read(1, "blcc://eth1.0/account/alice/code")
	if err != nil { // Shouldn't be able to read
		t.Error(err)
	}

	err = url.Write(1, "blcc://eth1.0/account/alice/code", noncommutative.NewString("New code"))
	if err == nil {
		t.Error("Error: Users shouldn't be updated blcc://eth1.0/account/alice/code")
	}

	/* Balance */
	v, err = url.Read(1, "blcc://eth1.0/account/alice/balance")
	if err != nil {
		t.Error(err)
	}

	err = url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(100), big.NewInt(100)))
	if err != nil {
		t.Error("Error: Failed to write the balance")
	}

	err = url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(100), big.NewInt(0)))
	if err != nil {
		t.Error("Error: Failed to initialize balance")
	}

	err = url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(100), big.NewInt(100)))
	if err != nil {
		t.Error("Error: Failed to initialize balance")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/balance")
	if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(200)) != 0 {
		t.Error("Error: blcc://eth1.0/account/alice/balance, should be 300")
	}

	/* Nonce */
	err = url.Write(1, "blcc://eth1.0/account/alice/nonce", commutative.NewInt64(0, 10))
	if err != nil {
		t.Error("Error: Failed to read the nonce value !")
	}

	/* Storage */
	meta, _ := commutative.NewMeta("")
	err = url.Write(1, "blcc://eth1.0/account/alice/storage/", meta)
	if err == nil {
		t.Error("Error: Users shouldn't be able to change the storage path !")
	}

	_, err = url.Read(1, "blcc://eth1.0/account/alice/storage/")
	if err != nil {
		t.Error("Error: Failed to read the storage !")
	}

	err = url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/storage/", meta)
	if err != nil {
		t.Error("Error: The system should be able to change the storage path !")
	}

	_, err = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/storage/")
	if err != nil {
		t.Error(err)
	}
}
