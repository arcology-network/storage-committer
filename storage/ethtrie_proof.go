package storage

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	// ethapi "github.com/ethereum/go-ethereum/internal/ethapi"
)

type ProofProvider struct {
	root        [32]byte
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
		root,
		1,
		1,
		store,
		ethdb,
	}, nil
}

func (this *ProofProvider) Root() ethcommon.Hash { this.visits++; return this.root }

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
