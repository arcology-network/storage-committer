package ccdb

import (
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	ethcommon "github.com/arcology-network/evm/common"
	"github.com/arcology-network/evm/core/rawdb"
	"github.com/arcology-network/evm/core/types"
	ethdb "github.com/arcology-network/evm/ethdb"
	ethmpt "github.com/arcology-network/evm/trie"
	trienode "github.com/arcology-network/evm/trie/trienode"
	"golang.org/x/crypto/sha3"
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
	ethdb      *ethmpt.Database
	latestRoot ethcommon.Hash
	memonly    bool
	nodeBuffer *trienode.NodeSet
	encoder    func(string, interface{}) []byte
	decoder    func([]byte, any) interface{}
}

func NewEthMemDataStore(memonly bool) *EthDataStore {
	diskdbs := [16]ethdb.Database{}
	common.Fill(diskdbs[:], rawdb.NewMemoryDatabase())
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	paraTrie := ethmpt.NewEmptyParallel(db)
	return &EthDataStore{
		ethdb:   db,
		trie:    paraTrie,
		memonly: memonly,
		encoder: Rlp{}.Encode,
		decoder: Rlp{}.Decode,
	}
}

func (this *EthDataStore) Clear() {
	var err error
	this.trie, err = ethmpt.NewParallel(ethmpt.TrieID(this.latestRoot), this.ethdb) // reopen the trie for future use
	if err != nil {
		panic(err)
	}
}

func (this *EthDataStore) Hash(key string) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(key))
	sum := hasher.Sum(nil)
	return sum
}

func (this *EthDataStore) LoadTrie(root [32]byte) *ethmpt.Trie {
	if this.latestRoot != root {
		if proofTrie, err := ethmpt.NewParallel(ethmpt.TrieID(this.latestRoot), this.ethdb); err == nil {
			return proofTrie
		}
		return nil
	}
	return this.trie
}

func (this *EthDataStore) Root() [32]byte { return this.latestRoot }

func (this *EthDataStore) Encoder() func(string, interface{}) []byte { return this.encoder }
func (this *EthDataStore) Decoder() func([]byte, any) interface{}    { return this.decoder }

func (this *EthDataStore) IfExists(key string) bool {
	buffer, _ := this.trie.Get(this.Hash(key))
	return len(buffer) > 0
}

// Update the trie
func (this *EthDataStore) WriteTrie(keys []string, values []interface{}, encode func(v interface{}) []byte) error {
	this.trie.ParallelUpdate(
		common.ParallelAppend(keys, func(i int) []byte { return this.Hash(keys[i]) }),
		common.ParallelAppend(values, func(i int) []byte { return encode(values[i]) }))
	return nil
}

func (this *EthDataStore) Inject(key string, value interface{}) error {
	return this.BatchInject([]string{key}, []interface{}{value})
}

func (this *EthDataStore) BatchInject(keys []string, values []interface{}) error {
	this.WriteTrie(keys, values, func(v interface{}) []byte {
		return v.(interfaces.Type).StorageEncode()
	})
	return nil
}

func (this *EthDataStore) Retrive(key string, T any) (interface{}, error) {
	v, err := this.trie.Get(this.Hash(key))
	if err != nil || len(v) == 0 { // Not found
		return nil, err
	}

	if T == nil { // A deletion
		return T, nil
	}
	return T.(interfaces.Type).StorageDecode(v), err
}

func (this *EthDataStore) BatchRetrive(keys []string, T []any) []interface{} {
	values := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		values[i], _ = this.trie.Get(this.Hash(keys[i]))
	}
	return values
}

// Encode KVs and write them to the trie.
func (this *EthDataStore) Precommit(keys []string, values interface{}) [32]byte {
	this.WriteTrie(keys, values.([]interface{}), func(v interface{}) []byte {
		if v.(interfaces.Univalue).Value() == nil {
			return []byte{} //Deletion
		}
		return v.(interfaces.Univalue).Value().(interfaces.Type).StorageEncode()
	})
	return this.trie.Hash()
}

// Write the DB
func (this *EthDataStore) Commit() error {
	this.latestRoot, this.nodeBuffer = this.trie.Commit(false)                                                                // Finalized the trie
	if err := this.ethdb.Update(this.latestRoot, types.EmptyRootHash, trienode.NewWithNodeSet(this.nodeBuffer)); err != nil { // Move to DB dirty node set
		return err
	}

	if !this.memonly {
		return this.ethdb.Commit(this.latestRoot, false) // Write to DB
	}
	return nil
}

// Place holders
func (this *EthDataStore) UpdateCacheStats([]interface{})  {}
func (this *EthDataStore) Dump() ([]string, []interface{}) { return nil, nil }
func (this *EthDataStore) GetRootHash() [32]byte           { return this.trie.Hash() }
func (this *EthDataStore) Print()                          {}
func (this *EthDataStore) CheckSum() [32]byte              { return [32]byte{} }
func (this *EthDataStore) Query(string, func(string, string) bool) ([]string, [][]byte, error) {
	return nil, nil, nil
}

func (this *EthDataStore) CacheRetrive(key string, valueTransformer func(interface{}) interface{}) (interface{}, error) {
	return nil, nil
}
