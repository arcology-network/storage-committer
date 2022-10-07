package ccurltest

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	cachedstorage "github.com/HPISTechnologies/common-lib/cachedstorage"
	datacompression "github.com/HPISTechnologies/common-lib/datacompression"
	"github.com/HPISTechnologies/concurrenturl/v2"
	ccurl "github.com/HPISTechnologies/concurrenturl/v2"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	ccurltype "github.com/HPISTechnologies/concurrenturl/v2/type"
	commutative "github.com/HPISTechnologies/concurrenturl/v2/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/v2/type/noncommutative"
)

func TestArbiCreateTwoAccountsNoConflict(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)

	meta, _ := commutative.NewMeta(ccurlcommon.NewPlatform().Eth10Account())
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.NewPlatform().Eth10Account(), meta)
	_, trans := url.Export(false)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(trans).Encode()).(ccurltype.Univalues))

	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	alice := datacompression.RandomAccount()
	url.Init(store)
	url.CreateAccount(1, url.Platform.Eth10(), alice) // CreateAccount account structure {
	accesses1, transitions1 := url.Export(true)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.CreateAccount(2, url.Platform.Eth10(), "bob") // CreateAccount account structure {
	accesses2, transitions2 := url2.Export(true)

	arib := ccurl.NewArbitratorSlow()

	concurrenturl.HashPaths(accesses1)
	concurrenturl.HashPaths(accesses2)
	_, conflictTx := arib.Detect(append(accesses1, accesses2...), []uint32{1, 2})

	// in := ccurltype.Univalues(append(transitions1, transitions2...)).Encode()
	// out := ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	// url.Import(url.Decode(ccurltype.Univalues(out).Encode()))

	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(append(transitions1, transitions2...)).Encode()).(ccurltype.Univalues))

	url.PostImport()
	url.Commit(conflictTx)

	if len(conflictTx) != 0 {
		t.Error("Error: There shouldn be 0 conflict")
	}
}

func TestArbiCreateTwoAccounts1Conflict(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)

	meta, _ := commutative.NewMeta(ccurlcommon.NewPlatform().Eth10Account())
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.NewPlatform().Eth10Account(), meta)
	_, trans := url.Export(false)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(trans).Encode()).(ccurltype.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	alice := datacompression.RandomAccount()
	url.CreateAccount(1, url.Platform.Eth10(), alice)                                      // CreateAccount account structure {
	path1, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-1/") // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path1)                 // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	accesses1, _ := url.Export(true)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.CreateAccount(2, url.Platform.Eth10(), alice)                                     // CreateAccount account structure {
	path2, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-2/") // create a path
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	accesses2, _ := url2.Export(true)

	arib := ccurl.NewArbitratorSlow()
	_, conflictTx := arib.Detect(append(accesses1, accesses2...), []uint32{1, 2})

	if len(conflictTx) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}
}

func TestArbiTwoTxModifyTheSameAccount(t *testing.T) {
	store := cachedstorage.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	meta, _ := commutative.NewMeta(ccurlcommon.NewPlatform().Eth10Account())
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.NewPlatform().Eth10Account(), meta)
	_, trans := url.Export(false)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(trans).Encode()).(ccurltype.Univalues))
	url.PostImport()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	alice := datacompression.RandomAccount()
	url.CreateAccount(1, url.Platform.Eth10(), alice)                                      // CreateAccount account structure {
	path1, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-1/") // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path1)                 // create a path
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	accesses1, transitions1 := url.Export(true)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.CreateAccount(2, url.Platform.Eth10(), alice)                                     // CreateAccount account structure {
	path2, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-2/") // create a path
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	accesses2, transitions2 := url2.Export(true)

	arib := ccurl.NewArbitratorSlow()
	concurrenturl.HashPaths(accesses1)
	concurrenturl.HashPaths(accesses2)
	_, conflictTx := arib.Detect(append(accesses1, accesses2...), []uint32{1, 2})
	if len(conflictTx) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}

	toCommit := ccurlcommon.Exclude([]uint32{1, 2}, conflictTx)

	in := ccurltype.Univalues(append(transitions1, transitions2...)).Encode()
	out := ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)
	// url.Import(url.Decode(ccurltype.Univalues(out).Encode()))

	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(out).Encode()).(ccurltype.Univalues))

	url.PostImport()
	url.Commit(toCommit)

	url3 := ccurl.NewConcurrentUrl(store)
	url3.Write(3, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("url3-1-by-tx-3"))
	accesses3, transitions3 := url3.Export(true)

	url4 := ccurl.NewConcurrentUrl(store)
	url4.Write(4, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("url4-1-by-tx-3"))
	accesses4, transitions4 := url4.Export(true)

	_, conflictTx = arib.Detect(append(accesses3, accesses4...), []uint32{3, 4})
	if len(conflictTx) != 1 || conflictTx[0] != 4 {
		t.Error("Error: There shouldn 1 conflict")
	}

	toCommit = ccurlcommon.Exclude([]uint32{3, 4}, conflictTx)

	in = ccurltype.Univalues(append(transitions3, transitions4...)).Encode()
	out = ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)

	trans = append(transitions3, transitions4...)
	url.Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(trans).Encode()).(ccurltype.Univalues))

	// url.Import(url.Decode(ccurltype.Univalues(append(transitions3, transitions4...)).Encode()))
	url.PostImport()
	url.Commit(toCommit)

	v, err := url3.Read(3, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1")
	if err != nil || string(*(v.(*noncommutative.String))) != "url3-1-by-tx-3" {
		t.Error("Error: Wrong value")
	}
}

func BenchmarkSimpleArbitrator(b *testing.B) {
	t0 := time.Now()
	alice := datacompression.RandomAccount()
	univalues := make([]ccurlcommon.UnivalueInterface, 5*200000)
	v, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	tx := make([]uint32, len(univalues)/5)
	for i := 0; i < len(univalues)/5; i++ {
		univalues[i*5] = ccurltype.NewUnivalue(ccurlcommon.VARIATE_TRANSITIONS, uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, v)
		univalues[i*5+1] = ccurltype.NewUnivalue(ccurlcommon.VARIATE_TRANSITIONS, uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, v)
		univalues[i*5+2] = ccurltype.NewUnivalue(ccurlcommon.VARIATE_TRANSITIONS, uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, v)
		univalues[i*5+3] = ccurltype.NewUnivalue(ccurlcommon.VARIATE_TRANSITIONS, uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, v)
		univalues[i*5+4] = ccurltype.NewUnivalue(ccurlcommon.VARIATE_TRANSITIONS, uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, v)
		tx[i] = uint32(i)
	}
	fmt.Println("Create "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))

	t0 = time.Now()
	arib := ccurl.NewArbitratorSlow()
	arib.Detect(univalues, tx)
	fmt.Println("Detect "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))
}
