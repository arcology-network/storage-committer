package storage

import (
	"sync"

	"github.com/arcology-network/common-lib/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	// ethapi "github.com/ethereum/go-ethereum/internal/ethapi"
)

// MkerkleProofManager is a manager for merkle proofs. It keeps track of the number of times a merkle
// tree has been accessed and keeps the most recently used merkle trees in memory. It a mkerkle tree isn't
// in memory, it will be loaded from the database.
type MerkleProofManager struct {
	maxCached  int                         // max number of merkle proofs to keep in memory
	merkleDict map[[32]byte]*ProofProvider // The merkle tree for each root.
	db         *ethmpt.Database

	lock sync.Mutex
}

// NewMerkleProofManager creates a new MerkleProofManager, which keeps a cache of merkle trees in memory.
// When the cache is full, the merkle tree with the lowest ratio of visits/totalVisits will be removed.
func NewMerkleProofManager(maxCached int, db *ethmpt.Database) *MerkleProofManager {
	return &MerkleProofManager{
		maxCached:  maxCached,
		merkleDict: map[[32]byte]*ProofProvider{},
		db:         db,
	}
}

// GetProof returns a merkle proof for the given account and storage keys.
func (this *MerkleProofManager) GetProof(rootHash [32]byte, acctStr ethcommon.Address, storageKeys []string) (*AccountResult, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	merkle, _ := this.merkleDict[rootHash]
	if merkle == nil {
		datastore, err := LoadEthDataStore(this.db, rootHash)
		if err != nil {
			return nil, err
		}

		// Create a new merkle tree and add it to the cache.
		merkle = &ProofProvider{totalVisits: 1, visits: 1, DataStore: datastore, Ethdb: this.db}
		this.merkleDict[rootHash] = merkle

		// Check if the cache is full. Clean up the cache if it is full.
		if len(this.merkleDict) > this.maxCached {
			keys, merkles := common.MapKVs(this.merkleDict)

			// The visit ratio is the number of times a merkle tree has been accessed divided by the total number of times all the merkle trees have been accessed.
			ratios := common.Append(merkles, func(_ int, v *ProofProvider) float64 { return float64(v.visits) / float64(v.totalVisits) })

			// The entry has the lowest ratio of visits/totalVisits will be removed.
			idx, _ := common.MinElement(ratios, func(v0, v1 float64) bool { return v0 < v1 })
			delete(this.merkleDict, keys[idx])
		}
	}

	// Increment the number of visits for all the merkle trees by 1.
	common.MapForeach(this.merkleDict, func(_ [32]byte, v **ProofProvider) { (**v).totalVisits++ })
	return merkle.GetProof(acctStr, storageKeys)
}

type ProofProvider struct {
	totalVisits uint64 // Total number of times all the merkle trees have been accessed since this Merkle tree is created.
	visits      int    // Number of times this merkle Merkle has been accessed.
	DataStore   *EthDataStore
	Ethdb       *ethmpt.Database
}

func NewProofProvider(ethdb *ethmpt.Database, root [32]byte) (*ProofProvider, error) {
	store, err := LoadEthDataStore(ethdb, root)
	if err != nil {
		return nil, err
	}

	return &ProofProvider{
		1,
		1,
		store,
		ethdb,
	}, nil
}

// GetProof returns a merkle proof for the given account and storage keys.
// Storage keys have to be in hex format with 0x prefix.
func (this *ProofProvider) GetProof(acctAddr ethcommon.Address, storageKeys []string) (*AccountResult, error) {
	this.visits++

	// Get the account either from the cache or from the database.
	account, err := this.DataStore.GetAccount(acctAddr, new(ethmpt.AccessListCache))
	if err != nil {
		return nil, err
	}

	// Debugging only. Will panic if the proof is invalid.
	// if data, _, err := account.IsStorageProvable(storageKeys[0]); err != nil || len(data) == 0 {
	// 	panic(err)
	// }

	storageHash := account.GetStorageRoot()
	codeHash := account.GetCodeHash()

	// Create the storage proof for each storage key.
	storageProof := make([]StorageResult, len(storageKeys))
	for i, hexKey := range storageKeys {
		key, keyLength, err := decodeHash(hexKey)
		if err != nil {
			return nil, err
		}

		// Output key encoding is a bit special: if the input was a 32-byte hash, it is
		// returned as such. Otherwise, we apply the QUANTITY encoding mandated by the
		// JSON-RPC spec for getProof. This behavior exists to preserve backwards
		// compatibility with older client versions.
		var outputKey string
		if keyLength != 32 {
			outputKey = hexutil.EncodeBig(key.Big())
		} else {
			outputKey = hexutil.Encode(key[:])
		}

		if account.storageTrie == nil {
			storageProof[i] = StorageResult{outputKey, &hexutil.Big{}, []string{}}
			continue
		}

		storageKey := crypto.Keccak256(key.Bytes()) // Calculate the key for retrieving the value from the trie.

		var proof proofList
		if err := account.storageTrie.Prove(storageKey, &proof); err != nil {
			return nil, err
		}
		// VerifyProof(account.storageTrie.Hash(), proof, key[:]) // Debugging only. Will panic if the proof is invalid.

		// Get the value from the storage trie.
		v, _ := account.storageTrie.Get(storageKey)
		storageProof[i] = StorageResult{outputKey, (*hexutil.Big)(ethcommon.BytesToHash(v).Big()), proof}
	}

	// create the account Proof
	accountProof, proofErr := this.DataStore.GetAccountProof(acctAddr) // Get the account proof
	if proofErr != nil {
		return nil, proofErr
	}

	VerifyProof(this.DataStore.worldStateTrie.Hash(), accountProof, acctAddr[:]) // Debugging only. Will panic if the proof is invalid.

	return &AccountResult{
		Address:      acctAddr,
		AccountProof: accountProof,
		Balance:      (*hexutil.Big)(account.StateAccount.Balance),
		CodeHash:     codeHash,
		Nonce:        hexutil.Uint64(account.StateAccount.Nonce),
		StorageHash:  storageHash,
		StorageProof: storageProof,
	}, nil // state.Error()
}
