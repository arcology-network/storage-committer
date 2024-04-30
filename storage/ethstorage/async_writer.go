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
	"math"
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
	Err      error
}

func NewAsyncWriter(ethStore *EthDataStore, version uint64) *AsyncWriter {
	pipe := async.NewPipeline(
		"ethstorage",
		4,
		10,
		// The function updates and storage tries and the world trie without writing to the db.
		func(idxer *EthIndexer, buffer *[]*EthIndexer) ([]*EthIndexer, bool) {
			if *buffer = append(*buffer, idxer); idxer.UnorderedIndexer == nil {
				v := slice.Move(buffer) // Move the buffer to the next function
				return v, true          // Forwards the an array of indexers, including the nil one to the next function
			}

			pairs := idxer.UnorderedIndexer.Values()
			idxer.dirtyAccounts = associative.Pairs[*Account, []*univalue.Univalue](pairs).Firsts()

			// Need to check if this is necessary or could be moved to the import phase
			slice.Foreach(idxer.dirtyAccounts, func(_ int, pair **Account) {
				ethStore.accountCache[(**pair).Address()] = (*pair) // Add the account to the cache
			})

			slice.ParallelForeach(pairs, runtime.NumCPU(), func(i int, acctTrans **associative.Pair[*Account, []*univalue.Univalue]) {
				if len((*acctTrans).Second) == 0 {
					return // All removed
				}

				keys, vals := univalue.Univalues((*acctTrans).Second).KVs() // Get all transitions under the same account
				err := idxer.dirtyAccounts[i].UpdateAccountTrie(keys, vals)
				if err != nil {
					ethStore.dbErr = errors.Join(ethStore.dbErr, err)
				}
			})
			ethStore.WriteWorldTrie(idxer.dirtyAccounts) // Update the world trie

			return nil, false // False means the data is only cached in the buffer provided by the Pipeliner for now.
		},

		// This function actually writes the data to the db
		func(idxer *EthIndexer, buffer *[]*EthIndexer) ([]*EthIndexer, bool) {
			if *buffer = append(*buffer, idxer); idxer.UnorderedIndexer != nil {
				return nil, false // Forwards the an array of indexers, including the nil one to the next function
			}

			// Write to the db
			mergedIdxer := new(EthIndexer).Merge(*buffer) // Merge all the indexers together to commit to the db at once.
			ethStore.WriteToEthStorage(mergedIdxer.Version, mergedIdxer.dirtyAccounts)

			*buffer = (*buffer)[:0] // Clear the buffer
			return nil, false
		},
	)

	return &AsyncWriter{
		Pipeline:   pipe.Start(),
		EthIndexer: NewEthIndexer(ethStore, version),
		ethStore:   ethStore,
	}
}

// Send the data to the downstream processor, this is called for each generation.
// If there are multiple generations, this can be called multiple times before Await.
// Each generation
func (this *AsyncWriter) Precommit() {
	this.EthIndexer.Finalize()                                     // Remove the nil transitions
	this.Pipeline.Push(this.EthIndexer)                            // push the indexer to the processor stream
	this.EthIndexer = NewEthIndexer(this.ethStore, math.MaxUint64) // Reset the indexer with a default version number.
}

// Signals a block is completed, time to write to the db.
func (this *AsyncWriter) Commit(version uint64) {
	this.Pipeline.Push(&EthIndexer{Version: version}) //
	this.Pipeline.Await()
}

// Await commits the data to the state db.
func (this *AsyncWriter) Close() { this.Pipeline.Close() }
