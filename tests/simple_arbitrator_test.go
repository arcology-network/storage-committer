package ccurltest

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl"
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
	trans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))

	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	alice := datacompression.RandomAccount()
	url.Init(store)
	url.CreateAccount(1, url.Platform.Eth10(), alice) // CreateAccount account structure {
	// accesses1, transitions1 := url.Export(ccurlcommon.Sorter)
	transitions1 := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	accesses1 := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.AccessFilters()...)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.CreateAccount(2, url.Platform.Eth10(), "bob") // CreateAccount account structure {

	accesses2 := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.AccessFilters()...)
	transitions2 := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	arib := indexer.NewArbitratorSlow()

	indexer.HashPaths(accesses1)
	indexer.HashPaths(accesses2)
	_, conflictTx := arib.Detect(append(accesses1, accesses2...))

	// in := univalue.Univalues(append(transitions1, transitions2...)).Encode()
	// out := univalue.Univalues{}.Decode(in).(univalue.Univalues)
	// url.Import(url.Decode(univalue.Univalues(out).Encode()))

	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(append(transitions1, transitions2...)).Encode()).(univalue.Univalues))

	url.PostImport()
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
	trans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	alice := datacompression.RandomAccount()
	url.CreateAccount(1, url.Platform.Eth10(), alice)                      // CreateAccount account structure {
	path1 := commutative.NewPath()                                         // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path1) // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, _ := url.Export(ccurlcommon.Sorter)
	accesses1 := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.AccessFilters()...)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.CreateAccount(2, url.Platform.Eth10(), alice) // CreateAccount account structure {
	path2 := commutative.NewPath()                     // create a path
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	// accesses2, _ := url2.Export(ccurlcommon.Sorter)
	accesses2 := univalue.Univalues(common.Clone(url2.Export(ccurlcommon.Sorter))).To(univalue.AccessFilters()...)

	arib := indexer.NewArbitratorSlow()
	_, conflictTx := arib.Detect(append(accesses1, accesses2...))

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
	acctTrans := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})
	url.Init(store)

	url.CreateAccount(1, url.Platform.Eth10(), alice)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", commutative.NewPath()) // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, transitions1 := url.Export(ccurlcommon.Sorter)
	accesses1 := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.AccessFilters()...)
	transitions1 := univalue.Univalues(common.Clone(url.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.CreateAccount(2, url.Platform.Eth10(), alice) // CreateAccount account structure {
	path2 := commutative.NewPath()                     // create a path

	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))

	// accesses2, transitions2 := url2.Export(ccurlcommon.Sorter)
	accesses2 := univalue.Univalues(common.Clone(url2.Export(ccurlcommon.Sorter))).To(univalue.AccessFilters()...)
	transitions2 := univalue.Univalues(common.Clone(url2.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	aribi := indexer.NewArbitratorSlow()
	// indexer.HashPaths(accesses1)
	// indexer.HashPaths(accesses2)
	_, conflictTx := aribi.Detect(append(accesses1, accesses2...))
	if len(conflictTx) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}

	toCommit := ccurlcommon.Exclude([]uint32{1, 2}, conflictTx)

	in := univalue.Univalues(append(transitions1, transitions2...)).Encode()
	out := univalue.Univalues{}.Decode(in).(univalue.Univalues)
	url.Import(out)
	url.PostImport()
	url.Commit(toCommit)

	url3 := ccurl.NewConcurrentUrl(store)
	if err := url3.Write(3, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("url3-1-by-tx-3")); err != nil {
		t.Error(err)
	}

	// accesses3, transitions3 := url3.Export(ccurlcommon.Sorter)
	accesses3 := univalue.Univalues(common.Clone(url3.Export(ccurlcommon.Sorter))).To(univalue.AccessFilters()...)
	transitions3 := univalue.Univalues(common.Clone(url3.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	url4 := ccurl.NewConcurrentUrl(store)
	if err := url4.Write(4, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("url4-1-by-tx-3")); err != nil {
		t.Error(err)
	}
	// accesses4, transitions4 := url4.Export(ccurlcommon.Sorter)
	accesses4 := univalue.Univalues(common.Clone(url4.Export(ccurlcommon.Sorter))).To(univalue.AccessFilters()...)
	transitions4 := univalue.Univalues(common.Clone(url4.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)

	aribi = indexer.NewArbitratorSlow()
	_, conflictTx = aribi.Detect(append(accesses3, accesses4...))
	if len(conflictTx) != 1 || conflictTx[0] != 4 {
		t.Error("Error: There should be 1 conflict")
	}
	toCommit = ccurlcommon.Exclude([]uint32{3, 4}, conflictTx)

	in = univalue.Univalues(append(transitions3, transitions4...)).Encode()
	out = univalue.Univalues{}.Decode(in).(univalue.Univalues)

	acctTrans = append(transitions3, transitions4...)
	url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	// url.Import(url.Decode(univalue.Univalues(append(transitions3, transitions4...)).Encode()))
	url.PostImport()
	url.Commit(toCommit)

	v, err := url3.Read(3, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1")
	if err != nil || v.(string) != "url3-1-by-tx-3" {
		t.Error("Error: Wrong value")
	}
}

func BenchmarkSimpleArbitrator(b *testing.B) {
	t0 := time.Now()
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
	fmt.Println("Create "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))

	t0 = time.Now()
	arib := indexer.NewArbitratorSlow()
	arib.Detect(univalues)
	fmt.Println("Detect "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))
}
