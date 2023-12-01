package ccurltest

// import (
// 	"reflect"
// 	"testing"

// 	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
// 	"github.com/arcology-network/common-lib/common"
// 	ccurl "github.com/arcology-network/concurrenturl"
// 	ccurlcommon "github.com/arcology-network/concurrenturl/common"
// 	indexer "github.com/arcology-network/concurrenturl/indexer"
// 	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
// 	storage "github.com/arcology-network/concurrenturl/storage"
// )

// func TestPartialCache(t *testing.T) {
// 	memDB := cachedstorage.NewMemDB()
// 	policy := cachedstorage.NewCachePolicy(10000000, 1.0)
// 	store := cachedstorage.NewDataStore(nil, policy, memDB, storage.Codec{}.Encode, storage.Codec{}.Decode)
// 	url := ccurl.NewConcurrentUrl(store)
// 	alice := AliceAccount()
// 	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
// 	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
// 	url.Sort()
// 	url.Commit([]uint32{ccurlcommon.SYSTEM})

// 	/* Filter persistent data source */
// 	excludeMemDB := func(db cachedstorage.PersistentStorageInterface) bool { // Do not access MemDB
// 		name := reflect.TypeOf(db).String()
// 		return name != "*cachedstorage.MemDB"
// 	}

// 	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
// 	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
// 	url.Importer().Store().(*cachedstorage.DataStore).Cache().Clear()
// 	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues), true, excludeMemDB) // The changes will be discarded.
// 	url.Sort()
// 	url.Commit([]uint32{1})

// 	// if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 	// 	t.Error("Error: The entry shouldn't be in the DB !")
// 	// } else {
// 	// 	if string(*(v.(*noncommutative.String))) != "1234" {
// 	// 		t.Error("Error: The entry shouldn't changed !")
// 	// 	}
// 	// }

// 	/* Don't filter persistent data source	*/
// 	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
// 	url.Importer().Store().(*cachedstorage.DataStore).Cache().Clear()                                 // Make sure only the persistent storage has the data.
// 	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues)) // This should take effect
// 	url.Sort()
// 	url.Commit([]uint32{1})

// 	if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 		t.Error("Error: The entry shouldn't be in the DB !")
// 	} else {
// 		if v.(string) != "9999" {
// 			t.Error("Error: The entry should have been changed !")
// 		}
// 	}
// }

// func TestPartialCacheWithFilter(t *testing.T) {
// 	memDB := cachedstorage.NewMemDB()
// 	policy := cachedstorage.NewCachePolicy(10000000, 1.0)

// 	excludeMemDB := func(db cachedstorage.PersistentStorageInterface) bool { /* Filter persistent data source */
// 		name := reflect.TypeOf(db).String()
// 		return name == "*cachedstorage.MemDB"
// 	}

// 	store := cachedstorage.NewDataStore(nil, policy, memDB, storage.Codec{}.Encode, storage.Codec{}.Decode, excludeMemDB)
// 	url := ccurl.NewConcurrentUrl(store)
// 	alice := AliceAccount()
// 	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
// 	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
// 	url.Sort()
// 	url.Commit([]uint32{ccurlcommon.SYSTEM})

// 	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999")); err != nil {
// 		t.Error(err)
// 	}

// 	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

// 	// url := ccurl.NewConcurrentUrl(store)

// 	url.WriteCache().Clear()

// 	// ccmap2 := url.Importer().Store().(*cachedstorage.DataStore).Cache()
// 	// fmt.Print(ccmap2)
// 	out := indexer.Univalues{}.Decode(indexer.Univalues(common.Clone(acctTrans)).Encode()).(indexer.Univalues)
// 	url.Import(out, true, excludeMemDB) // The changes will be discarded.
// 	url.Sort()
// 	url.Commit([]uint32{1})

// 	if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 		t.Error("Error: The entry shouldn't be in the DB as the persistent DB has been excluded !")
// 	} else {
// 		if v.(string) != "9999" {
// 			t.Error("Error: The entry shouldn't changed !")
// 		}
// 	}
// }
