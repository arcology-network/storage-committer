package storage

import (
	"errors"
	"sync"

	common "github.com/arcology-network/common-lib/common"
	committercommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	ethcommon "github.com/ethereum/go-ethereum/common"
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

	AccountCache  map[ethcommon.Address]*Account // Account cache holds the accounts that are being accessed in the current cycle.
	DirtyAccounts []*Account                     // Dirty accounts are the accounts that have been updated in the current cycle.

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

		AccountCache:  map[ethcommon.Address]*Account{},
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

func (this *EthDataStore) IfExists(key string) bool {
	accesses := ethmpt.AccessListCache{}
	_, acctKey, suffix := committercommon.ParseAccountAddr(key)
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
	if v, _ := this.AccountCache[address]; v != nil {
		return len(key) == committercommon.ETH10_ACCOUNT_FULL_LENGTH+1 || v.Has(key) // If the account has the key
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
		_, acctKey, _ := committercommon.ParseAccountAddr(keys[i])
		if len(acctKey) == 0 {
			continue
		}

		address := ethcommon.BytesToAddress(hexutil.MustDecode(acctKey))
		// address := ethcommon.BytesToAddress([]byte(acctKey))
		account, ok := this.AccountCache[address]
		// if account != nil {
		// 	account = account.(*Account)
		// }

		// if v := common.FilterFirst(this.AccountCache[address]); v != nil {
		// 	account = this.AccountCache[address]
		// }

		if !ok {
			account = NewAccount(address, this.diskdbs, EmptyAccountState()) // empty account
			this.AccountCache[address] = account
		}

		// v := values[i]
		if values[i] == nil {
			account.UpdateAccountTrie([]string{keys[i]}, []interfaces.Type{nil})
		} else {
			account.UpdateAccountTrie([]string{keys[i]}, []interfaces.Type{values[i].(interfaces.Type)})
		}
		acctDict[acctKey] = account
	}

	acctKeys, accounts := common.MapKVs(acctDict)
	common.Foreach(accounts, func(i int, acct **Account) {
		this.worldStateTrie.Update([]byte(acctKeys[i]), (**acct).Encode())
	})

	return nil
}

// Get the account from the cache first, if not found, get it from the trie.
func (this *EthDataStore) GetAccount(address ethcommon.Address, accesses *ethmpt.AccessListCache) (*Account, error) {
	if len(address) > 0 {
		if v, _ := this.AccountCache[address]; v != nil { // Lookup in the cache first
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

	keyHash := crypto.Keccak256(address.Bytes()) // Hash the key string
	buffer, err := this.worldStateTrie.Get(keyHash)
	if err == nil && len(buffer) > 0 { // Not found
		var acctState types.StateAccount
		rlp.DecodeBytes(buffer, &acctState)

		if trie, err := ethmpt.New(ethmpt.TrieID(acctState.Root), this.EthDB()); trie != nil && err == nil {
			return &Account{
				address,
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

func (this *EthDataStore) Retrive(key string, T any) (interface{}, error) {
	accesses := ethmpt.AccessListCache{}
	_, acctKey, _ := committercommon.ParseAccountAddr(key) // Get the address
	if len(acctKey) == 0 {
		return nil, errors.New("Invalid account: " + acctKey)
	}

	acctBytes, err := hexutil.Decode(acctKey)
	if err != nil {
		return nil, errors.New("Invalid account format: " + acctKey)
	}

	address := ethcommon.BytesToAddress(acctBytes)
	// address := ethcommon.BytesToAddress([]byte(acctKey))
	account, err := this.GetAccount(address, &accesses)
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
		if len(*key) == 0 {
			return
		}

		acctBytes, err := hexutil.Decode(*key)
		if err != nil {
			return
		}

		address := ethcommon.BytesToAddress(acctBytes)

		// address := ethcommon.BytesToAddress(hexutil.MustDecode(*key))
		if this.DirtyAccounts[i], _ = this.GetAccount(address, &accesses); this.DirtyAccounts[i] == nil {
			this.DirtyAccounts[i] = NewAccount( // Create a new account if not found
				address,
				this.diskdbs,
				EmptyAccountState()) // empty account state
		}
	})

	common.RemoveIf(&this.DirtyAccounts, func(acct *Account) bool { return acct == nil }) // Remove the nil accounts

	// Precommit the changes to the accounts and update the account storage trie.
	common.ParallelForeach(this.DirtyAccounts, 16, func(idx int, acct **Account) {
		(*acct).Precommit(common.FromPairs(stateGroups[idx]))
	})

	// Move dirty accounts to cache, the difference between the cache and dirty accounts is that the
	// cache is for accounts that are being accessed in the current cycle including the newly created, which isn't available in the cache yet
	// The dirty accounts are the accounts that have been updated in the current cycle.
	common.Foreach(this.DirtyAccounts, func(_ int, acct **Account) {
		this.AccountCache[(**acct).addr] = *acct
	})

	// Encode the account addresses.
	acctBytes := common.Append(this.DirtyAccounts, func(_ int, acct *Account) []byte {
		// addr, _ := hexutil.Decode(acct.addr) // Remove the 0x prefix
		return crypto.Keccak256(acct.addr[:]) // Account keys
	})

	// Encode the account content.
	encoded := common.Append(this.DirtyAccounts, func(_ int, acct *Account) []byte {
		return acct.Encode()
	})

	// Write the account trie to the DB.
	errs := this.worldStateTrie.ParallelUpdate(acctBytes, encoded) // Encoded accounts

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
	common.MapForeach(this.AccountCache, func(_ ethcommon.Address, acct **Account) {
		if err := (**acct).Commit(block); err != nil {
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
func (this *EthDataStore) GetRootHash() [32]byte                     { return this.worldStateTrie.Hash() }
func (this *EthDataStore) Print()                                    {}
func (this *EthDataStore) CheckSum() [32]byte                        { return [32]byte{} }
func (this *EthDataStore) Query(string, func(string, string) bool) ([]string, [][]byte, error) {
	return nil, nil, nil
}
