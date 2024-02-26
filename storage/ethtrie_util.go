package storage

import (
	"github.com/arcology-network/common-lib/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

// ethapi "github.com/ethereum/go-ethereum/internal/ethapi"

func commitToDB(trie *ethmpt.Trie, ethdb *ethmpt.Database, block uint64) (*ethmpt.Trie, error) {
	root, nodes, err := trie.Commit(false) // Finalized the trie
	if err != nil {
		return nil, err
	}

	if nodes != nil {
		if err := ethdb.Update(root, types.EmptyRootHash, block, trienode.NewWithNodeSet(nodes), nil); err != nil { // Move to DB dirty node set
			return nil, err
		}

		if err := ethdb.Commit(root, false); err != nil {
			return nil, err
		}
	}
	return ethmpt.NewParallel(ethmpt.TrieID(root), ethdb)
}

func ParallelCommitToDB(trie *ethmpt.Trie, ethdb *ethmpt.Database, block uint64) (*ethmpt.Trie, error) {
	root, nodes, err := trie.Commit(false) // Finalized the trie
	if err != nil {
		return nil, err
	}

	if nodes != nil {
		if err := ethdb.Update(root, types.EmptyRootHash, block, trienode.NewWithNodeSet(nodes), nil); err != nil { // Move to DB dirty node set
			return nil, err
		}

		if err := ethdb.Commit(root, false); err != nil {
			return nil, err
		}
	}
	return ethmpt.NewParallel(ethmpt.TrieID(root), ethdb)
}

func ProofArrayToDB(proofs []string) (*memorydb.Database, error) {
	proofDB := memorydb.New()
	for i := 0; i < len(proofs); i++ {
		proofBytes := hexutil.MustDecode(proofs[i])

		keyBytes := common.IfThen(len(proofBytes) >= 32, crypto.Keccak256([]byte(proofBytes)), proofBytes)
		if err := proofDB.Put(keyBytes, proofBytes); err != nil {
			return nil, err
		}
	}
	return proofDB, nil
}

func VerifyProof(rootHash ethcommon.Hash, proof []string, addr []byte) {
	proofDB, _ := ProofArrayToDB(proof)
	data, err := ethmpt.VerifyProof(rootHash, crypto.Keccak256(addr), proofDB)
	if err != nil || len(data) == 0 {
		panic(err)
	}
}
