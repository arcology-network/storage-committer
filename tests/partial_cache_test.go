package committertest

// import (
// 	"reflect"
// 	"testing"

// 	storage "github.com/arcology-network/common-lib/storage"
// 	"github.com/arcology-network/common-lib/common"
// 	stgcomm "github.com/arcology-network/storage-committer"
// 	stgcommcommon "github.com/arcology-network/storage-committer/common"
// 	importer "github.com/arcology-network/storage-committer/importer"
// 	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
// 	storage "github.com/arcology-network/storage-committer/storage"
// )

// func TestPartialCache(t *testing.T) {
// 	memDB := storage.NewMemoryDB()
// 	policy := storage.NewCachePolicy(10000000, 1.0)
// 	store := storage.NewDataStore(nil, policy, memDB, platform.Codec{}.Encode, platform.Codec{}.Decode)
// 		committer := stgcommitter.NewStorageCommitter(store)
// writeCache := committer.WriteCache()
// 	alice := AliceAccount()
// 	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	committer.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
// 	committer.Sort()
// 	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
// committer.Commit()

// 	/* Filter persistent data source */
// 	excludeMemDB := func(db storage.PersistentStorageInterface) bool { // Do not access MemDB
// 		name := reflect.TypeOf(db).String()
// 		return name != "*storage.MemDB"
// 	}

// 	committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
// 	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
// 	committer.Importer().Store().(*storage.DataStore).Cache().Clear()
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues), true, excludeMemDB) // The changes will be discarded.
// 	committer.Sort()
// 	committer.Precommit([]uint32{1})
// committer.Commit()

// 	// if v, _ := committer.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 	// 	t.Error("Error: The entry shouldn't be in the DB !")
// 	// } else {
// 	// 	if string(*(v.(*noncommutative.String))) != "1234" {
// 	// 		t.Error("Error: The entry shouldn't changed !")
// 	// 	}
// 	// }

// 	/* Don't filter persistent data source	*/
// 	committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
// 	committer.Importer().Store().(*storage.DataStore).Cache().Clear()                                 // Make sure only the persistent storage has the data.
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues)) // This should take effect
// 	committer.Sort()
// 	committer.Precommit([]uint32{1})
// committer.Commit()

// 	if v, _ := committer.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 		t.Error("Error: The entry shouldn't be in the DB !")
// 	} else {
// 		if v.(string) != "9999" {
// 			t.Error("Error: The entry should have been changed !")
// 		}
// 	}
// }

// func TestPartialCacheWithFilter(t *testing.T) {
// 	memDB := storage.NewMemoryDB()
// 	policy := storage.NewCachePolicy(10000000, 1.0)

// 	excludeMemDB := func(db storage.PersistentStorageInterface) bool { /* Filter persistent data source */
// 		name := reflect.TypeOf(db).String()
// 		return name == "*storage.MemDB"
// 	}

// 	store := storage.NewDataStore(nil, policy, memDB, platform.Codec{}.Encode, platform.Codec{}.Decode, excludeMemDB)
// 		committer := stgcommitter.NewStorageCommitter(store)
// writeCache := committer.WriteCache()
// 	alice := AliceAccount()
// 	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	committer.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
// 	committer.Sort()
// 	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
// committer.Commit()

// 	if _, err := committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999")); err != nil {
// 		t.Error(err)
// 	}

// 	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

// 	// 	committer := stgcommitter.NewStorageCommitter(store)
// writeCache := committer.WriteCache()

// 	committer.WriteCache().Clear()

// 	// ccmap2 := committer.Importer().Store().(*storage.DataStore).Cache()
// 	// fmt.Print(ccmap2)
// 	out := univalue.Univalues{}.Decode(univalue.Univalues(slice.Clone(acctTrans)).Encode()).(univalue.Univalues)
// 	committer.Import(out, true, excludeMemDB) // The changes will be discarded.
// 	committer.Sort()
// 	committer.Precommit([]uint32{1})
// committer.Commit()

// 	if v, _ := committer.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 		t.Error("Error: The entry shouldn't be in the DB as the persistent DB has been excluded !")
// 	} else {
// 		if v.(string) != "9999" {
// 			t.Error("Error: The entry shouldn't changed !")
// 		}
// 	}
// }
