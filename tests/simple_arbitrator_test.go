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
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

func TestArbiCreateTwoAccountsNoConflict(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)

	meta := commutative.NewPath()
	url.Write(ccurlcommon.SYSTEM, ccurl.NewPlatform().Eth10Account(), meta)
	trans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	alice := datacompression.RandomAccount()
	url.Init(store)
	url.CreateAccount(1, url.Platform.Eth10(), alice) // CreateAccount account structure {
	// accesses1, transitions1 := url.Export(indexer.Sorter)
	accesses1 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(univalue.AccessCodecFilterSet()...)
	transitions1 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)

	bob := datacompression.RandomAccount()
	url2 := ccurl.NewConcurrentUrl(store)
	url2.CreateAccount(2, url.Platform.Eth10(), bob) // CreateAccount account structure {

	accesses2 := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(univalue.AccessCodecFilterSet()...)
	transitions2 := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)

	arib := (&arbitrator.Arbitrator{})
	ids := arib.Detect(append(accesses1, accesses2...))

	conflictTx := arbitrator.Conflicts(ids).TxIDs()
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(append(transitions1, transitions2...)).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit(conflictTx)

	if len(conflictTx) != 0 {
		t.Error("Error: There shouldn be 0 conflict")
	}
}

func TestArbiCreateTwoAccounts1Conflict(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)

	meta := commutative.NewPath()
	url.Write(ccurlcommon.SYSTEM, ccurl.NewPlatform().Eth10Account(), meta)
	trans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	alice := datacompression.RandomAccount()
	url.CreateAccount(1, url.Platform.Eth10(), alice)                      // CreateAccount account structure {
	path1 := commutative.NewPath()                                         // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path1) // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, _ := url.Export(indexer.Sorter)
	accesses1 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(univalue.AccessCodecFilterSet()...)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.CreateAccount(2, url.Platform.Eth10(), alice) // CreateAccount account structure {
	path2 := commutative.NewPath()                     // create a path
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	// accesses2, _ := url2.Export(indexer.Sorter)
	accesses2 := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(univalue.AccessCodecFilterSet()...)

	arib := (&arbitrator.Arbitrator{})
	ids := arib.Detect(append(accesses1, accesses2...))
	conflictTx := arbitrator.Conflicts(ids).TxIDs()

	conflictTx = common.UniqueInts(conflictTx)
	if len(conflictTx) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}
}

func TestArbiTwoTxModifyTheSameAccount(t *testing.T) {
	store := cachedstorage.NewDataStore()
	alice := datacompression.RandomAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	// url.Write(ccurlcommon.SYSTEM, ccurl.NewPlatform().Eth10Account(), commutative.NewPath())
	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})
	url.Init(store)

	url.CreateAccount(1, url.Platform.Eth10(), alice)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", commutative.NewPath()) // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, transitions1 := url.Export(indexer.Sorter)
	accesses1 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(univalue.AccessCodecFilterSet()...)
	transitions1 := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.CreateAccount(2, url.Platform.Eth10(), alice) // CreateAccount account structure {
	path2 := commutative.NewPath()                     // create a path

	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))

	// accesses2, transitions2 := url2.Export(indexer.Sorter)
	accesses2 := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(univalue.AccessCodecFilterSet()...)
	transitions2 := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)

	// aribi := (&arbitrator.Arbitrator{})

	// _, conflictTx := aribi.Detect(append(accesses1, accesses2...))

	ids := (&arbitrator.Arbitrator{}).Detect(append(accesses1, accesses2...))
	conflictTx := arbitrator.Conflicts(ids).TxIDs()

	if len(conflictTx) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}

	toCommit := common.Exclude([]uint32{1, 2}, conflictTx)

	in := indexer.Univalues(append(transitions1, transitions2...)).Encode()
	out := indexer.Univalues{}.Decode(in).(indexer.Univalues)
	url.Import(out)
	url.Sort()
	url.Commit(toCommit)

	url3 := ccurl.NewConcurrentUrl(store)
	if _, err := url3.Write(3, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("url3-1-by-tx-3")); err != nil {
		t.Error(err)
	}

	// accesses3, transitions3 := url3.Export(indexer.Sorter)
	accesses3 := indexer.Univalues(common.Clone(url3.Export(indexer.Sorter))).To(univalue.AccessCodecFilterSet()...)
	transitions3 := indexer.Univalues(common.Clone(url3.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)

	url4 := ccurl.NewConcurrentUrl(store)
	if _, err := url4.Write(4, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("url4-1-by-tx-3")); err != nil {
		t.Error(err)
	}
	// accesses4, transitions4 := url4.Export(indexer.Sorter)
	accesses4 := indexer.Univalues(common.Clone(url4.Export(indexer.Sorter))).To(univalue.AccessCodecFilterSet()...)
	transitions4 := indexer.Univalues(common.Clone(url4.Export(indexer.Sorter))).To(indexer.TransitionCodecFilterSet()...)

	ids = (&arbitrator.Arbitrator{}).Detect(append(accesses3, accesses4...))
	conflictTx = arbitrator.Conflicts(ids).TxIDs()

	if len(conflictTx) != 1 || conflictTx[0] != 4 {
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

func TestTimeSimpleArbitrator(b *testing.T) {
	// t0 := time.Now()
	alice := datacompression.RandomAccount()
	univalues := make([]ccurlcommon.UnivalueInterface, 5*200000)
	v := commutative.NewPath()
	tx := make([]uint32, len(univalues)/5)
	for i := 0; i < len(univalues)/5; i++ {
		univalues[i*5] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		univalues[i*5+1] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		univalues[i*5+2] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		univalues[i*5+3] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		univalues[i*5+4] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		tx[i] = uint32(i)
	}
	// fmt.Println("Create "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))

	t0 := time.Now()
	arib := (&arbitrator.Arbitrator{})
	arib.Detect(univalues)
	fmt.Println("Detect "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))
}

func BenchmarkSimpleArbitrator(b *testing.B) {
	// t0 := time.Now()
	alice := datacompression.RandomAccount()
	univalues := make([]ccurlcommon.UnivalueInterface, 5*200000)
	v := commutative.NewPath()
	tx := make([]uint32, len(univalues)/5)
	for i := 0; i < len(univalues)/5; i++ {
		univalues[i*5] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		univalues[i*5+1] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		univalues[i*5+2] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		univalues[i*5+3] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		univalues[i*5+4] = univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v)
		tx[i] = uint32(i)
	}
	// fmt.Println("Create "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))

	t0 := time.Now()
	arib := (&arbitrator.Arbitrator{})
	arib.Detect(univalues)
	fmt.Println("Detect "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))
}
