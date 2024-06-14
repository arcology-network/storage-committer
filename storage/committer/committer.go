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

// Package storagecommitter provides functionality for committing storage changes to url2a datastore.
package statestore

import (
	"github.com/arcology-network/common-lib/common"
	indexer "github.com/arcology-network/common-lib/storage/indexer"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/storage/proxy"
	cache "github.com/arcology-network/storage-committer/storage/writecache"
	"github.com/arcology-network/storage-committer/univalue"

	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
)

// StateCommitter represents a storage committer.
// The main purpose of the StateCommitter is to commit the transitions to the different stores.
type StateCommitter struct {
	readonlyStore intf.ReadOnlyStore
	platform      *platform.Platform

	writers []intf.AsyncWriter[*univalue.Univalue] // db writers

	byPath *indexer.UnorderedIndexer[string, *univalue.Univalue, []*univalue.Univalue]
	byTxID *indexer.UnorderedIndexer[uint32, *univalue.Univalue, []*univalue.Univalue]

	Err error
}

// NewStateCommitter creates a new StateCommitter instance. The stores are the stores that can be committed.
// A Committable store is a pair of an index and a store. The index is used to index the input transitions as they are
// received, and the store is used to commit the indexed transitions. Since multiple store can share the same index, each
// CommittableStore is an indexer and a list of Committable stores.
func NewStateCommitter(readonlyStore intf.ReadOnlyStore, writers []intf.AsyncWriter[*univalue.Univalue]) *StateCommitter {
	return &StateCommitter{
		readonlyStore: readonlyStore,
		platform:      platform.NewPlatform(),

		writers: writers,
		byPath:  PathIndexer(readonlyStore), // By storage path
		byTxID:  TxIndexer(readonlyStore),   // By tx ID
	}
}

// New creates a new StateCommitter instance.
func (this *StateCommitter) New(args ...interface{}) *StateCommitter {
	return &StateCommitter{
		platform: platform.NewPlatform(),
	}
}

// Importer returns the importer of the StateCommitter.
func (this *StateCommitter) Store() intf.ReadOnlyStore { return this.readonlyStore }
func (this *StateCommitter) SetStore(store intf.ReadOnlyStore) { // Testing only
	this.readonlyStore = store
}

// Import imports the given transitions into the StateCommitter.
func (this *StateCommitter) Import(transitions []*univalue.Univalue) *StateCommitter {
	this.byPath.Import(transitions)
	this.byTxID.Import(transitions)

	for _, writer := range this.writers {
		writer.Import(transitions)
	}
	return this
}

// Finalize finalizes the transitions in the StateCommitter.
func (this *StateCommitter) whitelist(txs []uint32) *StateCommitter {
	if len(txs) == 0 {
		return this
	}

	whitelistDict := mapi.FromSlice(txs, func(_ uint32) bool { return true })
	this.byTxID.ParallelForeachDo(func(txid uint32, vec *[]*univalue.Univalue) {
		if _, ok := whitelistDict[uint32(txid)]; !ok {
			for _, v := range *vec {
				v.SetPath(nil) // Mark the transition status, so that it can be removed later.
			}
		}
	})
	return this
}

// Commit commits the transitions to different stores.
func (this *StateCommitter) Finalize(txs []uint32) {
	this.whitelist(txs) // Mark the transitions that are not in the whitelist

	// Finalize all the transitions by merging the transitions
	// for both the ETH storage and the concurrent container transitions
	this.byPath.ParallelForeachDo(func(_ string, v *[]*univalue.Univalue) {
		slice.RemoveIf(v, func(_ int, val *univalue.Univalue) bool { return val.GetPath() == nil }) // Remove conflicting ones.
		if len(*v) > 0 {
			DeltaSequence(*v).Finalize(this.readonlyStore) // Finalize the transitions and flag the merged ones.
		}
	})
	this.byPath.Clear()
	this.byTxID.Clear()
}

// Commit commits the transitions in the StateCommitter.
// 1. For the block write cache, it commits the transitions to the cache.
// 2. For the eth storage, it updates the tries without committing the transitions to the DB
func (this *StateCommitter) Precommit(txs []uint32) [32]byte {
	this.Finalize(txs)
	this.SyncPrecommit()
	this.AsyncPrecommit()
	return [32]byte{}
}

// Only the global write cache needs to be synchronized before the next precommit or commit.
func (this *StateCommitter) SyncPrecommit() {
	slice.ParallelForeach(this.writers, 1,
		func(i int, writer *intf.AsyncWriter[*univalue.Univalue]) {
			if common.IsType[*cache.ExecutionCacheWriter](*writer) {
				(*writer).Precommit()
			}
		})
}

// Only the global write cache needs to be synchronized before the next precommit or commit.
func (this *StateCommitter) AsyncPrecommit() {
	slice.ParallelForeach(this.writers, len(this.writers),
		func(_ int, writer *intf.AsyncWriter[*univalue.Univalue]) {
			if !common.IsType[*cache.ExecutionCacheWriter](*writer) {
				(*writer).Precommit()
			}
		})
}

// Commit commits the transitions to different stores.
func (this *StateCommitter) Commit(blockNum uint64) *StateCommitter {
	this.SyncCommit(blockNum)
	this.AsyncCommit(blockNum)
	return this
}

// Only the global write cache needs to be synchronized before the next precommit.
func (this *StateCommitter) SyncCommit(blockNum uint64) {
	slice.ParallelForeach(this.writers, len(this.writers),
		func(_ int, writer *intf.AsyncWriter[*univalue.Univalue]) {
			if common.IsType[*cache.ExecutionCacheWriter](*writer) || common.IsType[*proxy.LiveCacheWriter](*writer) {
				(*writer).Commit(blockNum)
				return
			}
		})
}

// Only the global write cache needs to be synchronized before the next precommit.
func (this *StateCommitter) AsyncCommit(blockNum uint64) {
	slice.ParallelForeach(this.writers, len(this.writers),
		func(_ int, writer *intf.AsyncWriter[*univalue.Univalue]) {
			if !common.IsType[*cache.ExecutionCacheWriter](*writer) && !common.IsType[*proxy.LiveCacheWriter](*writer) {
				(*writer).Commit(blockNum)
				return
			}
		})
}
