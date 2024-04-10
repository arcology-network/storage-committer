package committertest

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	arbitrator "github.com/arcology-network/storage-committer/arbitrator"
	importer "github.com/arcology-network/storage-committer/committer"
	stgcommitter "github.com/arcology-network/storage-committer/committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	platform "github.com/arcology-network/storage-committer/platform"
	cache "github.com/arcology-network/storage-committer/storage/writecache"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

func TestArbiCreateTwoAccountsNoConflict(t *testing.T) {
	store := chooseDataStore()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	meta := commutative.NewPath()
	writeCache.Write(stgcommcommon.SYSTEM, stgcommcommon.ETH10_ACCOUNT_PREFIX, meta)
	trans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)
	writeCache.Clear()

	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// accesses1, transitions1 := writeCache.Export(univalue.Sorter)
	accesses1 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITAccess{})
	univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITTransition{})

	writeCache.Clear()
	bob := BobAccount()
	if _, err := adaptorcommon.CreateNewAccount(2, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	accesses2 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITAccess{})
	univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITTransition{})

	arib := (&arbitrator.Arbitrator{})

	IDVec := append(slice.Fill(make([]uint32, len(accesses1)), 0), slice.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := arib.Detect(IDVec, append(accesses1, accesses2...))

	conflictdict, _, _ := arbitrator.Conflicts(ids).ToDict()
	if len(conflictdict) != 0 {
		t.Error("Error: There should be NO conflict")
	}
}

func TestArbiCreateTwoAccounts1Conflict(t *testing.T) {
	store := chooseDataStore()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	meta := commutative.NewPath()
	writeCache.Write(stgcommcommon.SYSTEM, stgcommcommon.ETH10_ACCOUNT_PREFIX, meta)
	trans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)

	committer.SetStore(store)
	alice := AliceAccount()

	writeCache.Clear()                                                              // = committer.WriteCache()
	if _, err := adaptorcommon.CreateNewAccount(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	path1 := commutative.NewPath()                                                // create a path
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path1) // create a path
	raw := writeCache.Export(univalue.Sorter)
	accesses1 := univalue.Univalues(slice.Clone(raw)).To(importer.IPTransition{})

	writeCache.Clear()                                                              // = committer.WriteCache()
	if _, err := adaptorcommon.CreateNewAccount(2, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	} // NewAccount account structure {

	// writeCache = committer.WriteCache()
	if _, err := adaptorcommon.CreateNewAccount(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	path2 := commutative.NewPath() // create a path
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	// committer.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	// committer.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	// accesses2, _ := committer.WriteCache().Export(univalue.Sorter)
	accesses2 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITAccess{})

	// accesses1.Print()
	// fmt.Print(" ++++++++++++++++++++++++++++++++++++++++++++++++ ")
	// accesses2.Print()

	IDVec := append(slice.Fill(make([]uint32, len(accesses1)), 0), slice.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses1, accesses2...))
	conflictdict, _, _ := arbitrator.Conflicts(ids).ToDict()

	if len(conflictdict) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}
}

func TestArbiTwoTxModifyTheSameAccount(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// writeCache.Write(stgcommcommon.SYSTEM, stgcommcommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)
	committer.SetStore(store)

	// committer.NewAccount(1, alice)
	writeCache.Clear()
	if _, err := adaptorcommon.CreateNewAccount(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/", commutative.NewPath()) // create a path
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, transitions1 := writeCache.Export(univalue.Sorter)
	accesses1 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITAccess{})
	transitions1 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITTransition{})

	// writeCache = committer.WriteCache()
	writeCache.Clear()
	if _, err := adaptorcommon.CreateNewAccount(2, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	} // NewAccount account structure {
	path2 := commutative.NewPath() // create a path

	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/", path2)
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))

	// accesses2, transitions2 := committer.WriteCache().Export(univalue.Sorter)
	accesses2 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITAccess{})
	transitions2 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITTransition{})

	IDVec := append(slice.Fill(make([]uint32, len(accesses1)), 0), slice.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses1, accesses2...))
	conflictDict, _, pairs := arbitrator.Conflicts(ids).ToDict()

	// pairs := arbitrator.Conflicts(ids).ToPairs()

	if len(conflictDict) != 1 || len(pairs) != 1 {
		t.Error("Error: There should be 1 conflict")
	}

	toCommit := slice.Exclude([]uint32{1, 2}, mapi.Keys(conflictDict))

	in := append(transitions1, transitions2...)
	buffer := univalue.Univalues(in).Encode()
	out := univalue.Univalues{}.Decode(buffer).(univalue.Univalues)

	committer = stgcommitter.NewStateCommitter(store)
	committer.Import(out)
	committer.Precommit(toCommit)
	committer.Commit(0)
	writeCache.Clear()

	if _, err := writeCache.Write(3, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("committer-1-by-tx-3")); err != nil {
		t.Error(err)
	}

	// accesses3, transitions3 := committer.Export(univalue.Sorter)
	accesses3 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITAccess{})
	transitions3 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.IPTransition{})

	writeCache.Clear()
	// url4 := stgcommitter.NewStateCommitter(store)
	if _, err := writeCache.Write(4, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("url4-1-by-tx-3")); err != nil {
		t.Error(err)
	}
	// accesses4, transitions4 := url4.Export(univalue.Sorter)
	accesses4 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITAccess{})
	transitions4 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(importer.ITTransition{})

	IDVec = append(slice.Fill(make([]uint32, len(accesses3)), 0), slice.Fill(make([]uint32, len(accesses4)), 1)...)
	ids = (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses3, accesses4...))
	conflictDict, _, _ = arbitrator.Conflicts(ids).ToDict()

	conflictTx := mapi.Keys(conflictDict)
	if len(conflictDict) != 1 || conflictTx[0] != 4 {
		t.Error("Error: There should be only 1 conflict")
	}

	toCommit = slice.RemoveIf(&[]uint32{3, 4}, func(_ int, tx uint32) bool {
		// conflictTx := mapi.Keys(*conflictDict)

		_, ok := conflictDict[tx]
		return ok
	})

	buffer = univalue.Univalues(append(transitions3, transitions4...)).Encode()
	out = univalue.Univalues{}.Decode(buffer).(univalue.Univalues)

	acctTrans = append(transitions3, transitions4...)
	committer = stgcommitter.NewStateCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	// committer.Import(committer.Decode(univalue.Univalues(append(transitions3, transitions4...)).Encode()))

	committer.Precommit(toCommit)
	committer.Commit(0)
	committer.SetStore(store)

	writeCache.Clear()
	v, _, _ := writeCache.Read(3, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", new(noncommutative.String))
	if v == nil || v.(string) != "committer-1-by-tx-3" {
		t.Error("Error: Wrong value, expecting:", "committer-1-by-tx-3 ", "actual:", v)
	}

	// have to mark balance and nonce persistent first !!!!!

	// v, _ = committer.Read(3, "blcc://eth1.0/account/"+alice+"/nonce", new(commutative.Uint64))
	// if v == nil || v.(uint64) != 2 {
	// 	t.Error("Error: Wrong value, expecting:", "2", "actual:", v)
	// }
}

func BenchmarkSimpleArbitrator(b *testing.B) {
	alice := AliceAccount()
	univalues := make([]*univalue.Univalue, 0, 5*200000)
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
