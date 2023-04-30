package ccurltest

import (
	"reflect"
	"testing"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl/v2"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	noncommutative "github.com/arcology-network/concurrenturl/v2/noncommutative"
	storage "github.com/arcology-network/concurrenturl/v2/storage"
	univalue "github.com/arcology-network/concurrenturl/v2/univalue"
)

func TestPartialCache(t *testing.T) {
	memDB := cachedstorage.NewMemDB()
	policy := cachedstorage.NewCachePolicy(10000000, 1.0)
	store := cachedstorage.NewDataStore(nil, policy, memDB, storage.ToBytes, storage.FromBytes)
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
	_, acctTrans := url.Export(ccurlcommon.Sorter)
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	/* Filter persistent data source */
	excludeMemDB := func(db cachedstorage.PersistentStorageInterface) bool { // Do not access MemDB
		name := reflect.TypeOf(db).String()
		return name != "*cachedstorage.MemDB"
	}

	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
	_, acctTrans = url.Export(ccurlcommon.Sorter)
	(*url.Store()).(*cachedstorage.DataStore).LocalCache().Clear()
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues), true, excludeMemDB) // The changes will be discarded.
	url.PostImport()
	url.Commit([]uint32{1})

	// if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
	// 	t.Error("Error: The entry shouldn't be in the DB !")
	// } else {
	// 	if string(*(v.(*noncommutative.String))) != "1234" {
	// 		t.Error("Error: The entry shouldn't changed !")
	// 	}
	// }

	/* Don't filter persistent data source	*/
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
	(*url.Store()).(*cachedstorage.DataStore).LocalCache().Clear()                                       // Make sure only the persistent storage has the data.
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues)) // This should take effect
	url.PostImport()
	url.Commit([]uint32{1})

	if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
		t.Error("Error: The entry shouldn't be in the DB !")
	} else {
		if string(*(v.(*noncommutative.String))) != "9999" {
			t.Error("Error: The entry should have been changed !")
		}
	}
}

func TestPartialCacheWithFilter(t *testing.T) {
	memDB := cachedstorage.NewMemDB()
	policy := cachedstorage.NewCachePolicy(10000000, 1.0)

	excludeMemDB := func(db cachedstorage.PersistentStorageInterface) bool { /* Filter persistent data source */
		name := reflect.TypeOf(db).String()
		return name == "*cachedstorage.MemDB"
	}

	store := cachedstorage.NewDataStore(nil, policy, memDB, storage.ToBytes, storage.FromBytes, excludeMemDB)
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
	_, acctTrans := url.Export(ccurlcommon.Sorter)
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	if err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999")); err != nil {
		t.Error(err)
	}

	_, acctTrans = url.Export(ccurlcommon.Sorter)
	(*url.Store()).(*cachedstorage.DataStore).LocalCache().Clear()
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues), true, excludeMemDB) // The changes will be discarded.
	url.PostImport()
	url.Commit([]uint32{1})

	if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
		t.Error("Error: The entry shouldn't be in the DB as the persistent DB has been excluded !")
	} else {
		if string(*(v.(*noncommutative.String))) != "9999" {
			t.Error("Error: The entry shouldn't changed !")
		}
	}
}
