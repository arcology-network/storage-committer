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

	"github.com/arcology-network/common-lib/exp/associative"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/storage-committer/type/univalue"
)

type EthStorageWriter struct {
	*EthIndexer
	buffer   []*EthIndexer
	ethStore *EthDataStore
	Err      error
}

func NewEthStorageWriter(ethStore *EthDataStore, version int64) *EthStorageWriter {
	return &EthStorageWriter{
		EthIndexer: NewEthIndexer(ethStore, version),
		ethStore:   ethStore,
		buffer:     []*EthIndexer{},
	}
}

func (this *EthStorageWriter) Precommit() {
	this.EthIndexer.Finalize() // Remove the nil transitions
	this.buffer = append(this.buffer, this.EthIndexer)

	pairs := this.EthIndexer.UnorderedIndexer.Values()
	this.EthIndexer.dirtyAccounts = (associative.Pairs[*Account, []*univalue.Univalue])(pairs).Firsts()

	// Need to check if this is necessary or could be moved to the import phase
	slice.Foreach(this.EthIndexer.dirtyAccounts, func(_ int, pair **Account) {
		this.ethStore.accountCache[(**pair).Address()] = (*pair) // Add the account to the cache
	})

	slice.ParallelForeach(pairs, runtime.NumCPU(), func(i int, acctTrans **associative.Pair[*Account, []*univalue.Univalue]) {
		if len((*acctTrans).Second) == 0 {
			return // All removed
		}

		keys, vals := univalue.Univalues((*acctTrans).Second).KVs() // Get all transitions under the same account
		err := this.EthIndexer.dirtyAccounts[i].UpdateAccountTrie(keys, vals)
		if err != nil {
			this.ethStore.dbErr = errors.Join(this.ethStore.dbErr, err)
		}
	})

	this.ethStore.WriteWorldTrie(this.EthIndexer.dirtyAccounts) // Update the world trie
	this.EthIndexer = NewEthIndexer(this.ethStore, -1)          // Reset the indexer with a default version number.
}

// Signals a block is completed, time to write to the db.
func (this *EthStorageWriter) Commit(version uint64) {
	mergedIdxer := new(EthIndexer).Merge(this.buffer[:]) // Merge all the indexers together to commit to the db at once.
	this.ethStore.WriteToEthStorage(uint64(mergedIdxer.Version), mergedIdxer.dirtyAccounts)
	this.buffer = this.buffer[:0]
}
