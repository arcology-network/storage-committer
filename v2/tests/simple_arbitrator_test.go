package ccurltest

import (
	"testing"

	ccurl "github.com/arcology/concurrenturl/v2"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
	commutative "github.com/arcology/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology/concurrenturl/v2/type/noncommutative"
)

func TestArbiCreateTwoAccountsNoConflict(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)

	meta, _ := commutative.NewMeta(ccurlcommon.NewPlatform().Eth10Account())
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.NewPlatform().Eth10Account(), meta)
	_, trans := url.Export(false)
	url.Commit(trans, []uint32{ccurlcommon.SYSTEM})

	url.Preload(1, url.Platform.Eth10(), "alice") // Preload account structure {
	accesses1, transitions1 := url.Export(true)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.Preload(2, url.Platform.Eth10(), "bob") // Preload account structure {
	accesses2, transitions2 := url2.Export(true)

	arib := ccurl.NewArbitratorSlow()
	_, conflictTx := arib.Detect(append(accesses1, accesses2...), []uint32{1, 2})
	url.Commit(append(transitions1, transitions2...), conflictTx)

	if len(conflictTx) != 0 {
		t.Error("Error: There shouldn be 0 conflict")
	}
}

func TestArbiCreateTwoAccounts1Conflict(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)

	meta, _ := commutative.NewMeta(ccurlcommon.NewPlatform().Eth10Account())
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.NewPlatform().Eth10Account(), meta)
	_, trans := url.Export(false)
	url.Commit(trans, []uint32{ccurlcommon.SYSTEM})

	url.Preload(1, url.Platform.Eth10(), "alice")                                  // Preload account structure {
	path1, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-1/") // create a path
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-2/", path1)             // create a path
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	accesses1, _ := url.Export(true)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.Preload(2, url.Platform.Eth10(), "alice")                                 // Preload account structure {
	path2, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-2/") // create a path
	url2.Write(2, "blcc://eth1.0/account/alice/storage/ctrn-2/", path2)
	url2.Write(2, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	url2.Write(2, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	accesses2, _ := url2.Export(true)

	arib := ccurl.NewArbitratorSlow()
	_, conflictTx := arib.Detect(append(accesses1, accesses2...), []uint32{1, 2})

	if len(conflictTx) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}
}

func TestArbiTwoTxModifyTheSameAccount(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	meta, _ := commutative.NewMeta(ccurlcommon.NewPlatform().Eth10Account())
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.NewPlatform().Eth10Account(), meta)
	_, trans := url.Export(false)
	url.Commit(trans, []uint32{ccurlcommon.SYSTEM})

	url.Preload(1, url.Platform.Eth10(), "alice")                                  // Preload account structure {
	path1, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-1/") // create a path
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-2/", path1)             // create a path
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	accesses1, transitions1 := url.Export(true)

	url2 := ccurl.NewConcurrentUrl(store)
	url2.Preload(2, url.Platform.Eth10(), "alice")                                 // Preload account structure {
	path2, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-2/") // create a path
	url2.Write(2, "blcc://eth1.0/account/alice/storage/ctrn-2/", path2)
	url2.Write(2, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	url2.Write(2, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	accesses2, transitions2 := url2.Export(true)

	arib := ccurl.NewArbitratorSlow()
	_, conflictTx := arib.Detect(append(accesses1, accesses2...), []uint32{1, 2})
	if len(conflictTx) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}

	toCommit := ccurlcommon.Exclude([]uint32{1, 2}, conflictTx)
	url.Commit(append(transitions1, transitions2...), toCommit)

	url3 := ccurl.NewConcurrentUrl(store)
	url3.Write(3, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("url3-1-by-tx-3"))
	accesses3, transitions3 := url3.Export(true)

	url4 := ccurl.NewConcurrentUrl(store)
	url4.Write(4, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1", noncommutative.NewString("url4-1-by-tx-3"))
	accesses4, transitions4 := url4.Export(true)

	_, conflictTx = arib.Detect(append(accesses3, accesses4...), []uint32{3, 4})
	if len(conflictTx) != 1 || conflictTx[0] != 4 {
		t.Error("Error: There shouldn 1 conflict")
	}

	toCommit = ccurlcommon.Exclude([]uint32{3, 4}, conflictTx)
	url.Commit(append(transitions3, transitions4...), toCommit)

	v, err := url3.Read(3, "blcc://eth1.0/account/alice/storage/ctrn-2/elem-1")
	if err != nil || string(*(v.(*noncommutative.String))) != "url3-1-by-tx-3" {
		t.Error("Error: Wrong value")
	}
}
