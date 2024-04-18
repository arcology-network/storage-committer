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
	"github.com/arcology-network/storage-committer/univalue"
)

type AsyncWriter struct {
	*async.Pipeline[*EthIndexer]
	*EthIndexer
	ethStore *EthDataStore
	version  uint64
	Err      error
}

func NewAsyncWriter(ethStore *EthDataStore, version uint64) *AsyncWriter {
	pipe := async.NewPipeline(
		4,
		10,
		// The function updates and storage tries and the world trie without writing to the db.
		func(indexers ...*EthIndexer) (*EthIndexer, bool) {
			if len(indexers) == 0 || indexers[0] == nil {
				return nil, true // Forwards the an array of indexers, including the nil one to the next function
			}
			ethIdxer := indexers[0]

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
			return ethIdxer, false                          // False means the data is only cached in the buffer provided by the Pipeliner for now.
		},

		// This function actually writes the data to the db
		func(indexers ...*EthIndexer) (*EthIndexer, bool) {
			mergedIdxer := new(EthIndexer).Merge(indexers) // Merge all the indexers together to commit to the db at once.

			// Write to the db
			mergedIdxer.err = ethStore.WriteToEthStorage(mergedIdxer.version, mergedIdxer.dirtyAccounts)
			return mergedIdxer, true
		},
	)

	return &AsyncWriter{
		Pipeline:   pipe.Start(),
		EthIndexer: NewEthIndexer(ethStore, 0),
		ethStore:   ethStore,
		version:    version,
	}
}

// Send the data to the downstream processor, this is called for each generation.
// If there are multiple generations, this can be called multiple times before Await.
// Each generation
func (this *AsyncWriter) Precommit() {
	this.EthIndexer.Finalize()                                              // Remove the nil transitions
	this.Pipeline.Push(this.EthIndexer)                                     // push the indexer to the processor stream
	this.EthIndexer = NewEthIndexer(this.ethStore, this.EthIndexer.version) // Reset the indexer
}

// Signals a block is completed, time to write to the db.
func (this *AsyncWriter) Commit() {
	this.Pipeline.Push(nil)
	this.Pipeline.Await()
}
