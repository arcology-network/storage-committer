package ccurltest

import (
	"testing"

	// codec "github.com/arcology-network/common-lib/codec"

	datacompression "github.com/arcology-network/common-lib/addrcompressor"
	"github.com/arcology-network/common-lib/common"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
)

func TestMultiAccountCreation(t *testing.T) {
	store := chooseDataStore()
	// store := datastore.NewDataStore(nil, datastore.NewCachePolicy(0, 1), datastore.NewMemDB(), encoder, decoder)

	store.Inject((ccurlcommon.ETH10_ACCOUNT_PREFIX), commutative.NewPath())
	url := ccurl.NewConcurrentUrl(store)

	accounts := make([]string, 10)
	for i := 0; i < len(accounts); i++ {
		accounts[i] = datacompression.RandomAccount()
		if _, err := url.NewAccount(0, accounts[i]); err != nil { // Preload account structure {
			t.Error(err)
		}
	}
	raw := url.WriteCache().Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.ITCTransition{})

	paths := ccurlcommon.NewPlatform().GetSysPaths()
	if len(acctTrans) != len(paths)*len(accounts) {
		t.Error("Error: Transition counts don't match up")
	}

	url.Import(acctTrans)
	url.Sort()
	url.Commit([]uint32{0})

	// acctTrans = indexer.Univalues(common.Clone(raw)).To(indexer.ITCTransition{})
	// encoded := indexer.Univalues(acctTrans).Encode()

	// out := indexer.Univalues{}.Decode(encoded).(indexer.Univalues)
	// if len(acctTrans) != len(out) {
	// 	t.Error("Error: Transition counts don't match up")
	// }

	// accountMerkle := indexer.NewAccountMerkle(url.Platform, rlpEncoder, merkle.Keccak256{}.Hash)
	// accountMerkle.Import(out)
	// accountMerkle.Build()
}
