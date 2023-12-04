package storage

import (
	ccmap "github.com/arcology-network/common-lib/container/map"
	ethcommon "github.com/arcology-network/evm/common"
	"github.com/arcology-network/evm/common/hexutil"
	"github.com/arcology-network/evm/core/types"
	"github.com/arcology-network/evm/crypto"
	ethmpt "github.com/arcology-network/evm/trie"
	// ethapi "github.com/arcology-network/evm/internal/ethapi"
)

// For readonly proof generation
func LoadDataStore(ethdb *ethmpt.Database, root [32]byte) (*EthDataStore, error) {
	trie, err := ethmpt.New(ethmpt.TrieID(root), ethdb)
	if trie == nil || err != nil {
		return nil, err
	}

	diskdb := ethmpt.GetBackendDB(ethdb).DBs()
	return &EthDataStore{
		ethdb:          ethmpt.NewParallelDatabase(diskdb, nil),
		diskdbs:        diskdb,
		acctLookup:     ccmap.NewConcurrentMap(),
		worldStateTrie: trie,
		encoder:        Rlp{}.Encode,
		decoder:        Rlp{}.Decode,
	}, nil
}

func GetProof(sourece *EthDataStore, ethdb *ethmpt.Database, acctStr string, storageKeys []string, rootHash [32]byte) (*AccountResult, error) {
	// acct, _ := hex.DecodeString(acctStr)
	acctAddr := string(acctStr)

	datastore, err := LoadDataStore(ethdb, rootHash)
	if datastore == nil || err != nil {
		return nil, err
	}

	// datastore = sourece

	account, err := datastore.GetAccount(acctAddr, new(ethmpt.AccessListCache))
	if err != nil {
		return nil, err
	}

	storageHash := types.EmptyRootHash
	codeHash := account.GetCodeHash()
	storageProof := make([]StorageResult, len(storageKeys))

	storageTrie := account.storageTrie
	if storageTrie != nil {
		storageHash = storageTrie.Hash()
	} else {
		// no storageTrie means the account does not exist, so the codeHash is the hash of an empty bytearray.
		codeHash = crypto.Keccak256Hash(nil)
	}

	for i, hexKey := range storageKeys {
		key, err := decodeHash(hexKey)
		if err != nil {
			return nil, err
		}
		if storageTrie != nil {
			proof, storageError := account.Prove(key) // Get the storage proof
			if storageError != nil {
				return nil, storageError
			}

			v, _ := account.storageTrie.Get(key[:]) // Get the storage value
			storageProof[i] = StorageResult{hexKey, (*hexutil.Big)(ethcommon.BytesToHash(v).Big()), toHexSlice(proof)}
		} else {
			storageProof[i] = StorageResult{hexKey, &hexutil.Big{}, []string{}}
		}
	}

	// create the accountProof
	accountProof, proofErr := datastore.GetAccountProof([]byte(acctAddr)) // Get the account proof
	if proofErr != nil {
		return nil, proofErr
	}

	return &AccountResult{
		Address:      ethcommon.BytesToAddress([]byte(acctAddr)),
		AccountProof: toHexSlice(accountProof),
		Balance:      (*hexutil.Big)(account.StateAccount.Balance),
		CodeHash:     codeHash,
		Nonce:        hexutil.Uint64(account.StateAccount.Nonce),
		StorageHash:  storageHash,
		StorageProof: storageProof,
	}, nil // state.Error()
}

func GetAccountProof(ethdb *ethmpt.Database, acctAddr string, storageKeys []string, rootHash [32]byte) ([][]byte, error) {
	datastore, err := LoadDataStore(ethdb, rootHash)
	if datastore == nil || err != nil {
		return nil, err
	}

	account, err := datastore.GetAccount(acctAddr, new(ethmpt.AccessListCache))
	if account == nil || err != nil {
		return nil, err
	}
	return datastore.GetAccountProof([]byte(acctAddr))
}
