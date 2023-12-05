package storage

import (
	ethcommon "github.com/arcology-network/evm/common"
	"github.com/arcology-network/evm/common/hexutil"
	"github.com/arcology-network/evm/core/types"
	"github.com/arcology-network/evm/crypto"
	ethmpt "github.com/arcology-network/evm/trie"
	// ethapi "github.com/arcology-network/evm/internal/ethapi"
)

type MerkleProof struct {
	DataStore *EthDataStore
	Ethdb     *ethmpt.Database
}

func NEwMerkleProof(ethdb *ethmpt.Database, root [32]byte) (*MerkleProof, error) {
	return &MerkleProof{
		LoadEthDataStore(ethdb, root),
		ethdb,
	}, nil
}

func (this *MerkleProof) GetProof(acctStr string, storageKeys []string) (*AccountResult, error) {
	acctAddr := string(acctStr)
	account, err := this.DataStore.GetAccount(acctAddr, new(ethmpt.AccessListCache))
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
	accountProof, proofErr := this.DataStore.GetAccountProof([]byte(acctAddr)) // Get the account proof
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
