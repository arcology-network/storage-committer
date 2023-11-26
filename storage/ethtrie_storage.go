package ccdb

import (
	"bytes"
	"sync"

	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	ccmap "github.com/arcology-network/common-lib/container/map"
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
	acctLookup     *ccmap.ConcurrentMap

	ethdb      *ethmpt.Database
	diskdbs    [16]ethdb.Database
	latestRoot ethcommon.Hash
	nodeBuffer *trienode.NodeSet
	encoder    func(string, interface{}) []byte
	decoder    func([]byte, any) interface{}
}

func NewParallelEthMemDataStore() *EthDataStore {
	diskdbs := [16]ethdb.Database{}
	common.Fill(diskdbs[:], rawdb.NewMemoryDatabase())
	db := ethmpt.NewParallelDatabase(diskdbs, nil)
	paraTrie := ethmpt.NewEmptyParallel(db)

	return &EthDataStore{
		ethdb:          db,
		diskdbs:        diskdbs,
		acctLookup:     ccmap.NewConcurrentMap(),
		worldStateTrie: paraTrie,
		encoder:        Rlp{}.Encode,
		decoder:        Rlp{}.Decode,
	}
}

func NewLevelDBDataStore(dir string) *EthDataStore {
	leveldb, err := rawdb.NewLevelDBDatabase(dir, 256, 16, "arcology", false)
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

var lock sync.Mutex

// Problem is here, need to load the storage trie first? and use storageKey as well
func (this *EthDataStore) IfExists(key string) bool {
	accesses := ethmpt.AccessListCache{}

	buffer, _ := this.worldStateTrie.ThreadSafeGet(bytes.Clone([]byte(ccurlcommon.ParseAccountAddr(key))), &accesses)
	if len(buffer) == 0 { // Not found
		return false
	}

	var stateAccount types.StateAccount
	lock.Lock()
	rlp.DecodeBytes(buffer, &stateAccount)
	lock.Unlock()

	// storage trie is still empty
	account := NewAccount(key, this.diskdbs, stateAccount)
	return account.IfExists(key)
}

func (this *EthDataStore) Inject(key string, value interface{}) error {
	return this.BatchInject([]string{key}, []interface{}{value})
}

func (this *EthDataStore) BatchInject(keys []string, values []interface{}) error {
	for i := 0; i < len(keys); i++ {
		key := ccurlcommon.ParseAccountAddr(keys[i])

		account, ok := this.acctLookup.Get(key)
		if account != nil {
			account = account.(*Account)
		}

		if v := common.FilterFirst(this.acctLookup.Get(key)); v != nil {
			account = v.(*Account)
		}

		if !ok {
			account = NewAccount(key, this.diskdbs, EmptyAccountState()) // empty account
			this.acctLookup.Set(key, account)
		}

		// v := values[i]
		if values[i] == nil {
			account.(*Account).UpdateAccountTrie([]string{keys[i]}, []interfaces.Type{nil})
		} else {
			account.(*Account).UpdateAccountTrie([]string{keys[i]}, []interfaces.Type{values[i].(interfaces.Type)})
		}
	}

	this.acctLookup.ForeachDo(func(k, accountTrie interface{}) {
		this.worldStateTrie.Update([]byte(k.(string)), accountTrie.(*Account).Encode())
	})

	// this.worldStateTrie.Hash()
	// this.Precommit(keys, values)
	return nil
}

func (this *EthDataStore) LoadExistingAccount(accountKey string, accesses *ethmpt.AccessListCache) *Account {
	if len(accountKey) > 0 {
		if v, _ := this.acctLookup.Get(accountKey); v != nil {
			return v.(*Account)
		}

		if buffer, err := this.worldStateTrie.ThreadSafeGet([]byte(accountKey), accesses); err == nil && len(buffer) > 0 { // Not found
			var acctState types.StateAccount
			rlp.DecodeBytes(buffer, &acctState)

			return &Account{
				accountKey,
				acctState,
				common.FilterFirst(this.diskdbs[0].Get(acctState.CodeHash)),
				ethmpt.NewEmptyParallel(this.ethdb),
				this.ethdb,
				this.diskdbs,
			}
		}
	}
	return nil
}

func (this *EthDataStore) Retrive(key string, T any) (interface{}, error) {
	accesses := ethmpt.AccessListCache{}
	if account := this.LoadExistingAccount(ccurlcommon.ParseAccountAddr(key), &accesses); account != nil {
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

func (this *EthDataStore) Precommit(keys []string, values interface{}) [32]byte {
	if len(keys) == 0 {
		return this.latestRoot
	}
	accountKeys, stateGroups := common.GroupBy(common.ToPairs(keys, values.([]interface{})),
		func(v struct {
			First  string
			Second interface{}
		}) *string {
			key := ccurlcommon.ParseAccountAddr(v.First)
			return &key
		})

	accounts := make([]*Account, len(accountKeys))
	common.ParallelForeach(accountKeys, 16, func(key *string, i int) {
		accesses := ethmpt.AccessListCache{}
		if accounts[i] = this.LoadExistingAccount(*key, &accesses); accounts[i] == nil {
			accounts[i] = NewAccount(
				*key,
				this.diskdbs,
				EmptyAccountState()) // empty account
		}
	})

	common.Foreach(accounts, func(acct **Account, _ int) {
		this.acctLookup.Set((**acct).addr, *acct)
	}) // Add to cache

	common.ParallelForeach(accounts, 16, func(acct **Account, idx int) {
		(*acct).Precommit(common.FromPairs(stateGroups[idx]))
	})

	keys, accts := this.acctLookup.KVs()
	encoded := common.Append(accts, func(acct interface{}) []byte { return acct.(*Account).Encode() })

	this.worldStateTrie.ParallelUpdate(codec.Strings(keys).ToBytes(), encoded)
	return this.worldStateTrie.Hash()
}

// Write the DB
func (this *EthDataStore) Commit() error {
	this.acctLookup.ParallelForeachDo(func(_, accountTrie interface{}) {
		accountTrie.(*Account).Commit() // Save the account tries to DB
	})

	// Save the world trie to DB
	this.latestRoot, this.nodeBuffer = this.worldStateTrie.Commit(false) // Finalized the trie
	if len(this.nodeBuffer.Nodes) == 0 {
		return nil
	}

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
