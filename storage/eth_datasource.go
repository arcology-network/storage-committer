package ccdb

import (
	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	ethcommon "github.com/arcology-network/evm/common"
	"github.com/arcology-network/evm/core/rawdb"
	"github.com/arcology-network/evm/core/types"
	ethdb "github.com/arcology-network/evm/ethdb"
	ethmpt "github.com/arcology-network/evm/trie"
	trienode "github.com/arcology-network/evm/trie/trienode"
)

type EthMemoryDatasource struct {
	trie       *ethmpt.Trie
	triedb     *ethmpt.Database
	latestRoot ethcommon.Hash
	memonly    bool
	nodeBuffer *trienode.NodeSet
}

func NewEthMemoryDataStore() *EthMemoryDatasource {
	datasource := &EthMemoryDatasource{}
	diskdbs := [16]ethdb.Database{}
	common.Fill(diskdbs[:], rawdb.NewMemoryDatabase())

	datasource.triedb = ethmpt.NewParallelDatabase(diskdbs, nil)
	datasource.trie = ethmpt.NewEmptyParallel(datasource.triedb)
	return datasource
}

func (this *EthMemoryDatasource) Trie() *ethmpt.Trie { return this.trie }

func (this *EthMemoryDatasource) Inject(key string, value interface{}) error {
	return this.trie.Update([]byte(key), value.([]byte))
}

func (this *EthMemoryDatasource) BatchInject(keys []string, values []interface{}) error {
	for i := 0; i < len(keys); i++ {
		this.Inject(keys[i], values[i])
	}
	return nil
}

func (this *EthMemoryDatasource) Retrive(key string) (interface{}, error) {
	v, err := this.trie.Get([]byte(key))
	if err != nil || len(v) == 0 {
		return nil, err
	}
	return v, err

	// if bytes != nil && err == nil { // Get from the cache
	// 	value = this.decoder(bytes)
	// }

}

func (this *EthMemoryDatasource) BatchRetrive(keys []string) []interface{} {
	values := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		values[i], _ = this.trie.Get([]byte(keys[i]))
	}
	return values
}

func (this *EthMemoryDatasource) Precommit(keys []string, values interface{}) {
	keyBytes := codec.Strings(keys).ToBytes()
	this.trie.ParallelUpdate(keyBytes, values.([][]byte))
	this.latestRoot, this.nodeBuffer = this.trie.Commit(false)
}

func (this *EthMemoryDatasource) Commit() error {
	if err := this.triedb.Update(this.latestRoot, types.EmptyRootHash, trienode.NewWithNodeSet(this.nodeBuffer)); err != nil {
		return err
	}

	if !this.memonly {
		return this.triedb.Commit(this.latestRoot, false)
	}
	return nil
}

func (this *EthMemoryDatasource) GetMerkleRoot(keys []string, values interface{}) ethcommon.Hash {
	keyBytes := codec.Strings(keys).ToBytes()
	this.trie.ParallelUpdate(keyBytes, values.([][]byte))
	return this.trie.Hash()
}

// Place holders
func (this *EthMemoryDatasource) UpdateCacheStats([]interface{})  {}
func (this *EthMemoryDatasource) Dump() ([]string, []interface{}) { return nil, nil }
func (this *EthMemoryDatasource) Clear()                          {}
func (this *EthMemoryDatasource) Print()                          {}
func (this *EthMemoryDatasource) CheckSum() [32]byte              { return [32]byte{} }
func (this *EthMemoryDatasource) Query(string, func(string, string) bool) ([]string, [][]byte, error) {
	return nil, nil, nil
}
func (this *EthMemoryDatasource) CacheRetrive(key string, valueTransformer func(interface{}) interface{}) (interface{}, error) {
	return nil, nil
}
