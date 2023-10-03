package merklepatriciatrie

import (
	"bytes"

	"github.com/arcology-network/common-lib/common"
)

type ParallelMerkle struct {
	RootMerkle   *Trie
	childMerkles [256]*Trie
}

func NewParallelMerkles() *ParallelMerkle {
	parallelMerkle := ParallelMerkle{}
	parallelMerkle.RootMerkle = NewTrie()
	for i := 0; i < len(parallelMerkle.childMerkles); i++ {
		parallelMerkle.childMerkles[i] = NewTrie()
	}
	return &parallelMerkle
}

func (this *ParallelMerkle) Put(key []byte, value []byte) {
	subtrie := (*this).childMerkles[key[0]]
	subtrie.Put(key[1:], value)
}

func (this *ParallelMerkle) Root() []byte { return this.RootMerkle.Hash() }

func (this *ParallelMerkle) BatchPut(keys [][]byte, values [][]byte) *ParallelMerkle {
	for i := 0; i < len(keys); i++ {
		subtrie := (*this).childMerkles[keys[i][0]]
		subtrie.Put(keys[i][1:], values[i])
	}

	for i := 0; i < len(this.childMerkles); i++ {
		this.RootMerkle.Put([]byte{uint8(i)}, this.childMerkles[i].Hash())
	}
	return this
}

func (this *ParallelMerkle) findRange(keys [][]byte) [][2]int {
	ranges := make([][2]int, 0, len(keys))

	offset := 0
	for i := 0; i < len(keys); i++ {
		for j := 0; j < len(this.childMerkles); j++ {
			newOffset, _ := common.FindFirstIf(keys[offset:], func(v []byte) bool { return bytes.Equal(v, []byte{uint8(j)}) })
			ranges = append(ranges, [2]int{offset, newOffset})
			offset = newOffset
		}
	}

	ranges = append(ranges, [2]int{offset, len(keys)})
	return ranges
}
