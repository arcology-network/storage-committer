package mkproof

import (
	"fmt"
	"testing"

	"github.com/arcology-network/common-lib/mempool"
	"github.com/arcology-network/common-lib/merkle"
	evmcommon "github.com/arcology-network/evm/common"
	trie "github.com/arcology-network/evm/trie"
)

func TestSimpleMerkleConvertor(t *testing.T) {
	data := [][]byte{
		EthRlp{}.Encode([]byte{10, 1, 2, 3}, []byte("00000 00000 00000 00000 00000 00000 00000")),
		EthRlp{}.Encode([]byte{11, 1, 2, 3, 4, 5, 6}, []byte("11111 11111 11111 11111 11111 11111 11111")),
		EthRlp{}.Encode([]byte{12, 1, 2, 3, 4, 5}, []byte("33333 33333 33333 33333 33333 33333 33333")),
	}
	mkTree := merkle.NewMerkle(16, merkle.Concatenator{}, merkle.Keccak256{})
	mkTree.Init(data, mempool.NewMempool("nodes", func() interface{} { return merkle.NewNode() }))

	proofNodes := mkTree.GetProofNodes(data[0])
	keys, proofs := mkTree.NodesToHashes(proofNodes)

	// encoded, err := rlp.EncodeToBytes(proofs[len(proofs)-1])
	// if err != nil {
	// 	log.Fatal("Error encoding data:", err)
	// }

	mkProof := EthMerkleProof{}
	// lookup := mkProof.ConvertProofs(keys, proofs)
	// merklepatriciatrie.VerifyProof(mkTree.GetRoot(), data[0], lookup)
	v := EthRlp{}.Encode(proofs[0])
	mkProof.Put(keybytesToHex(keys[0]), v)

	root := evmcommon.Hash(mkTree.GetRoot())
	if _, err := trie.VerifyProof(root, data[0], mkProof); err != nil {
		fmt.Println("trie hash: ", err)
	}
}
