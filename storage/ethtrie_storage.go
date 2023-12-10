package storage

import (
	"errors"
	"sync"

	common "github.com/arcology-network/common-lib/common"
	ccmap "github.com/arcology-network/common-lib/container/map"
	"github.com/arcology-network/common-lib/merkle"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	ethdb "github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	trienode "github.com/ethereum/go-ethereum/trie/trienode"
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

	lock  sync.RWMutex
	dbErr error
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
		acctLookup:     ccmap.NewConcurrentMap(),
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

func (this *EthDataStore) Prove(key [20]byte) ([][]byte, error) {
	var proofs proofList
	err := this.worldStateTrie.Prove(key[:], &proofs)

	return proofs, err
}

func (this *EthDataStore) LoadParallelTrie(root [32]byte) (*ethmpt.Trie, error) {
	if this.latestRoot != root {
		return ethmpt.NewParallel(ethmpt.TrieID(this.latestRoot), this.ethdb)
	}
	return this.worldStateTrie, nil
}

// Problem is here, need to load the storage trie first? and use storageKey as well
func (this *EthDataStore) IfExists(key string) bool {
	accesses := ethmpt.AccessListCache{}

	_, accountKey, suffix := ccurlcommon.ParseAccountAddr(key)
	if v, _ := this.acctLookup.Get(accountKey); v != nil {
		return v.(*Account).Has(key) // If the account has the key
	}

	// Not in cache, look up in the trie
	buffer, _ := this.worldStateTrie.ThreadSafeGet([]byte(accountKey), &accesses)
	if len(buffer) == 0 {
		return false // Not found
	}

	if len(suffix) == 0 {
		return true
	}

	var stateAccount types.StateAccount
	rlp.DecodeBytes(buffer, &stateAccount)

	return NewAccount(key, this.diskdbs, stateAccount).Has(key) // Load the account but don't keep it in the cache.
}

func (this *EthDataStore) Inject(key string, value interface{}) error {
	return this.BatchInject([]string{key}, []interface{}{value})
}

func (this *EthDataStore) BatchInject(keys []string, values []interface{}) error {
	for i := 0; i < len(keys); i++ {
		_, key, _ := ccurlcommon.ParseAccountAddr(keys[i])

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

func (this *EthDataStore) GetAccount(accountKey string, accesses *ethmpt.AccessListCache) (*Account, error) {
	if len(accountKey) > 0 {
		if v, _ := this.acctLookup.Get(accountKey); v != nil {
			return v.(*Account), nil
		}
		return this.GetAccountFromTrie(accountKey, accesses)
	}
	return nil, errors.New("Invalid account: " + accountKey)
}

func (this *EthDataStore) GetAccountFromTrie(accountKey string, accesses *ethmpt.AccessListCache) (*Account, error) {
	if len(accountKey) > 0 {
		buffer, err := this.worldStateTrie.Get([]byte(accountKey))
		if err == nil && len(buffer) > 0 { // Not found
			var acctState types.StateAccount
			rlp.DecodeBytes(buffer, &acctState)

			return &Account{
				accountKey,
				acctState,
				common.FilterFirst(this.diskdbs[0].Get(acctState.CodeHash)),
				ethmpt.NewEmptyParallel(this.ethdb),
				this.ethdb,
				this.diskdbs,
				nil,
			}, nil
		}
		return nil, err
	}
	return nil, errors.New("Empty key")
}

func (this *EthDataStore) Retrive(key string, T any) (interface{}, error) {
	accesses := ethmpt.AccessListCache{}
	_, acct, _ := ccurlcommon.ParseAccountAddr(key) // Get the address
	account, err := this.GetAccount(acct, &accesses)
	if account != nil {
		return account.Retrive(key, T) // Get the storage from the key
	}
	return nil, err
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
			_, key, _ := ccurlcommon.ParseAccountAddr(v.First)
			return &key
		})

	accounts := make([]*Account, len(accountKeys))

	numThd := common.IfThen(len(accountKeys) <= 1024, 8, 16)
	common.ParallelForeach(accountKeys, numThd, func(key *string, i int) {
		accesses := ethmpt.AccessListCache{}
		if accounts[i], _ = this.GetAccount(*key, &accesses); accounts[i] == nil {
			accounts[i] = NewAccount(
				*key,
				this.diskdbs,
				EmptyAccountState()) // empty account
		}
	})

	common.Foreach(accounts, func(acct **Account, _ int) {
		this.acctLookup.Set((**acct).addr, *acct)
	}) // Add to cache

	// Update the account storage trie
	common.ParallelForeach(accounts, 16, func(acct **Account, idx int) {
		(*acct).Precommit(common.FromPairs(stateGroups[idx]))
	})

	keys, accts := this.acctLookup.KVs()
	encoded := common.Append(accts, func(acct interface{}) []byte { return acct.(*Account).Encode() })
	this.worldStateTrie.ParallelUpdate(common.Append(keys, func(key string) []byte { return merkle.Sha256{}.Hash([]byte(key)) }), encoded)

	return this.worldStateTrie.Hash()
}

// Write the DB
func (this *EthDataStore) Commit() error {
	this.acctLookup.ParallelForeachDo(func(_, accountTrie interface{}) {
		accountTrie.(*Account).Commit() // Save the account tries to DB
	})

	// Save the world trie to DB
	this.latestRoot, this.nodeBuffer, _ = this.worldStateTrie.Commit(false) // Finalized the trie
	if this.nodeBuffer == nil || len(this.nodeBuffer.Nodes) == 0 {
		return nil
	}

	// DB update
	if err := this.ethdb.Update(this.latestRoot, types.EmptyRootHash, 0, trienode.NewWithNodeSet(this.nodeBuffer), nil); err != nil { // Move to DB dirty node set
		return err
	}

	if err := this.ethdb.Commit(this.latestRoot, false); err != nil {
		return err
	}

	this.worldStateTrie, _ = ethmpt.New(ethmpt.TrieID(this.latestRoot), this.ethdb)
	return nil
}

func (this *EthDataStore) DiskDBs() [16]ethdb.Database {
	return this.diskdbs
}

// Place holders
func (this *EthDataStore) Root() [32]byte                            { return this.latestRoot }
func (this *EthDataStore) Encoder() func(string, interface{}) []byte { return this.encoder }
func (this *EthDataStore) Decoder() func([]byte, any) interface{}    { return this.decoder }
func (this *EthDataStore) EthDB() *ethmpt.Database                   { return this.ethdb }
func (this *EthDataStore) Trie() *ethmpt.Trie                        { return this.worldStateTrie }
func (this *EthDataStore) UpdateCacheStats([]interface{})            {}
func (this *EthDataStore) Dump() ([]string, []interface{})           { return nil, nil }
func (this *EthDataStore) GetRootHash() [32]byte                     { return this.worldStateTrie.Hash() }
func (this *EthDataStore) Print()                                    {}
func (this *EthDataStore) CheckSum() [32]byte                        { return [32]byte{} }
func (this *EthDataStore) Query(string, func(string, string) bool) ([]string, [][]byte, error) {
	return nil, nil, nil
}
