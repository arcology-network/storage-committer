package ccdb

import (
	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	ethcommon "github.com/arcology-network/evm/common"
	"github.com/arcology-network/evm/core/rawdb"
	"github.com/arcology-network/evm/core/types"
	ethdb "github.com/arcology-network/evm/ethdb"
	ethmpt "github.com/arcology-network/evm/trie"
	trienode "github.com/arcology-network/evm/trie/trienode"
)

type Rlp struct{}

func (Rlp) Encode(key string, v interface{}) []byte {
	if v == nil {
		return []byte{} // Deletion
	}
	return v.(interfaces.Type).StorageEncode()
}

func (Rlp) Decode(buffer []byte, T any) interface{} {
	return T.(interfaces.Type).StorageDecode(buffer)
}

type EthDataStore struct {
	trie       *ethmpt.Trie
	triedb     *ethmpt.Database
	latestRoot ethcommon.Hash
	memonly    bool
	nodeBuffer *trienode.NodeSet
	encoder    func(string, interface{}) []byte
	decoder    func([]byte, any) interface{}
}

func NewEthMemDataStore() *EthDataStore {
	diskdbs := [16]ethdb.Database{}
	common.Fill(diskdbs[:], rawdb.NewMemoryDatabase())
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	return &EthDataStore{
		triedb:  db,
		trie:    ethmpt.NewEmptyParallel(db),
		encoder: Rlp{}.Encode,
		decoder: Rlp{}.Decode,
	}
}

func (this *EthDataStore) Encoder() func(string, interface{}) []byte { return this.encoder }
func (this *EthDataStore) Decoder() func([]byte, any) interface{}    { return this.decoder }

func (this *EthDataStore) Trie() *ethmpt.Trie { return this.trie }

func (this *EthDataStore) IfExists(key string) bool {
	buffer, _ := this.trie.Get([]byte(key))
	return len(buffer) > 0
}

func (this *EthDataStore) Inject(key string, value interface{}) error {
	return this.trie.Update([]byte(key), value.([]byte))
}

func (this *EthDataStore) BatchInject(keys []string, values []interface{}) error {
	for i := 0; i < len(keys); i++ {
		this.Inject(keys[i], values[i])
	}
	return nil
}

func (this *EthDataStore) Retrive(key string, T any) (interface{}, error) {
	v, err := this.trie.Get([]byte(key))
	if err != nil || len(v) == 0 {
		return nil, err
	}
	return v, err

	// if bytes != nil && err == nil { // Get from the cache
	// 	value = this.decoder(bytes)
	// }
}

func (this *EthDataStore) BatchRetrive(keys []string, T []any) []interface{} {
	values := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		values[i], _ = this.trie.Get([]byte(keys[i]))
	}
	return values
}

func (this *EthDataStore) Precommit(keys []string, values interface{}) {
	encodedVals := make([][]byte, len(keys))

	valVec := values.([]interface{})
	encoder := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			encodedVals[i] = valVec[i].(interfaces.Univalue).Value().(interfaces.Type).Encode()
		}
	}
	common.ParallelWorker(len(keys), 8, encoder)

	keyBytes := codec.Strings(keys).ToBytes()
	this.trie.ParallelUpdate(keyBytes, encodedVals)
}

func (this *EthDataStore) Commit() error {
	this.latestRoot, this.nodeBuffer = this.trie.Commit(false)
	if err := this.triedb.Update(this.latestRoot, types.EmptyRootHash, trienode.NewWithNodeSet(this.nodeBuffer)); err != nil {
		return err
	}

	if !this.memonly {
		return this.triedb.Commit(this.latestRoot, false)
	}
	return nil
}

func (this *EthDataStore) GetMerkleRoot(keys []string, values interface{}) ethcommon.Hash {
	keyBytes := codec.Strings(keys).ToBytes()
	this.trie.ParallelUpdate(keyBytes, values.([][]byte))
	return this.trie.Hash()
}

// Place holders
func (this *EthDataStore) UpdateCacheStats([]interface{})  {}
func (this *EthDataStore) Dump() ([]string, []interface{}) { return nil, nil }
func (this *EthDataStore) Clear()                          {}
func (this *EthDataStore) Print()                          {}
func (this *EthDataStore) CheckSum() [32]byte              { return [32]byte{} }
func (this *EthDataStore) Query(string, func(string, string) bool) ([]string, [][]byte, error) {
	return nil, nil, nil
}
func (this *EthDataStore) CacheRetrive(key string, valueTransformer func(interface{}) interface{}) (interface{}, error) {
	return nil, nil
}
