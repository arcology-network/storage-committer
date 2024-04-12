/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package ethstorage

import (
	"errors"
	"runtime"

	async "github.com/arcology-network/common-lib/async"
	"github.com/arcology-network/common-lib/exp/associative"
	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
)

type AsyncWriter struct {
	*async.Pipeline[intf.Indexer[*univalue.Univalue]]
	*EthIndexer
	ethStore *EthDataStore
	Err      error
}

func NewAsyncWriter(ethStore *EthDataStore) *AsyncWriter {
	// version := uint64(0) // TODO: get the block number from the block header
	ethIdxer := NewEthIndexer(ethStore, 0)

	pipe := async.NewPipeline(
		4,
		10,
		// The function updates and storage tries and the world trie without writing to the db.
		func(indexer intf.Indexer[*univalue.Univalue]) (intf.Indexer[*univalue.Univalue], bool) {
			ethIdxer := indexer.(*EthIndexer)
			ethIdxer.Finalize()

			pairs := ethIdxer.UnorderedIndexer.Values()
			ethIdxer.dirtyAccounts = associative.Pairs[*Account, []*univalue.Univalue](pairs).Firsts()

			// Need to check if this is necessary or could be moved to the import phase
			slice.Foreach(ethIdxer.dirtyAccounts, func(_ int, pair **Account) {
				ethStore.accountCache[(**pair).Address()] = (*pair) // Add the account to the cache
			})

			slice.ParallelForeach(pairs, runtime.NumCPU(), func(i int, acctTrans **associative.Pair[*Account, []*univalue.Univalue]) {
				if len((*acctTrans).Second) == 0 {
					return // All removed
				}

				keys, vals := univalue.Univalues((*acctTrans).Second).KVs() // Get all transitions under the same account
				err := ethIdxer.dirtyAccounts[i].UpdateAccountTrie(keys, vals)
				if err != nil {
					ethStore.dbErr = errors.Join(ethStore.dbErr, err)
				}
			})
			ethStore.WriteWorldTrie(ethIdxer.dirtyAccounts) // Update the world trie
			return ethIdxer, true
		},

		// This function actually writes the data to the db
		func(indexer intf.Indexer[*univalue.Univalue]) (intf.Indexer[*univalue.Univalue], bool) {
			ethIdxer := indexer.(*EthIndexer)
			ethIdxer.err = ethStore.WriteToEthStorage(ethIdxer.version, ethIdxer.dirtyAccounts) // Write to the db
			return indexer, true
		},
	)

	return &AsyncWriter{
		Pipeline:   pipe.Start(),
		EthIndexer: ethIdxer,
		ethStore:   ethStore,
		Err:        ethIdxer.err,
	}
}

// // Add a batch of univalues to the indexer of the async writer
func (this *AsyncWriter) Add(univ []*univalue.Univalue) *AsyncWriter {
	if len(univ) == 0 {
		this.EthIndexer.Finalize()
		this.Pipeline.Push(this.EthIndexer) // push the indexer to the processor stream
	} else {
		this.EthIndexer.Add(univ)
	}
	return this
}

func (this *AsyncWriter) Await() { this.Pipeline.Await() }
