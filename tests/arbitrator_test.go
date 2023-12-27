package ccurltest

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	datacompression "github.com/arcology-network/common-lib/addrcompressor"
	common "github.com/arcology-network/common-lib/common"
	ccurl "github.com/arcology-network/concurrenturl"
	arbitrator "github.com/arcology-network/concurrenturl/arbitrator"
	committercommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/univalue"
	cache "github.com/arcology-network/eu/cache"
)

func TestArbiCreateTwoAccountsNoConflict(t *testing.T) {
	store := chooseDataStore()

	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	meta := commutative.NewPath()
	writeCache.Write(committercommon.SYSTEM, committercommon.ETH10_ACCOUNT_PREFIX, meta)
	trans := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url := ccurl.NewStorageCommitter(store)
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{committercommon.SYSTEM})
	writeCache.Clear()

	alice := AliceAccount()
	// url.Init(store)
	// url.NewAccount(1, alice) // NewAccount account structure {
	if _, err := writeCache.CreateNewAccount(1, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// accesses1, transitions1 := writeCache.Export(indexer.Sorter)
	accesses1 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	writeCache.Clear()
	bob := datacompression.RandomAccount()
	if _, err := writeCache.CreateNewAccount(2, bob); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	accesses2 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	arib := (&arbitrator.Arbitrator{})

	IDVec := append(common.Fill(make([]uint32, len(accesses1)), 0), common.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := arib.Detect(IDVec, append(accesses1, accesses2...))

	conflictdict, _, _ := arbitrator.Conflicts(ids).ToDict()
	if len(*conflictdict) != 0 {
		t.Error("Error: There should be NO conflict")
	}
}

func TestArbiCreateTwoAccounts1Conflict(t *testing.T) {
	store := chooseDataStore()

	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())
	meta := commutative.NewPath()
	writeCache.Write(committercommon.SYSTEM, committercommon.ETH10_ACCOUNT_PREFIX, meta)
	trans := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url := ccurl.NewStorageCommitter(store)
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{committercommon.SYSTEM})

	url.Init(store)
	alice := AliceAccount()
	// url.NewAccount(1, alice) // NewAccount account structure {

	writeCache.Clear()                                               // = url.WriteCache()
	if _, err := writeCache.CreateNewAccount(1, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	path1 := commutative.NewPath()                                                // create a path
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path1) // create a path
	// writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	// writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, _ := writeCache.Export(indexer.Sorter)
	raw := writeCache.Export(indexer.Sorter)
	accesses1 := indexer.Univalues(common.Clone(raw)).To(indexer.IPCTransition{})

	// url2 := ccurl.NewStorageCommitter(store)
	writeCache.Clear()                                               // = url2.WriteCache()
	if _, err := writeCache.CreateNewAccount(2, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	} // NewAccount account structure {

	// writeCache = url2.WriteCache()
	if _, err := writeCache.CreateNewAccount(1, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	path2 := commutative.NewPath() // create a path
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	// url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	// url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	// accesses2, _ := url2.WriteCache().Export(indexer.Sorter)
	accesses2 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCAccess{})

	// accesses1.Print()
	// fmt.Print(" ++++++++++++++++++++++++++++++++++++++++++++++++ ")
	// accesses2.Print()

	IDVec := append(common.Fill(make([]uint32, len(accesses1)), 0), common.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses1, accesses2...))
	conflictdict, _, _ := arbitrator.Conflicts(ids).ToDict()

	if len(*conflictdict) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}
}

func TestArbiTwoTxModifyTheSameAccount(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())
	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// writeCache.Write(committercommon.SYSTEM, committercommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())
	acctTrans := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{committercommon.SYSTEM})
	url.Init(store)

	// url.NewAccount(1, alice)
	writeCache.Clear()
	if _, err := writeCache.CreateNewAccount(1, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/", commutative.NewPath()) // create a path
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, transitions1 := writeCache.Export(indexer.Sorter)
	accesses1 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	transitions1 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	// url2 := ccurl.NewStorageCommitter(store)
	// writeCache = url2.WriteCache()
	writeCache.Clear()
	if _, err := writeCache.CreateNewAccount(2, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	} // NewAccount account structure {
	path2 := commutative.NewPath() // create a path

	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/", path2)
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))

	// accesses2, transitions2 := url2.WriteCache().Export(indexer.Sorter)
	accesses2 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	transitions2 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	IDVec := append(common.Fill(make([]uint32, len(accesses1)), 0), common.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses1, accesses2...))
	conflictDict, _, pairs := arbitrator.Conflicts(ids).ToDict()

	// pairs := arbitrator.Conflicts(ids).ToPairs()

	if len(*conflictDict) != 1 || len(pairs) != 1 {
		t.Error("Error: There should be 1 conflict")
	}

	toCommit := common.Exclude([]uint32{1, 2}, common.MapKeys(*conflictDict))

	in := indexer.Univalues(append(transitions1, transitions2...)).Encode()
	out := indexer.Univalues{}.Decode(in).(indexer.Univalues)
	url.Import(out)
	url.Sort()
	url.Commit(toCommit)
	writeCache.Clear()

	// url3 := ccurl.NewStorageCommitter(store)
	if _, err := writeCache.Write(3, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("url3-1-by-tx-3")); err != nil {
		t.Error(err)
	}

	// accesses3, transitions3 := url3.Export(indexer.Sorter)
	accesses3 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	transitions3 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.IPCTransition{})

	writeCache.Clear()
	// url4 := ccurl.NewStorageCommitter(store)
	if _, err := writeCache.Write(4, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("url4-1-by-tx-3")); err != nil {
		t.Error(err)
	}
	// accesses4, transitions4 := url4.Export(indexer.Sorter)
	accesses4 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	transitions4 := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	IDVec = append(common.Fill(make([]uint32, len(accesses3)), 0), common.Fill(make([]uint32, len(accesses4)), 1)...)
	ids = (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses3, accesses4...))
	conflictDict, _, _ = arbitrator.Conflicts(ids).ToDict()

	conflictTx := common.MapKeys(*conflictDict)
	if len(*conflictDict) != 1 || conflictTx[0] != 4 {
		t.Error("Error: There should be only 1 conflict")
	}
	toCommit = common.Exclude([]uint32{3, 4}, conflictTx)

	in = indexer.Univalues(append(transitions3, transitions4...)).Encode()
	out = indexer.Univalues{}.Decode(in).(indexer.Univalues)

	acctTrans = append(transitions3, transitions4...)
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))

	// url.Import(url.Decode(indexer.Univalues(append(transitions3, transitions4...)).Encode()))
	url.Sort()
	url.Commit(toCommit)
	url.Init(store)

	writeCache.Clear()
	v, _, _ := writeCache.Read(3, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", new(noncommutative.String))
	if v == nil || v.(string) != "url3-1-by-tx-3" {
		t.Error("Error: Wrong value, expecting:", "url3-1-by-tx-3 ", "actual:", v)
	}

	// have to mark balance and nonce persistent first !!!!!

	// v, _ = url3.Read(3, "blcc://eth1.0/account/"+alice+"/nonce", new(commutative.Uint64))
	// if v == nil || v.(uint64) != 2 {
	// 	t.Error("Error: Wrong value, expecting:", "2", "actual:", v)
	// }
}

func BenchmarkSimpleArbitrator(b *testing.B) {
	alice := AliceAccount()
	univalues := make([]interfaces.Univalue, 0, 5*200000)
	groupIDs := make([]uint32, 0, len(univalues))

	v := commutative.NewPath()
	for i := 0; i < len(univalues)/5; i++ {
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))

		groupIDs = append(groupIDs, uint32(i))
		groupIDs = append(groupIDs, uint32(i))
		groupIDs = append(groupIDs, uint32(i))
		groupIDs = append(groupIDs, uint32(i))
		groupIDs = append(groupIDs, uint32(i))
	}

	t0 := time.Now()
	(&arbitrator.Arbitrator{}).Detect(groupIDs, univalues)
	fmt.Println("Detect "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))
}
