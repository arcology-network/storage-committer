package merklepatriciatrie

import (
	"fmt"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/merkle"
)

func TestSimplePerformance(t *testing.T) {
	keys := make([][]byte, 10000)
	data := make([][]byte, len(keys))
	// t0 := time.Now()
	for i := 0; i < len(data); i++ {
		keys[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i)))
		data[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i)))
	}
	// fmt.Println("100,000 keys "+fmt.Sprint(len(data)), time.Since(t0))

	// common.Reverse(&keys)
	// common.Reverse(&data)

	t0 := time.Now()
	trie := NewTrie()
	for i := 0; i < len(data); i++ {
		trie.Put(keys[i], data[i])
	}
	trie.Hash()
	fmt.Println("trie put "+fmt.Sprint(len(data)), time.Since(t0), trie.Hash())

	for i := 0; i < 3; i++ {
		if _, ok := trie.Prove(keys[i]); !ok {
			t.Error("Error: Proof not found")
			return
		}
	}

	// t0 = time.Now()
	// paraTrie := NewParallelMerkles()
	// for i := 0; i < len(data); i++ {
	// 	paraTrie.Put(keys[i], data[i])
	// }
	// trie.Hash()
	// fmt.Println("paraTrie put "+fmt.Sprint(len(data)), time.Since(t0))

	// t0 = time.Now()
	// NewParallelMerkles().BatchPut(keys, data)
	// fmt.Println("paraTrie BatchPut "+fmt.Sprint(len(data)), time.Since(t0))

	t0 = time.Now()
	trie = NewTrie()
	ParallelInserter{}.Insert(trie, keys, data)
	// h := trie.Hash()
	fmt.Println("ParallelInserter put "+fmt.Sprint(len(data)), time.Since(t0), " Hash:")
	fmt.Print(trie.Hash())

}
