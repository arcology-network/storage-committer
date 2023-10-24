package ccurltest

import (
	"fmt"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	"github.com/arcology-network/common-lib/merkle"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	storage "github.com/arcology-network/concurrenturl/storage"
)

func TestMultiAccountCreation(t *testing.T) {
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	store.Inject((ccurlcommon.ETH10_ACCOUNT_PREFIX), commutative.NewPath())
	url := ccurl.NewConcurrentUrl(store)

	accounts := make([]string, 10)
	for i := 0; i < len(accounts); i++ {
		accounts[i] = datacompression.RandomAccount()
		if err := url.NewAccount(0, accounts[i]); err != nil { // Preload account structure {
			t.Error(err)
		}
	}
	raw := url.Export(indexer.Sorter)
	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.ITCTransition{})

	paths := ccurlcommon.NewPlatform().GetSysPaths()
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

	accountMerkle := indexer.NewAccountMerkle(url.Platform, rlpEncoder, merkle.Keccak256{}.Hash)
	accountMerkle.Import(out)

	// accountMerkle.Build()
}

func TestRlps(t *testing.T) {
	// fmt.Println(rlpEncoder(uint64(1230)))
	fmt.Println(rlpEncoder(string("1234")))

	fmt.Println(rlpEncoder(uint64(1230), string("1234")))
	v := commutative.NewUint64(uint64(10), uint64(12000000000000)).(*commutative.Uint64)
	fmt.Println(v)

	t0 := time.Now()
	for i := 0; i < 1000000; i++ {
		v := commutative.NewUint64(uint64(10), uint64(12000000000000)).(*commutative.Uint64)
		rlpEncoder(v.Value().(*codec.Uint64), v.Min().(*codec.Uint64), v.Max().(*codec.Uint64), v.Delta().(*codec.Uint64))
	}
	fmt.Println("Encode() rlp: ", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 1000000; i++ {
		v := commutative.NewUint64(uint64(10), uint64(12000000000000)).(*commutative.Uint64)
		v.Encode()
	}
	fmt.Println("Encode() native: ", time.Since(t0))
	// codec.ByteSet()
}
