package committertest

import (
	"testing"

	// codec "github.com/arcology-network/common-lib/codec"

	"github.com/arcology-network/common-lib/exp/array"
	cache "github.com/arcology-network/eu/cache"
	stgcommitter "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	importer "github.com/arcology-network/storage-committer/importer"
	platform "github.com/arcology-network/storage-committer/platform"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

func TestMultiAccountCreation(t *testing.T) {
	store := chooseDataStore()
	// store := datastore.NewDataStore(nil, datastore.NewCachePolicy(0, 1), datastore.NewMemoryDB(), encoder, decoder)

	store.Inject((stgcommcommon.ETH10_ACCOUNT_PREFIX), commutative.NewPath())

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	accounts := make([]string, 10)
	for i := 0; i < len(accounts); i++ {
		accounts[i] = RandomAccount()
		if _, err := writeCache.CreateNewAccount(0, accounts[i]); err != nil { // NewAccount account structure {
			t.Error(err)
		}
	}
	raw := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(array.Clone(raw)).To(importer.ITTransition{})

	paths := platform.NewPlatform().GetSysPaths()
	if len(acctTrans) != len(paths)*len(accounts) {
		t.Error("Error: Transition counts don't match up")
	}

	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(acctTrans)
	committer.Sort()
	committer.Precommit([]uint32{0})
	committer.Commit()
	writeCache.Reset(writeCache)
	// acctTrans = univalue.Univalues(array.Clone(raw)).To(importer.ITTransition{})
	// encoded := univalue.Univalues(acctTrans).Encode()

	// out := univalue.Univalues{}.Decode(encoded).(univalue.Univalues)
	// if len(acctTrans) != len(out) {
	// 	t.Error("Error: Transition counts don't match up")
	// }

	// accountMerkle := importer.NewAccountMerkle(committer.Platform, rlpEncoder, merkle.Keccak256{}.Hash)
	// accountMerkle.Import(out)
	// accountMerkle.Build()
}
