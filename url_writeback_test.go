package concurrenturl

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	ccurltype "github.com/arcology-network/concurrenturl/type"
	commutative "github.com/arcology-network/concurrenturl/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/type/noncommutative"
)

func SimulatedTx0() []byte {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		fmt.Println(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/") // create a path
	url.Write(0, "blcc://eth1.0/Alice/storage/ctrn-0/", path)
	url.Write(0, "blcc://eth1.0/Alice/storage/ctrn-0/elem-00", noncommutative.NewString("tx0-elem-00")) /* The first Element */
	url.Write(0, "blcc://eth1.0/Alice/storage/ctrn-0/elem-01", noncommutative.NewString("tx0-elem-01")) /* The second Element */

	_, transitions := url.Export()
	return ccurltype.Univalues(transitions).Encode()
}

func SimulatedTx1() []byte {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		fmt.Println(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-1/") // create a path
	url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-1/", path)
	url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-1/elem-00", noncommutative.NewString("tx1-elem-00")) /* The first Element */
	url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-1/elem-01", noncommutative.NewString("tx1-elem-00")) /* The second Element */

	_, transitions := url.Export()
	return ccurltype.Univalues(transitions).Encode()
}

func CheckPaths(url *ConcurrentUrl) error {
	v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-0/elem-00")
	if v.(ccurlcommon.TypeInterface).Value() == "tx0-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-0/elem-01")
	if v.(ccurlcommon.TypeInterface).Value() == "tx0-elem-01" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-1/elem-00")
	if v.(ccurlcommon.TypeInterface).Value() == "tx1-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-1/elem-01")
	if v.(ccurlcommon.TypeInterface).Value() == "tx1-elem-01" {
		return errors.New("Error: Not match")
	}

	//Read the path
	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Keys don't match !")
	}

	// Read the path again
	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Keys don't match !")
	}

	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-1/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Keys don't match !")
	}
	return nil
}

func TestStateUpdate(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}

	tx0bytes := SimulatedTx0()
	tx0Out := ccurltype.Univalues{}.Decode(tx0bytes, Decoder{}).(ccurltype.Univalues)

	tx1bytes := SimulatedTx1()
	tx1Out := ccurltype.Univalues{}.Decode(tx1bytes, Decoder{}).(ccurltype.Univalues)

	url.indexer.Import(tx0Out)
	url.indexer.Import(tx1Out)

	errs := url.indexer.Commit([]uint32{0, 1})
	if len(errs) != 0 || CheckPaths(url) != nil {
		t.Error(errs)
	}

	/* The second round */
	// _, transitions := url.Export()
	// bytes := ccurltype.Univalues(transitions).Encode()
	// outTrans := ccurltype.Univalues{}.Decode(bytes, Decoder{}).(ccurltype.Univalues)
	// url.indexer.Import(outTrans)

	// errs = url.indexer.Commit([]uint32{0, 1})
	// if len(errs) != 0 || CheckPaths(url) != nil {
	// 	t.Error(errs)
	// }

	// Delete an nonexistent entry, should fail !
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0", nil); err == nil {
		t.Error("Error: Writing Should fail !")
	}

	v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		t.Error("Error: Keys don't match !")
	}

	// Delete the container-0
	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/", nil); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	v, _ = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-0/")
	if v != nil {
		t.Error("Error: The element should be gone already !")
	}
	_, transitions := url.Export()
	/* */

	url.indexer.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(transitions).Encode(), Decoder{}).(ccurltype.Univalues))
	errs = url.indexer.Commit([]uint32{0, 1})

	// if len(errs) != 0 {
	// 	t.Error(errs)
	// }

	if v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-0/"); v != nil {
		t.Error("Error: Should be gone already !")
	}

	//url.indexer.Store().Print()
}

func TestMultipleTxStateUpdate(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		t.Error(err)
	}

	tx0bytes := SimulatedTx0()
	tx0Out := ccurltype.Univalues{}.Decode(tx0bytes, Decoder{}).(ccurltype.Univalues)

	tx1bytes := SimulatedTx1()
	tx1Out := ccurltype.Univalues{}.Decode(tx1bytes, Decoder{}).(ccurltype.Univalues)

	url.indexer.Import(tx0Out)
	url.indexer.Import(tx1Out)

	errs := url.indexer.Commit([]uint32{0, 1})
	if len(errs) != 0 {
		t.Error(errs)
	}

	/* Check Paths */
	CheckPaths(url)

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-111", noncommutative.NewString("tx0-elem-111")); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	if err := url.Write(1, "blcc://eth1.0/Alice/storage/ctrn-0/elem-222", noncommutative.NewString("tx0-elem-222")); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	_, transitions := url.Export()
	url.indexer.Import(transitions)
	if errs = url.indexer.Commit([]uint32{1}); len(errs) > 0 {
		t.Error(errs)
	}

	v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/ctrn-0/")
	keys := v.(ccurlcommon.TypeInterface).Value()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01", "elem-111", "elem-222"}) {
		t.Error("Error: Keys don't match !")
	}
	//url.indexer.Store().Print()
}

func TestAccessControl(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	tx0bytes := SimulatedTx0()
	tx0Out := ccurltype.Univalues{}.Decode(tx0bytes, Decoder{}).(ccurltype.Univalues)

	tx1bytes := SimulatedTx1()
	tx1Out := ccurltype.Univalues{}.Decode(tx1bytes, Decoder{}).(ccurltype.Univalues)

	url.indexer.Import(tx0Out)
	url.indexer.Import(tx1Out)

	errs := url.indexer.Commit([]uint32{0, 1})
	if len(errs) != 0 {
		t.Error(errs)
	}

	// Account root Path
	v, err := url.Read(1, "blcc://eth1.0/Alice/")
	if v == nil {
		t.Error(err) // Users shouldn't be able to read any of the system paths
	}

	v, err = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/")
	if v == nil {
		t.Error(err)
	}

	/* Code */
	v, err = url.Read(1, "blcc://eth1.0/Alice/code")
	if v != nil { // Should be able to read
		t.Error(err)
	}

	err = url.Write(1, "blcc://eth1.0/Alice/code", noncommutative.NewString("New code"))
	if err == nil {
		t.Error("Error: Users shouldn't be updated blcc://eth1.0/Alice/code")
	}

	/* Balance */
	v, err = url.Read(1, "blcc://eth1.0/Alice/balance")
	if v != nil {
		t.Error(err)
	}

	err = url.Write(1, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(100), big.NewInt(100)))
	if err != nil {
		t.Error("Error: Failed to write the balance")
	}

	err = url.Write(1, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(100), big.NewInt(0)))
	if err != nil {
		t.Error("Error: Failed to initialize balance")
	}

	err = url.Write(1, "blcc://eth1.0/Alice/balance", commutative.NewBalance(big.NewInt(100), big.NewInt(100)))
	if err != nil {
		t.Error("Error: Failed to initialize balance")
	}

	v, _ = url.Read(1, "blcc://eth1.0/Alice/balance")
	if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(300)) != 0 {
		t.Error("Error: blcc://eth1.0/Alice/balance, should be 200")
	}

	/* Nonce */
	err = url.Write(1, "blcc://eth1.0/Alice/nonce", noncommutative.NewInt64(10))
	if err != nil {
		t.Error("Error: Failed to read the nonce value !")
	}

	/* Storage */
	meta, _ := commutative.NewMeta("")
	err = url.Write(1, "blcc://eth1.0/Alice/storage/", meta)
	if err == nil {
		t.Error("Error: Users shouldn't be able to change the storage path !")
	}

	_, err = url.Read(1, "blcc://eth1.0/Alice/storage/")
	if err != nil {
		t.Error("Error: Failed to read the storage !")
	}

	err = url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/", meta)
	if err != nil {
		t.Error("Error: The system should be able to change the storage path !")
	}

	_, err = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/Alice/storage/")
	if err != nil {
		t.Error(err)
	}
}
