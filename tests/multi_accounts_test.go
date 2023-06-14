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
)

func TestMultiAccountCreation(t *testing.T) {
	store := cachedstorage.NewDataStore()
	store.Inject((ccurl.NewPlatform().Eth10Account()), commutative.NewPath())
	url := ccurl.NewConcurrentUrl(store)

	accounts := make([]string, 10)
	for i := 0; i < len(accounts); i++ {
		accounts[i] = datacompression.RandomAccount()
		if err := url.NewAccount(0, (url.Platform.Eth10()), accounts[i]); err != nil { // Preload account structure {
			t.Error(err)
		}
	}
	raw := url.Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.ITCTransition{})

	paths := concurrenturl.NewPlatform().GetSysPaths()

	if len(acctTrans) != len(paths)*len(accounts) {
		t.Error("Error: Transition counts don't match up")
	}

	url.Import(acctTrans)
	url.Sort()
	url.Commit([]uint32{0})

	acctTrans = indexer.Univalues(common.Clone(raw)).To(indexer.ITCTransition{})
	encoded := indexer.Univalues(acctTrans).Encode()

	out := indexer.Univalues{}.Decode(encoded).(indexer.Univalues)
	if len(acctTrans) != len(out) {
		t.Error("Error: Transition counts don't match up")
	}

	accountMerkle := indexer.NewAccountMerkle(url.Platform)
	accountMerkle.Import(out)

	// accountMerkle.Build()

}
