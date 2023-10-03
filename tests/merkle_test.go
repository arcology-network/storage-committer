package ccurltest

import (
	"fmt"
	"testing"

	merklepatriciatrie "github.com/arcology-network/concurrenturl/ethmpt"
	rlp "github.com/arcology-network/evm/rlp"
)

func CreateSubtree(keys [][]byte, values [][]byte, offset int) *merklepatriciatrie.ExtensionNode {
	root := merklepatriciatrie.NewBranchNode()
	for i := 0; i < len(keys); i++ {
		n1 := merklepatriciatrie.NewLeafNodeFromNibbles(merklepatriciatrie.FromBytes([]byte(keys[i])[offset+1:]), values[i])
		root.Branches[[]byte(keys[i])[offset]] = n1
	}

	rootNode := merklepatriciatrie.NewExtensionNode(merklepatriciatrie.FromBytes([]byte{}), root)
	rootNode.Path = []merklepatriciatrie.Nibble{merklepatriciatrie.Nibble(0)}
	return rootNode
}

func TestSimpleMerkleConvertor(t *testing.T) {
	v := [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 'a', 'b', 'c', 'd', 'e', 'f'}

	// storage.EthRlp.Encode(v)
	encoded, err := rlp.EncodeToBytes(v)
	fmt.Println(encoded)
	fmt.Println(err)

	k0 := []byte{11, 1, 2, 3}
	v0 := []byte("hello hello hello hello hello hello hello")

	k1 := []byte{12, 1, 2, 3, 4, 6}
	v1 := []byte("world world world world world world world")

	k2 := []byte{12, 1, 2, 3, 4, 7}
	v2 := []byte("world world world world world world world")

	// -----------------------------------------------------------------------------------------------

	// subtrie := CreateSubtree([][]byte{k0, k1}, [][]byte{v0, v1}, 0)

	// fmt.Println("subtree hash:", subtrie.Hash())
	// fmt.Println()
	// -----------------------------------------------------------------------------------------------
	trie := merklepatriciatrie.NewTrie()

	trie.Put(k0, []byte(v0))
	trie.Put(k1, []byte(v1))
	trie.Put(k2, []byte(v2))
	fmt.Println("trie hash: ", trie.Hash())

	trie.Prove(k0)
}
