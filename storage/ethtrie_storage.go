package ccdb

import (
	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	ethcommon "github.com/arcology-network/evm/common"
	"github.com/arcology-network/evm/core/rawdb"
	"github.com/arcology-network/evm/core/types"
	ethdb "github.com/arcology-network/evm/ethdb"
	"github.com/arcology-network/evm/rlp"
	ethmpt "github.com/arcology-network/evm/trie"
	trienode "github.com/arcology-network/evm/trie/trienode"
	"golang.org/x/crypto/sha3"
)

type EthDataStore struct {
	worldStateTrie *ethmpt.Trie
	acctDict       map[string]*Account
	ethdb          *ethmpt.Database
	diskdbs        [16]ethdb.Database
	latestRoot     ethcommon.Hash
	nodeBuffer     *trienode.NodeSet
	encoder        func(string, interface{}) []byte
	decoder        func([]byte, any) interface{}
}

func NewParallelEthMemDataStore() *EthDataStore {
	diskdbs := [16]ethdb.Database{}
	common.Fill(diskdbs[:], rawdb.NewMemoryDatabase())
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	paraTrie := ethmpt.NewEmptyParallel(db)
	return &EthDataStore{
		ethdb:          db,
		diskdbs:        diskdbs,
		acctDict:       make(map[string]*Account),
		worldStateTrie: paraTrie,
		encoder:        Rlp{}.Encode,
		decoder:        Rlp{}.Decode,
	}
}

func NewLevelDBDataStore(dir string) *EthDataStore {
	leveldb, err := rawdb.NewLevelDBDatabase(dir, 256, 16, "temp", false)
	if err != nil {
		return nil
	}

	diskdbs := [16]ethdb.Database{}
	common.Fill(diskdbs[:], leveldb)
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	paraTrie := ethmpt.NewEmptyParallel(db)
	return &EthDataStore{
		ethdb:          db,
		diskdbs:        diskdbs,
		worldStateTrie: paraTrie,
		encoder:        Rlp{}.Encode,
		decoder:        Rlp{}.Decode,
	}
}

func (this *EthDataStore) Clear() {
	var err error
	this.worldStateTrie, err = ethmpt.NewParallel(ethmpt.TrieID(this.latestRoot), this.ethdb) // reopen the trie for future use
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

// For readonly proof generation
func (this *EthDataStore) LoadTrie(root [32]byte) (*ethmpt.Trie, error) {
	if this.latestRoot != root {
		return ethmpt.New(ethmpt.TrieID(this.latestRoot), this.ethdb)
	}
	return this.worldStateTrie, nil
}

func (this *EthDataStore) LoadParallelTrie(root [32]byte) (*ethmpt.Trie, error) {
	if this.latestRoot != root {
		return ethmpt.NewParallel(ethmpt.TrieID(this.latestRoot), this.ethdb)
	}
	return this.worldStateTrie, nil
}

func (this *EthDataStore) Root() [32]byte { return this.latestRoot }

func (this *EthDataStore) Encoder() func(string, interface{}) []byte { return this.encoder }
func (this *EthDataStore) Decoder() func([]byte, any) interface{}    { return this.decoder }

// Problem is here, need to load the storage trie first? and use storageKey as well
func (this *EthDataStore) IfExists(key string) bool {
	buffer, _ := this.worldStateTrie.Get([]byte(ccurlcommon.ParseAccountAddr(key)))
	if len(buffer) == 0 { // Not found
		return false
	}

	var stateAccount types.StateAccount
	rlp.DecodeBytes(buffer, &stateAccount)

	// storage trie is still empty
	account := NewAccount([]byte(key), this.ethdb, this.diskdbs[0], stateAccount)
	return account.IfExists(key)
}

func (this *EthDataStore) Inject(key string, value interface{}) error {
	return this.BatchInject([]string{key}, []interface{}{value})
}

func (this *EthDataStore) BatchInject(keys []string, values []interface{}) error {
	for i := 0; i < len(keys); i++ {
		key := ccurlcommon.ParseAccountAddr(keys[i])
		account, ok := this.acctDict[key]
		if !ok {
			account = NewAccount([]byte(key), this.ethdb, this.diskdbs[0], EmptyAccountState()) // empty account
			this.acctDict[key] = account
		}

		// v := values[i]
		if values[i] == nil {
			account.updateAccountTrie([]string{keys[i]}, []interfaces.Type{nil})
		} else {
			account.updateAccountTrie([]string{keys[i]}, []interfaces.Type{values[i].(interfaces.Type)})
		}
	}

	for k, account := range this.acctDict {
		this.worldStateTrie.Update([]byte(k), account.Encode())
	}
	// this.worldStateTrie.Hash()

	// this.Precommit(keys, values)
	return nil
}

func (this *EthDataStore) LoadAccount(accountAddr string) *Account {
	if len(accountAddr) == 0 {
		return nil
	}

	account, ok := this.acctDict[accountAddr]
	if !ok { // Not in cache yet.
		if buffer, err := this.worldStateTrie.Get([]byte(accountAddr)); err == nil && len(buffer) > 0 { // Not found
			var acctState types.StateAccount
			rlp.DecodeBytes(buffer, &acctState)

			code, _ := this.diskdbs[0].Get(acctState.CodeHash)
			account = &Account{
				[]byte(accountAddr),
				acctState,
				code,
				ethmpt.NewEmptyParallel(this.ethdb),
				this.ethdb,
				this.diskdbs[0],
			}
		}
	}
	return account
}

func (this *EthDataStore) Retrive(key string, T any) (interface{}, error) {
	if account := this.LoadAccount(ccurlcommon.ParseAccountAddr(key)); account != nil {
		return account.Retrive(key, T)
	}
	return nil, nil
}

func (this *EthDataStore) BatchRetrive(keys []string, T []any) []interface{} {
	values := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		values[i], _ = this.Retrive(keys[i], T[i])
	}
	return values
}

// Encode KVs and write them to the trie.
func (this *EthDataStore) Precommit(keys []string, values interface{}) [32]byte {
	pairGroups := common.GroupBy(common.ToPairs(keys, values.([]interface{})),
		func(v struct {
			First  string
			Second interface{}
		}) *string {
			key := ccurlcommon.ParseAccountAddr(v.First)
			return &key
		})

	for i := 0; i < len(pairGroups); i++ {
		key := ccurlcommon.ParseAccountAddr(pairGroups[i][0].First)
		account, ok := this.acctDict[key]
		if !ok {
			account = NewAccount([]byte(key), this.ethdb, this.diskdbs[0], EmptyAccountState()) // empty account
			this.acctDict[key] = account
		}
		account.Precommit(common.FromPairs(pairGroups[i]))
	}

	for k, account := range this.acctDict {
		this.worldStateTrie.Update([]byte(k), account.Encode())
	}
	return this.worldStateTrie.Hash()
}

// Write the DB
func (this *EthDataStore) Commit() error {
	for _, accountTrie := range this.acctDict {
		accountTrie.Commit()
	}

	this.latestRoot, this.nodeBuffer = this.worldStateTrie.Commit(false)                                                      // Finalized the trie
	if err := this.ethdb.Update(this.latestRoot, types.EmptyRootHash, trienode.NewWithNodeSet(this.nodeBuffer)); err != nil { // Move to DB dirty node set
		return err
	}
	return this.ethdb.Commit(this.latestRoot, false) // Write to DB
}

// Place holders
func (this *EthDataStore) UpdateCacheStats([]interface{})  {}
func (this *EthDataStore) Dump() ([]string, []interface{}) { return nil, nil }
func (this *EthDataStore) GetRootHash() [32]byte           { return this.worldStateTrie.Hash() }
func (this *EthDataStore) Print()                          {}
func (this *EthDataStore) CheckSum() [32]byte              { return [32]byte{} }
func (this *EthDataStore) Query(string, func(string, string) bool) ([]string, [][]byte, error) {
	return nil, nil, nil
}