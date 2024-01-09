package storage

import (
	"errors"
	"sync"

	common "github.com/arcology-network/common-lib/common"
	ccmap "github.com/arcology-network/common-lib/container/map"
	committercommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethdb "github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	"golang.org/x/crypto/sha3"
)

type EthDataStore struct {
	worldStateTrie *ethmpt.Trie

	AccountCache  *ccmap.ConcurrentMap // Account cache holds the accounts that are being accessed in the current cycle.
	DirtyAccounts []*Account           // Dirty accounts are the accounts that have been updated in the current cycle.

	ethdb   *ethmpt.Database
	diskdbs [16]ethdb.Database

	encoder func(string, interface{}) []byte
	decoder func([]byte, any) interface{}

	lock  sync.RWMutex
	dbErr error
}

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
	return &EthDataStore{
		ethdb:   triedb,
		diskdbs: diskdb,

		AccountCache:  ccmap.NewConcurrentMap(),
		DirtyAccounts: []*Account{},

		worldStateTrie: trie,
		encoder:        Rlp{}.Encode,
		decoder:        Rlp{}.Decode,
	}
}

func NewParallelEthMemDataStore() *EthDataStore {
	diskdbs := [16]ethdb.Database{}
	common.Fill(diskdbs[:], rawdb.NewMemoryDatabase())
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	return NewEthDataStore(ethmpt.NewEmptyParallel(db), db, diskdbs)
}

func NewLevelDBDataStore(dir string) *EthDataStore {
	leveldb, err := rawdb.NewLevelDBDatabase(dir, 256, 16, "temp", false)
	if err != nil {
		return nil
	}

	diskdbs := [16]ethdb.Database{}
	common.Fill(diskdbs[:], leveldb)
	db := ethmpt.NewParallelDatabase(diskdbs, nil)

	return NewEthDataStore(ethmpt.NewEmptyParallel(db), ethmpt.NewParallelDatabase(diskdbs, nil), diskdbs)
}

func (this *EthDataStore) Clear() {}

func (this *EthDataStore) Hash(key string) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(key))
	sum := hasher.Sum(nil)
	return sum
}

func (this *EthDataStore) GetAccountProof(addr []byte) ([][]byte, error) {
	addHash := crypto.Keccak256(addr)

	var proof proofList
	if trie, _ := this.worldStateTrie.Get(addHash); len(trie) > 0 {
		err := this.worldStateTrie.Prove(addHash, &proof)
		return proof, err
	}
	return [][]byte{}, nil
}

func (this *EthDataStore) IsProvable(addr string) ([]byte, error) {
	addrBytes, _ := hexutil.Decode(addr)
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

func (this *EthDataStore) IfExists(key string) bool {
	accesses := ethmpt.AccessListCache{}

	_, accountKey, suffix := committercommon.ParseAccountAddr(key)
	if v, _ := this.AccountCache.Get(accountKey); v != nil {
		return len(key) == committercommon.ETH10_ACCOUNT_FULL_LENGTH+1 || v.(*Account).Has(key) // If the account has the key
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
	if err := rlp.DecodeBytes(buffer, &stateAccount); err != nil {
		return false
	}
	return NewAccount(key, this.diskdbs, stateAccount).Has(key) // Load the account but don't keep it in the cache.
}

func (this *EthDataStore) Inject(key string, value interface{}) error {
	return this.BatchInject([]string{key}, []interface{}{value})
}

func (this *EthDataStore) BatchInject(keys []string, values []interface{}) error {
	acctDict := make(map[string]*Account)

	for i := 0; i < len(keys); i++ {
		_, acctKey, _ := committercommon.ParseAccountAddr(keys[i])

		account, ok := this.AccountCache.Get(acctKey)
		if account != nil {
			account = account.(*Account)
		}

		if v := common.FilterFirst(this.AccountCache.Get(acctKey)); v != nil {
			account = v.(*Account)
		}

		if !ok {
			account = NewAccount(acctKey, this.diskdbs, EmptyAccountState()) // empty account
			this.AccountCache.Set(acctKey, account)
		}

		// v := values[i]
		if values[i] == nil {
			account.(*Account).UpdateAccountTrie([]string{keys[i]}, []interfaces.Type{nil})
		} else {
			account.(*Account).UpdateAccountTrie([]string{keys[i]}, []interfaces.Type{values[i].(interfaces.Type)})
		}
		acctDict[acctKey] = account.(*Account)
	}

	acctKeys, accounts := common.MapKVs(acctDict)
	common.Foreach(accounts, func(acct **Account, i int) {
		this.worldStateTrie.Update([]byte(acctKeys[i]), (**acct).Encode())
	})

	return nil
}

// Get the account from the cache first, if not found, get it from the trie.
func (this *EthDataStore) GetAccount(accountKey string, accesses *ethmpt.AccessListCache) (*Account, error) {
	if len(accountKey) > 0 {
		if v, _ := this.AccountCache.Get(accountKey); v != nil { // Lookup in the cache first
			return v.(*Account), nil
		}
		return this.GetAccountFromTrie(accountKey, accesses)
	}
	return nil, errors.New("Invalid account: " + accountKey)
}

// Get the account from the trie
func (this *EthDataStore) GetAccountFromTrie(accountKey string, accesses *ethmpt.AccessListCache) (*Account, error) {
	if len(accountKey) > 0 {
		acctAddr, _ := hexutil.Decode(accountKey) // Remove the 0x prefix

		keyHash := crypto.Keccak256(acctAddr) // Hash the key string
		buffer, err := this.worldStateTrie.Get(keyHash)
		if err == nil && len(buffer) > 0 { // Not found
			var acctState types.StateAccount
			rlp.DecodeBytes(buffer, &acctState)

			if trie, err := ethmpt.New(ethmpt.TrieID(acctState.Root), this.EthDB()); trie != nil && err == nil {
				return &Account{
					accountKey,
					acctState,
					common.FilterFirst(this.diskdbs[0].Get(acctState.CodeHash)), // code
					trie,
					false,
					this.ethdb,
					this.diskdbs,
					nil,
				}, nil
			}
		}
		return nil, err
	}
	return nil, errors.New("Empty key")
}

func (this *EthDataStore) Retrive(key string, T any) (interface{}, error) {
	accesses := ethmpt.AccessListCache{}
	_, acct, _ := committercommon.ParseAccountAddr(key) // Get the address
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

// Update the account trie
func (this *EthDataStore) Precommit(keys []string, values interface{}) [32]byte {
	if len(keys) == 0 {
		return this.worldStateTrie.Hash()
	}

	// Group the keys and transactions by their account addresses.
	accountKeys, stateGroups := common.GroupBy(common.ToPairs(keys, values.([]interface{})),
		func(v struct {
			First  string
			Second interface{}
		}) *string {
			_, key, _ := committercommon.ParseAccountAddr(v.First)
			return &key
		})

	this.DirtyAccounts = common.Resize(this.DirtyAccounts, len(accountKeys)) // Reset the dirty accounts

	// Load the accounts from the cache or the trie in parallel, ready for update.
	numThd := common.IfThen(len(accountKeys) <= 1024, 8, 16) // 8 threads for small batch fewer than 1024 accounts, 16 threads for larger batches
	common.ParallelForeach(accountKeys, numThd, func(i int, key *string) {
		accesses := ethmpt.AccessListCache{} // This doesn't serve any purpose for now. It is only a place holder, because the parallelized trie update requires it.
		if this.DirtyAccounts[i], _ = this.GetAccount(*key, &accesses); this.DirtyAccounts[i] == nil {
			this.DirtyAccounts[i] = NewAccount( // Create a new account if not found
				*key,
				this.diskdbs,
				EmptyAccountState()) // empty account state
		}
	})

	// Precommit the changes to the accounts and update the account trie.
	common.ParallelForeach(this.DirtyAccounts, 16, func(idx int, acct **Account) {
		(*acct).Precommit(common.FromPairs(stateGroups[idx]))
	})

	// Move dirty accounts to cache, the difference between the cache and dirty accounts is that the
	// cache is for accounts that are being accessed in the current cycle including the newly created, which isn't available in the cache yet
	// The dirty accounts are the accounts that have been updated in the current cycle.
	common.Foreach(this.DirtyAccounts, func(acct **Account, _ int) {
		this.AccountCache.Set((**acct).addr, *acct)
	})

	// Encode the account keys.
	keyBytes := common.Append(this.DirtyAccounts, func(_ int, acct *Account) []byte {
		addr, _ := hexutil.Decode(acct.addr) // Remove the 0x prefix
		return crypto.Keccak256(addr)        // Account keys
	})

	// Encode the account content.
	encoded := common.Append(this.DirtyAccounts, func(_ int, acct *Account) []byte {
		return acct.Encode()
	})

	// Write the account trie to the DB.
	errs := this.worldStateTrie.ParallelUpdate(keyBytes, encoded) // Encoded accounts

	// Return the first error if any.
	if _, err := common.FindFirstIf(errs, func(err error) bool { return err != nil }); err != nil {
		panic("Error in updating the trie: " + (*err).Error())
	}

	//=======================================================================
	// Debug only
	// for _, k := range keys {
	// 	acctBuffer, err := this.worldStateTrie.Get([]byte(k))
	// 	if err != nil || len(acctBuffer) == 0 {
	// 		panic(err)
	// 	}
	// }

	return this.worldStateTrie.Hash()
}

// Write the DB
func (this *EthDataStore) Commit(block uint64) error {
	this.AccountCache.ParallelForeachDo(func(_, accountTrie interface{}) { // Save the account tries to DB
		if err := accountTrie.(*Account).Commit(block); err != nil {
			panic(err)
		}
	})

	common.ParallelForeach(this.DirtyAccounts, 16, func(_ int, acct **Account) {
		if err := (**acct).Commit(block); err != nil {
			panic(err)
		}
	})

	// Debugging only
	// keys, _ := this.DirtyAccounts
	// for _, k := range keys {
	// 	acctBuffer, err := this.worldStateTrie.Get([]byte(k))
	// 	if err != nil || len(acctBuffer) == 0 {
	// 		panic(err)
	// 	}
	// }

	if len(this.DirtyAccounts) == 0 {
		return nil
	}

	var err error
	this.worldStateTrie, err = commitToDB(this.worldStateTrie, this.ethdb, block) // Reload the trie for the next block
	this.DirtyAccounts = this.DirtyAccounts[:0]
	return err
}

func (this *EthDataStore) DiskDBs() [16]ethdb.Database {
	return this.diskdbs
}

// Place holders
func (this *EthDataStore) Root() [32]byte                            { return this.worldStateTrie.Hash() }
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
