package ccurltest

import (
	"testing"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	"github.com/arcology-network/concurrenturl"
	ccurl "github.com/arcology-network/concurrenturl"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

func TestMultiAccountCreation(t *testing.T) {
	store := cachedstorage.NewDataStore()
	store.Inject((ccurl.NewPlatform().Eth10Account()), commutative.NewPath())
	url := ccurl.NewConcurrentUrl(store)

	accounts := make([]string, 10)
	for i := 0; i < len(accounts); i++ {
		accounts[i] = datacompression.RandomAccount()
		if err := url.CreateAccount(0, (url.Platform.Eth10()), accounts[i]); err != nil { // Preload account structure {
			t.Error(err)
		}
	}
	raw := url.Export(univalue.Sorter)
	acctTrans := univalue.Univalues(common.Clone(raw)).To(univalue.TransitionCodecFilterSet()...)

	paths := concurrenturl.NewPlatform().GetSysPaths()

	if len(acctTrans) != len(paths)*len(accounts) {
		t.Error("Error: Transition counts don't match up")
	}

	url.Import(acctTrans)
	url.Sort()
	url.Commit([]uint32{0})

	acctTrans = univalue.Univalues(common.Clone(raw)).To(univalue.TransitionCodecFilterSet()...)
	encoded := univalue.Univalues(acctTrans).Encode()

	out := univalue.Univalues{}.Decode(encoded).(univalue.Univalues)
	if len(acctTrans) != len(out) {
		t.Error("Error: Transition counts don't match up")
	}

	accountMerkle := indexer.NewAccountMerkle(url.Platform)
	accountMerkle.Import(out)

	// accountMerkle.Build()

}
