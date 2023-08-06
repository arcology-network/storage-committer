package ccurltest

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	common "github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl"
	arbitrator "github.com/arcology-network/concurrenturl/arbitrator"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

func TestArbiCreateTwoAccountsNoConflict(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)

	meta := commutative.NewPath()

	url.Write(ccurlcommon.SYSTEM, ccurlcommon.ETH10_ACCOUNT_PREFIX, meta, true)
	trans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	alice := datacompression.RandomAccount()
	url.Init(store)
	url.NewAccount(1, alice) // NewAccount account structure {
	// accesses1, transitions1 := url.Export(indexer.Sorter)
	accesses1 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	bob := datacompression.RandomAccount()
	url2 := ccurl.NewConcurrentUrl(store)
	url2.NewAccount(2, bob) // NewAccount account structure {

	accesses2 := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	arib := (&arbitrator.Arbitrator{})

	IDVec := append(common.Fill(make([]uint32, len(accesses1)), 0), common.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := arib.Detect(IDVec, append(accesses1, accesses2...))

	conflictdict, _ := arbitrator.Conflicts(ids).ToDict()
	if len(*conflictdict) != 0 {
		t.Error("Error: There shouldn be 0 conflict")
	}
}

func TestArbiCreateTwoAccounts1Conflict(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)

	meta := commutative.NewPath()
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.ETH10_ACCOUNT_PREFIX, meta, true)
	trans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	alice := datacompression.RandomAccount()
	url.NewAccount(1, alice)                                                     // NewAccount account structure {
	path1 := commutative.NewPath()                                               // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path1, true) // create a path
	// url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	// url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, _ := url.Export(indexer.Sorter)
	accesses1 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCAccess{})

	url2 := ccurl.NewConcurrentUrl(store)
	url2.NewAccount(2, alice)      // NewAccount account structure {
	path2 := commutative.NewPath() // create a path
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2, true)
	// url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	// url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	// accesses2, _ := url2.Export(indexer.Sorter)
	accesses2 := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(indexer.ITCAccess{})

	accesses1.Print()
	fmt.Print(" ++++++++++++++++++++++++++++++++++++++++++++++++ ")
	accesses2.Print()

	IDVec := append(common.Fill(make([]uint32, len(accesses1)), 0), common.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses1, accesses2...))
	conflictdict, _ := arbitrator.Conflicts(ids).ToDict()

	if len(*conflictdict) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}
}

func TestArbiTwoTxModifyTheSameAccount(t *testing.T) {
	store := cachedstorage.NewDataStore()
	alice := datacompression.RandomAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// url.Write(ccurlcommon.SYSTEM, ccurlcommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())
	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})
	url.Init(store)

	url.NewAccount(1, alice)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", commutative.NewPath(), true) // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"), true)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"), true)
	// accesses1, transitions1 := url.Export(indexer.Sorter)
	accesses1 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	transitions1 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url2 := ccurl.NewConcurrentUrl(store)
	url2.NewAccount(2, alice)      // NewAccount account structure {
	path2 := commutative.NewPath() // create a path

	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2, true)
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"), true)
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"), true)

	// accesses2, transitions2 := url2.Export(indexer.Sorter)
	accesses2 := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	transitions2 := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	IDVec := append(common.Fill(make([]uint32, len(accesses1)), 0), common.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses1, accesses2...))
	conflictDict, pairs := arbitrator.Conflicts(ids).ToDict()

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

	url3 := ccurl.NewConcurrentUrl(store)
	if _, err := url3.Write(3, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("url3-1-by-tx-3"), true); err != nil {
		t.Error(err)
	}

	// accesses3, transitions3 := url3.Export(indexer.Sorter)
	accesses3 := indexer.Univalues(common.Clone(url3.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	transitions3 := indexer.Univalues(common.Clone(url3.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url4 := ccurl.NewConcurrentUrl(store)
	if _, err := url4.Write(4, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("url4-1-by-tx-3"), true); err != nil {
		t.Error(err)
	}
	// accesses4, transitions4 := url4.Export(indexer.Sorter)
	accesses4 := indexer.Univalues(common.Clone(url4.Export(indexer.Sorter))).To(indexer.ITCAccess{})
	transitions4 := indexer.Univalues(common.Clone(url4.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	IDVec = append(common.Fill(make([]uint32, len(accesses3)), 0), common.Fill(make([]uint32, len(accesses4)), 1)...)
	ids = (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses3, accesses4...))
	conflictDict, _ = arbitrator.Conflicts(ids).ToDict()

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

	v, _ := url3.Read(3, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1")
	if v == nil || v.(string) != "url3-1-by-tx-3" {
		t.Error("Error: Wrong value")
	}
}

// func TestTimeSimpleArbitrator(b *testing.T) {
// 	// t0 := time.Now()
// 	alice := datacompression.RandomAccount()
// 	univalues := make([]interfaces.Univalue, 0, 5*10000)
// 	v := commutative.NewPath()
// 	tx := make([]uint32, 0, len(univalues)/5)
// 	for i := 0; i < len(univalues)/5; i++ {
// 		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))
// 		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))
// 		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))
// 		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))
// 		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))

// 		tx = append(tx, uint32(i))
// 		tx = append(tx, uint32(i))
// 		tx = append(tx, uint32(i))
// 		tx = append(tx, uint32(i))
// 		tx = append(tx, uint32(i))
// 	}
// 	// fmt.Println("Create "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))

// 	t0 := time.Now()
// 	(&arbitrator.Arbitrator{}).Detect(tx, univalues)
// 	fmt.Println("Detect "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))
// }

func BenchmarkSimpleArbitrator(b *testing.B) {
	alice := datacompression.RandomAccount()
	univalues := make([]interfaces.Univalue, 0, 5*200000)
	groupIDs := make([]uint32, 0, len(univalues))

	v := commutative.NewPath()
	for i := 0; i < len(univalues)/5; i++ {
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v))

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
