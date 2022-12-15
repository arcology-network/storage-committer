package ccurltest

import (
	"reflect"
	"testing"

	cachedstorage "github.com/arcology/common-lib/cachedstorage"
	datacompression "github.com/arcology/common-lib/datacompression"
	ccurl "github.com/arcology/concurrenturl/v2"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
	ccurltype "github.com/arcology/concurrenturl/v2/type"
	noncommutative "github.com/arcology/concurrenturl/v2/type/noncommutative"
)

func TestPartialCache(t *testing.T) {
	memDB := cachedstorage.NewMemDB()
	policy := cachedstorage.NewCachePolicy(10000000, 1.0)
	store := cachedstorage.NewDataStore(nil, policy, memDB, ccurltype.ToBytes, ccurltype.FromBytes)
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
	_, acctTrans := url.Export(true)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	/* Filter persistent data source */
	excludeMemDB := func(db cachedstorage.PersistentStorageInterface) bool { // Do not access MemDB
		name := reflect.TypeOf(db).String()
		return name != "*cachedstorage.MemDB"
	}

	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
	_, acctTrans = url.Export(true)
	(*url.Store()).(*cachedstorage.DataStore).LocalCache().Clear()
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues), true, excludeMemDB) // The changes will be discarded.
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
	(*url.Store()).(*cachedstorage.DataStore).LocalCache().Clear()                                          // Make sure only the persistent storage has the data.
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues)) // This should take effect
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

	store := cachedstorage.NewDataStore(nil, policy, memDB, ccurltype.ToBytes, ccurltype.FromBytes, excludeMemDB)
	url := ccurl.NewConcurrentUrl(store)
	alice := datacompression.RandomAccount()
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
	_, acctTrans := url.Export(true)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
	_, acctTrans = url.Export(true)
	(*url.Store()).(*cachedstorage.DataStore).LocalCache().Clear()
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(acctTrans).Encode()).(ccurltype.Univalues), true, excludeMemDB) // The changes will be discarded.
	url.PostImport()
	url.Commit([]uint32{1})

	if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
		t.Error("Error: The entry shouldn't be in the DB as the persistent DB has been excluded !")
	} else {
		if string(*(v.(*noncommutative.String))) != "1234" {
			t.Error("Error: The entry shouldn't changed !")
		}
	}
}
