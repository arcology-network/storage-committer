package ethstorage

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	common "github.com/arcology-network/common-lib/common"

	"github.com/arcology-network/common-lib/exp/associative"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethdb "github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/triedb/hashdb"
	"golang.org/x/crypto/sha3"
)

// Account is the structure that holds the account information and a world storage trie, db instance and a disk db instance.
// It is mainly used for updating the storage tries. The disk db has 16 shards, which could be used for parallel updates in theroy.
// All 16 shards can be directed to the same disk db, as long as the disk db has internal thread safety. It is the case for now with the leveldb.
type EthDataStore struct {
	worldStateTrie *ethmpt.Trie

	accountCacheEnabled bool
	accountCache        map[ethcommon.Address]*Account // Account cache holds the accountCache that are being accessed in the current cycle.

	// dirties       []*AccountUpdate // Dirty accountCache are the accountCache that have been updated in the current cycle.
	dirtyAccounts []*Account

	dirtyVals [][]interfaces.Type // Dirty accountCache are the accountCache that have been updated in the current cycle.
	dirtyKeys [][]string          // Dirty accountCache are the accountCache that have been updated in the current cycle.

	ethdb   *ethmpt.Database
	diskdbs [16]ethdb.Database
	queue   chan *EthIndexer

	encoder func(string, interface{}) []byte
	decoder func(string, []byte, any) interface{}

	lock  sync.RWMutex
	dbErr error

	trieDbConfig *hashdb.Config   // The config for the hash db underlying the trie.
	encodedCache *fastcache.Cache // A shared cache holding the encoded account states to be used by different instances of the Eth Database
}

// LoadEthDataStore loads the trie from the database with the root provided.
func LoadEthDataStore(triedb *ethmpt.Database, root [32]byte) (*EthDataStore, error) {
	trie, err := ethmpt.New(ethmpt.TrieID(root), triedb)
	if err != nil {
		return nil, err
	}

	if trie == nil {
		return nil, errors.New("Failed to load the trie from the database with the root provided!")
	}

	diskdb := ethmpt.GetBackendDB(triedb).DBs()
	return NewEthDataStore(trie, triedb, diskdb), nil
}

func NewEthDataStore(trie *ethmpt.Trie, triedb *ethmpt.Database, diskdb [16]ethdb.Database) *EthDataStore {
	trieDbConfig := &hashdb.Config{CleanCacheSize: 1024 * 1024 * 100} // 100MB of the shared cache
	return &EthDataStore{
		ethdb:          triedb,
		diskdbs:        diskdb,
		trieDbConfig:   trieDbConfig,
		encodedCache:   fastcache.New(trieDbConfig.CleanCacheSize),
		accountCache:   map[ethcommon.Address]*Account{},
		worldStateTrie: trie,
		encoder:        Rlp{}.Encode,
		decoder:        Rlp{}.Decode,
	}
}

// NewParallelEthMemDataStore creates a new EthDataStore with a memory database.
func NewParallelEthMemDataStore() *EthDataStore {
	diskdbs := [16]ethdb.Database{}
	slice.Fill(diskdbs[:], rawdb.NewMemoryDatabase())
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	return NewEthDataStore(ethmpt.NewEmptyParallel(db), db, diskdbs)
}

// NewParallelEthMemDataStore creates a new EthDataStore with a memory database.
func NewParallelEthMemDataStoreWithSharedCache(trieDbConfig *hashdb.Config, cleanCache *fastcache.Cache) *EthDataStore {
	diskdbs := [16]ethdb.Database{}
	slice.Fill(diskdbs[:], rawdb.NewMemoryDatabase())
	db := ethmpt.NewParallelDatabaseWithSharedCache(diskdbs, cleanCache, nil)

	return NewEthDataStore(ethmpt.NewEmptyParallel(db), db, diskdbs)
}

// NewLevelDBDataStore creates a new EthDataStore with a leveldb database.
func NewLevelDBDataStore(dir string) *EthDataStore {
	leveldb, err := rawdb.NewLevelDBDatabase(dir, 256, 16, "temp", false)
	if err != nil {
		return nil
	}

	diskdbs := [16]ethdb.Database{}
	slice.Fill(diskdbs[:], leveldb)
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	return NewEthDataStore(ethmpt.NewEmptyParallel(db), ethmpt.NewParallelDatabase(diskdbs, nil), diskdbs)
}

func (this *EthDataStore) EnableAccountCache()  { this.accountCacheEnabled = true }
func (this *EthDataStore) DisableAccountCache() { this.accountCacheEnabled = false }

// func (this *EthDataStore) EnableCache(active bool) { this.cacheActive = active }
// func (this *EthDataStore) ClearCache()             { this.cache.Clear() }

// Preload loads an existing account from the trie and the disk db.
// If the account is not found, it creates a new account with default account state and shared cache.
func (this *EthDataStore) Preload(addr []byte) interface{} {
	// AccessListCache doesn't serve any purpose for now. It is only a place holder, the parallelized trie update requires its presence.
	acct, _ := this.GetAccount(ethcommon.BytesToAddress(addr), common.Reference(ethmpt.AccessListCache{}))
	if acct == nil {
		acct = NewAccountWithSharedCache( // Account not found, create a new account
			ethcommon.BytesToAddress(addr),
			this.diskdbs,
			EmptyAccountState(),
			this.trieDbConfig,
			this.encodedCache) // empty account state
	}
	return acct
}

func (this *EthDataStore) GetNewIndex(store interfaces.Datastore) interface {
	Add([]*univalue.Univalue)
	Clear()
} {
	return NewIndexer(store)
}

func (this *EthDataStore) AccountDict() map[ethcommon.Address]*Account { return this.accountCache }
func (this *EthDataStore) DirtyAccounts() []*Account                   { return this.dirtyAccounts }
func (this *EthDataStore) Clear()                                      {}

func (this *EthDataStore) Hash(key string) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(key))
	sum := hasher.Sum(nil)
	return sum
}

func (this *EthDataStore) GetAccountProof(addr ethcommon.Address) ([]string, error) {
	addrHash := crypto.Keccak256(addr.Bytes())

	var proof proofList
	if trie, _ := this.worldStateTrie.Get(addrHash); len(trie) > 0 {
		err := this.worldStateTrie.Prove(addrHash, &proof)
		// VerifyProof(this.worldStateTrie, proof, addr[:]) // Debugging only
		return proof, err
	}

	return []string{}, nil
}

func (this *EthDataStore) IsAccountProvable(addr string) ([]byte, error) {
	addrBytes, _ := hexutil.Decode(addr) // Decode to remove the 0x prefix
	keyHash := crypto.Keccak256(addrBytes)

	proofs := memorydb.New()
	if trie, _ := this.worldStateTrie.Get(keyHash); len(trie) > 0 {
		if err := this.worldStateTrie.Prove(keyHash, proofs); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Failed to find the proof")
	}

	v, err := ethmpt.VerifyProof(this.Root(), keyHash, proofs)
	if err != nil || len(v) == 0 {
		return v, errors.New("Failed to find the proof")
	}
	return v, nil
}

// Get the account from the cache first, if not found, get it from the trie.
func (this *EthDataStore) IfExists(key string) bool {
	accesses := ethmpt.AccessListCache{}
	_, acctKey, suffix := platform.ParseAccountAddr(key)
	if len(suffix) == 0 {
		return false
	}

	if len(acctKey) == 0 {
		return false
	}

	acctBytes, err := hexutil.Decode(acctKey)
	if err != nil {
		return false
	}

	address := ethcommon.BytesToAddress(acctBytes)
	// address := ethcommon.BytesToAddress([]byte(accountKey))
	if v := this.accountCache[address]; v != nil {
		return len(key) == stgcommcommon.ETH10_ACCOUNT_FULL_LENGTH+1 || v.Has(key) // If the account has the key
	}

	// Not in cache, look up in the trie
	buffer, _ := this.worldStateTrie.ThreadSafeGet([]byte(address[:]), &accesses)
	if len(buffer) == 0 {
		return false // Not found
	}

	if len(suffix) == 0 {
		return true
	}

	var stateAccount types.StateAccount
	if err := rlp.DecodeBytes(buffer, &stateAccount); err != nil {
		return false
	}

	// address = ethcommon.BytesToAddress([]byte(key))
	return NewAccount(address, this.diskdbs, stateAccount).Has(key) // Load the account but don't keep it in the cache.
}

func (this *EthDataStore) Inject(key string, value interface{}) error {
	return this.BatchInject([]string{key}, []interface{}{value})
}

func (this *EthDataStore) BatchInject(keys []string, values []interface{}) error {
	acctDict := make(map[string]*Account)

	for i := 0; i < len(keys); i++ {
		_, acctKey, _ := platform.ParseAccountAddr(keys[i])
		if len(acctKey) == 0 {
			continue
		}

		address := ethcommon.BytesToAddress(hexutil.MustDecode(acctKey))
		// address := ethcommon.BytesToAddress([]byte(acctKey))
		account, ok := this.accountCache[address]
		// if account != nil {
		// 	account = account.(*Account)
		// }

		// if v := common.FilterFirst(this.accountCache[address]); v != nil {
		// 	account = this.accountCache[address]
		// }

		if !ok {
			account = NewAccount(address, this.diskdbs, EmptyAccountState()) // empty account
			this.accountCache[address] = account
		}

		// v := values[i]
		if values[i] == nil {
			account.UpdateAccountTrie([]string{keys[i]}, []interfaces.Type{nil})
		} else {
			account.UpdateAccountTrie([]string{keys[i]}, []interfaces.Type{values[i].(interfaces.Type)})
		}
		acctDict[acctKey] = account
	}

	acctKeys, accountCache := common.MapKVs(acctDict)
	slice.Foreach(accountCache, func(i int, acct **Account) {
		this.worldStateTrie.Update([]byte(acctKeys[i]), (**acct).Encode())
	})

	return nil
}

// Get the account from the cache first, if not found, get it from the trie.
func (this *EthDataStore) GetAccount(address ethcommon.Address, accesses *ethmpt.AccessListCache) (*Account, error) {
	if len(address) > 0 {
		if v, _ := this.accountCache[address]; v != nil { // Lookup in the cache first
			return v, nil
		}
		return this.GetAccountFromTrie(address, accesses)
	}
	return nil, errors.New("Invalid account: " + address.String())
}

// Get the account from the trie
func (this *EthDataStore) GetAccountFromTrie(address ethcommon.Address, accesses *ethmpt.AccessListCache) (*Account, error) {
	// if len(accountKey) > 0 {
	// acctAddr, _ := hexutil.Decode(accountKey) // Remove the 0x prefix

	acctHash := crypto.Keccak256(address.Bytes()) // Hash the key string
	buffer, err := this.worldStateTrie.Get(acctHash)
	if err == nil && len(buffer) > 0 { // Not found
		var acctState types.StateAccount
		rlp.DecodeBytes(buffer, &acctState)

		stgTrie, err := ethmpt.New(ethmpt.TrieID(acctState.Root), this.EthDB()) // Get the storage trie
		if stgTrie != nil && err == nil {
			return &Account{
				address,
				acctState,
				common.FilterFirst(this.diskdbs[0].Get(acctState.CodeHash)), // code
				stgTrie,
				false,
				this.ethdb,
				this.diskdbs,
				nil,
			}, nil
		}
		return nil, err
	}
	return nil, err
}

// Eth storage doesn't keep an object cache, so the cache is not used.
func (this *EthDataStore) RetriveFromStorage(key string, T any) (interface{}, error) {
	return this.Retrive(key, T)
}

// Skip the cache and get from the trie
func (this *EthDataStore) Retrive(key string, T any) (interface{}, error) {
	accesses := ethmpt.AccessListCache{}
	_, acctKey, _ := platform.ParseAccountAddr(key) // Get the address
	if len(acctKey) == 0 {
		return nil, errors.New("Invalid account: " + acctKey)
	}

	acctBytes, err := hexutil.Decode(acctKey)
	if err != nil {
		return nil, errors.New("Invalid account format: " + acctKey)
	}

	// Get the account the key belongs to.
	address := ethcommon.BytesToAddress(acctBytes)
	account, err := this.GetAccount(address, &accesses)

	if account != nil {
		return account.Retrive(key, T) // Get the storage from the key
	}
	return nil, err
}

func (this *EthDataStore) Precommit(arg ...interface{}) [32]byte {
	pairs := arg[0].(*EthIndexer).Get().([]*associative.Pair[*Account, []*univalue.Univalue])
	this.dirtyAccounts = associative.Pairs[*Account, []*univalue.Univalue](pairs).Firsts()
	if len(pairs) == 0 {
		return this.worldStateTrie.Hash() // No updates
	}

	// Need to check if this is necessary or could be moved to the import phase
	slice.Foreach(this.dirtyAccounts, func(_ int, pair **Account) {
		this.accountCache[(**pair).Address()] = (*pair) // Add the account to the cache
	})

	slice.ParallelForeach(pairs, runtime.NumCPU(), func(i int, acctTrans **associative.Pair[*Account, []*univalue.Univalue]) {
		if len((*acctTrans).Second) == 0 {
			return // All removed
		}

		keys, vals := univalue.Univalues((*acctTrans).Second).KVs() // Get all transitions under the same account
		err := this.dirtyAccounts[i].UpdateAccountTrie(keys, vals)
		if err != nil {
			this.dbErr = errors.Join(this.dbErr, err)
		}
	})
	return this.WriteWorldTrie(this.dirtyAccounts)
}

// The WriteWorldTrie writes the updated accounts to the world trie.
func (this *EthDataStore) WriteWorldTrie(dirtyAccounts []*Account) [32]byte {
	encodedAddrs, encodedAcct := [][]byte{}, [][]byte{} // Encode the account key and values
	common.ParallelExecute(
		func() { // Account keys
			encodedAddrs = slice.Transform(dirtyAccounts, func(_ int, acct *Account) []byte {
				return crypto.Keccak256(acct.addr[:]) // Hash the account address
			})
		},
		func() { // Encode the account content.
			encodedAcct = slice.Transform(dirtyAccounts, func(_ int, acct *Account) []byte {
				return acct.Encode() // Encode the account
			})
		},
	)

	// Write the world tree and return the first error if any.
	errs := this.worldStateTrie.ParallelUpdate(encodedAddrs, encodedAcct)
	this.dbErr = errors.Join(this.dbErr, errors.Join(errs...))
	return this.worldStateTrie.Hash()
}

func (this *EthDataStore) Commit(blockNum uint64) error {
	// var err error
	// common.ParallelExecute(
	// 	func() {
	// 		// this.RefreshCache(blockNum)
	// 	},
	// 	func() { err = this.CommitToEthStorage(blockNum) },
	// )
	for len(this.queue) != 0 {
		fmt.Println("Waiting for the job queue to be empty")
		time.Sleep(100 * time.Millisecond)
	}

	err := this.CommitToEthStorage(blockNum)
	this.dirtyAccounts = this.dirtyAccounts[:0]
	this.dirtyKeys = this.dirtyKeys[:0] // Reset the dirties buffer
	this.dirtyVals = this.dirtyVals[:0]
	return err
}

func (this *EthDataStore) CommitToEthStorage(blockNum uint64) error {
	slice.ParallelForeach(this.dirtyAccounts, runtime.NumCPU(), func(_ int, acct **Account) {
		if err := (**acct).Commit(blockNum); err != nil {
			panic(err)
		}
	})

	var err error
	this.worldStateTrie, err = parallelCommitToDB(this.worldStateTrie, this.ethdb, blockNum) // Reload the trie for the next block
	return err
}

func (this *EthDataStore) BatchRetrive(keys []string, T []any) []interface{} {
	values := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		values[i], _ = this.Retrive(keys[i], T[i])
	}
	return values
}

func (this *EthDataStore) DiskDBs() [16]ethdb.Database {
	return this.diskdbs
}

// Place holders
func (this *EthDataStore) Root() [32]byte                                    { return this.worldStateTrie.Hash() }
func (this *EthDataStore) Encoder(any) func(string, interface{}) []byte      { return this.encoder }
func (this *EthDataStore) Decoder(any) func(string, []byte, any) interface{} { return this.decoder }
func (this *EthDataStore) EthDB() *ethmpt.Database                           { return this.ethdb }
func (this *EthDataStore) Trie() *ethmpt.Trie                                { return this.worldStateTrie }
func (this *EthDataStore) UpdateCacheStats([]interface{})                    {}
func (this *EthDataStore) GetRootHash() [32]byte                             { return this.worldStateTrie.Hash() }
func (this *EthDataStore) Print()                                            {}
func (this *EthDataStore) CheckSum() [32]byte                                { return [32]byte{} }
func (this *EthDataStore) Query(string, func(string, string) bool) ([]string, [][]byte, error) {
	return nil, nil, nil
}
