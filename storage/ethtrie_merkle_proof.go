package storage

import (
	"encoding/hex"
	"fmt"
	"strings"

	ccmap "github.com/arcology-network/common-lib/container/map"
	ethcommon "github.com/arcology-network/evm/common"
	"github.com/arcology-network/evm/common/hexutil"
	"github.com/arcology-network/evm/core/types"
	"github.com/arcology-network/evm/crypto"
	ethmpt "github.com/arcology-network/evm/trie"
	// ethapi "github.com/arcology-network/evm/internal/ethapi"
)

type AccountResult struct {
	Address      ethcommon.Address `json:"address"`
	AccountProof []string          `json:"accountProof"`
	Balance      *hexutil.Big      `json:"balance"`
	CodeHash     ethcommon.Hash    `json:"codeHash"`
	Nonce        hexutil.Uint64    `json:"nonce"`
	StorageHash  ethcommon.Hash    `json:"storageHash"`
	StorageProof []StorageResult   `json:"storageProof"`
}

type StorageResult struct {
	Key   string       `json:"key"`
	Value *hexutil.Big `json:"value"`
	Proof []string     `json:"proof"`
}

type proofList [][]byte

func (n *proofList) Put(key []byte, value []byte) error {
	*n = append(*n, value)
	return nil
}

func (n *proofList) Delete(key []byte) error {
	panic("not supported")
}

// toHexSlice creates a slice of hex-strings based on []byte.
func toHexSlice(b [][]byte) []string {
	r := make([]string, len(b))
	for i := range b {
		r[i] = hexutil.Encode(b[i])
	}
	return r
}

func decodeHash(s string) (ethcommon.Hash, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	if (len(s) & 1) > 0 {
		s = "0" + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("hex string invalid")
	}
	if len(b) > 32 {
		return ethcommon.Hash{}, fmt.Errorf("hex string too long, want at most 32 bytes")
	}
	return ethcommon.BytesToHash(b), nil
}

// For readonly proof generation
func LoadDataStore(ethdb *ethmpt.Database, root [32]byte) (*EthDataStore, error) {
	trie, err := ethmpt.New(ethmpt.TrieID(root), ethdb)
	if trie == nil || err != nil {
		return nil, err
	}

	return &EthDataStore{
		ethdb: ethdb,
		// diskdbs:        ,
		acctLookup:     ccmap.NewConcurrentMap(),
		worldStateTrie: trie,
		encoder:        Rlp{}.Encode,
		decoder:        Rlp{}.Decode,
	}, nil
}

func GetProof(ethdb *ethmpt.Database, address ethcommon.Address, storageKeys []string, rootHash [32]byte) (*AccountResult, error) {
	datastore, err := LoadDataStore(ethdb, rootHash)
	if datastore == nil || err != nil {
		return nil, err
	}

	account, err := datastore.GetAccount(string(address[:]), new(ethmpt.AccessListCache))
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
			proof, storageError := account.Prove(key)
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
	accountProof, proofErr := datastore.Prove(address)
	if proofErr != nil {
		return nil, proofErr
	}

	return &AccountResult{
		Address:      address,
		AccountProof: toHexSlice(accountProof),
		Balance:      (*hexutil.Big)(account.StateAccount.Balance),
		CodeHash:     codeHash,
		Nonce:        hexutil.Uint64(account.StateAccount.Nonce),
		StorageHash:  storageHash,
		StorageProof: storageProof,
	}, nil // state.Error()
}
