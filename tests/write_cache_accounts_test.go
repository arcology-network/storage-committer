package committertest

import (
	"testing"

	// codec "github.com/arcology-network/common-lib/codec"

	"github.com/arcology-network/common-lib/exp/slice"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	statestore "github.com/arcology-network/storage-committer"
	stgcommitter "github.com/arcology-network/storage-committer/committer"
	platform "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/storage/proxy"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

func TestCacheMultiAccountCreation(t *testing.T) {
	store := chooseDataStore()
	// store := datastore.NewDataStore( datastore.NewCachePolicy(0, 1), datastore.NewMemoryDB(), encoder, decoder)

	// store.Inject((stgcommcommon.ETH10_ACCOUNT_PREFIX), commutative.NewPath())

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	accounts := make([]string, 10)
	for i := 0; i < len(accounts); i++ {
		accounts[i] = RandomAccount()
		if _, err := adaptorcommon.CreateNewAccount(0, accounts[i], writeCache); err != nil { // NewAccount account structure {
			t.Error(err)
		}
	}
	raw := writeCache.Export(univalue.Sorter)
	acctTrans := univalue.Univalues(slice.Clone(raw)).To(univalue.ITTransition{})

	paths := platform.NewPlatform().GetSysPaths()
	if len(acctTrans) != len(paths)*len(accounts) {
		t.Error("Error: Transition counts don't match up")
	}

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters()...)
	committer.Import(acctTrans)

	committer.Precommit([]uint32{0})
	committer.Commit(0)

	// acctTrans = univalue.Univalues(slice.Clone(raw)).To(univalue.ITTransition{})
	// encoded := univalue.Univalues(acctTrans).Encode()

	// out := univalue.Univalues{}.Decode(encoded).(univalue.Univalues)
	// if len(acctTrans) != len(out) {
	// 	t.Error("Error: Transition counts don't match up")
	// }

	// accountMerkle := importer.NewAccountMerkle(committer.Platform, rlpEncoder, merkle.Keccak256{}.Hash)
	// accountMerkle.Import(out)
	// accountMerkle.Build()
}
