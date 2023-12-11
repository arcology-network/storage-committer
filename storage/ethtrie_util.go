package storage

import (
	common "github.com/arcology-network/common-lib/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

// ethapi "github.com/ethereum/go-ethereum/internal/ethapi"

func commitToDB(trie *ethmpt.Trie, ethdb *ethmpt.Database, block uint64) (*ethmpt.Trie, error) {
	root, nodes, err := trie.Commit(false) // Finalized the trie
	if err != nil {
		return nil, err
	}

	nodes = common.IfThen(nodes == nil, trienode.NewNodeSet(types.EmptyRootHash), nodes)

	// DB update
	if err := ethdb.Update(root, types.EmptyRootHash, block, trienode.NewWithNodeSet(nodes), nil); err != nil { // Move to DB dirty node set
		return nil, err
	}

	if err := ethdb.Commit(root, false); err != nil {
		return nil, err
	}

	// keys, _ := this.acctCache.KVs()
	// for _, k := range keys {
	// acct, err := this.GetAccountFromTrie(k, &ethmpt.AccessListCache{})
	// acctBuffer, err := trie.Get([]byte(k))
	// if err != nil || len(acctBuffer) == 0 {
	// 	panic(err)
	// }
	// for _, state := range stateGroups {
	// 	hash := ethcommon.BytesToHash([]byte(state[0].First))
	// 	if _, err := acct.IsProvable(hash); err != nil {
	// 		panic("err ")
	// 	}
	// }
	// }
	return ethmpt.NewParallel(ethmpt.TrieID(root), ethdb)
}
