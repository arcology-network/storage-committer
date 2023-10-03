package merklepatriciatrie

import (
	"github.com/arcology-network/common-lib/common"
)

type ParallelInserter struct{}

func (ParallelInserter) Insert(trie *Trie, keys [][]byte, values [][]byte) []byte {
	inserters := func(start, end, index int, args ...interface{}) {
		for j := 0; j < len(keys); j++ {
			nibble := Nibble(byte(keys[j][0] >> 4))
			if int(nibble) == start {
				trie.Put(keys[j], values[j])
				// counter++
			}
		}

		if trie.root.(*BranchNode).Branches[start] != nil {
			trie.root.(*BranchNode).Branches[start].Hash()
		}
	}
	common.ParallelWorker(16, 16, inserters)

	// trie.Hash()
	return []byte{}
}
