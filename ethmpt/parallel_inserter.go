package merklepatriciatrie

import (
	"github.com/arcology-network/common-lib/common"
	ethrlp "github.com/arcology-network/concurrenturl/ethrlp"
	"github.com/arcology-network/evm/crypto"
)

type ParallelInserter struct{}

func (ParallelInserter) Insert(trie *Trie, keys [][]byte, values [][]byte) []byte {
	branches := make([][]byte, 17)
	inserters := func(start, end, index int, args ...interface{}) {
		for j := 0; j < len(keys); j++ {
			nibble := Nibble(byte(keys[j][0] >> 4))
			if int(nibble) == start {
				trie.Put(keys[j], values[j])
				// counter++
			}
		}

		branches[start] = EmptyNodeHash
		if trie.root.(*BranchNode).Branches[start] != nil {
			branches[start] = trie.root.(*BranchNode).Branches[start].Hash()
		}
	}
	common.ParallelWorker(16, 16, inserters)

	return crypto.Keccak256(ethrlp.Bytes{}.Encode(branches))
}
