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
	stgcommon "github.com/arcology-network/storage-committer/common"
	platform "github.com/arcology-network/storage-committer/platform"
	cache "github.com/arcology-network/storage-committer/storage/cache"
	"github.com/arcology-network/storage-committer/type/univalue"

	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/storage-committer/storage/ethstorage"
)

/*
	StateCommitter is a core component responsible for managing and committing state transitions
	within the Arcology Network's storage system. It handles the import, indexing, finalization,
	and commitment of state changes (transitions) to various underlying storage backends.
	The committer supports features such as wildcard substitution, transaction whitelisting,
	and efficient batch operations, ensuring consistency and performance across different storage layers.
	This implementation leverages indexers and asynchronous writers to optimize the processing and persistence
	of state updates.
*/

type StateCommitter struct {
	readonlyStore stgcommon.ReadOnlyStore
	platform      *platform.Platform

	writers []stgcommon.Writer[*univalue.Univalue] // db writers

	syncWriters  []stgcommon.Writer[*univalue.Univalue] // db writers that need to be synchronized
	asyncWriters []stgcommon.Writer[*univalue.Univalue] // db writer that is used for asynchronous commit

	byPath *indexer.UnorderedIndexer[string, *univalue.Univalue, []*univalue.Univalue]
	byTxID *indexer.UnorderedIndexer[uint64, *univalue.Univalue, []*univalue.Univalue]

	Err error
}

// NewStateCommitter creates a new StateCommitter instance. The stores are the stores that can be committed.
// A Committable store is a pair of an index and a store. The index is used to index the input transitions as they are
// received, and the store is used to commit the indexed transitions. Since multiple store can share the same index, each
// CommittableStore is an indexer and a list of Committable stores.
func NewStateCommitter(readonlyStore stgcommon.ReadOnlyStore, writers []stgcommon.Writer[*univalue.Univalue]) *StateCommitter {
	committer := &StateCommitter{
		readonlyStore: readonlyStore,
		platform:      platform.NewPlatform(),

		writers: writers,
		byPath:  PathIndexer(readonlyStore), // By storage path
		byTxID:  TxIndexer(readonlyStore),   // By tx ID, used to quickly remove the transitions that are not in the whitelist.
	}

	// Filter the writers into synchronous first.
	committer.syncWriters = slice.MoveIf(&writers, func(_ int, v stgcommon.Writer[*univalue.Univalue]) bool {
		return v.IsSync()
	})

	// Whatever is left is asynchronous writers.
	committer.asyncWriters = writers
	return committer
}

// New creates a new StateCommitter instance.
func (this *StateCommitter) New(args ...interface{}) *StateCommitter {
	return &StateCommitter{
		platform: platform.NewPlatform(),
	}
}

// Importer returns the importer of the StateCommitter.
func (this *StateCommitter) Store() stgcommon.ReadOnlyStore { return this.readonlyStore }
func (this *StateCommitter) SetStore(store stgcommon.ReadOnlyStore) { // Testing only
	this.readonlyStore = store
}

// Import imports the given transitions into the StateCommitter.
func (this *StateCommitter) Import(transitions []*univalue.Univalue) *StateCommitter {
	transitions = this.SubstitueWildcards(transitions) // Import the wildcards, if any.

	// Import the regular transitions to the indexers.
	this.byPath.Import(transitions)
	this.byTxID.Import(transitions)

	for _, writer := range this.writers {
		writer.Import(transitions)
	}
	return this
}

// Finalize finalizes the transitions in the StateCommitter.
func (this *StateCommitter) whitelist(txs []uint64) *StateCommitter {
	if len(txs) == 0 {
		return this
	}

	whitelistDict := mapi.FromSlice(txs, func(_ uint64) bool { return true })
	this.byTxID.ParallelForeachDo(func(txid uint64, vec *[]*univalue.Univalue) {
		if _, ok := whitelistDict[uint64(txid)]; !ok {
			for _, v := range *vec {
				v.SetPath(nil) // Mark the transition status, so that it can be removed later.
			}
		}
	})
	return this
}

// Commit commits the transitions to stores.
func (this *StateCommitter) Finalize(txs []uint64) {
	this.whitelist(txs) // Mark the transitions that are not in the whitelist

	// Finalize all the transitions by merging the transitions
	// for both the ETH storage and the concurrent container transitions
	this.byPath.ParallelForeachDo(func(_ string, v *[]*univalue.Univalue) {
		slice.RemoveIf(v, func(_ int, val *univalue.Univalue) bool { return val.GetPath() == nil }) // Remove conflicting ones.
		if len(*v) > 0 {
			// Finalize the transitions and flag the merged ones.
			DeltaSequence(*v).Finalize(this.readonlyStore) // Sort and finalize the transitions.
		}
	})

	// Import the affiliated deletes, no need to finalize them because they are already finalized.
	// this.Import(GetCascadeSub)

	this.byPath.Clear()
	this.byTxID.Clear()
}

func (this *StateCommitter) CascadeDelete(txs []uint64) {
	// When deleting a path, all the sub paths are also deleted. But deleted sub paths aren't part of the transitions.
	// to save bandwidth, we need generate these transitions here.
	// GetCascadeSub := []*univalue.Univalue{}
	// this.byPath.ForeachDo(func(_ string, v []*univalue.Univalue) {
	// 	if v[0].TypeID() == commutative.PATH && v[0].Value() == nil {
	// 		for _, k := range v[0].Value().(*softdeltaset.DeltaSet[string]).Elements() {
	// 			key := *(v[0].Property.GetPath()) + k // Concatenate the path and the subkey
	// 			v := univalue.NewUnivalue(v[0].GetTx(), key, 0, 1, 0, nil, nil)
	// 			GetCascadeSub = append(GetCascadeSub, v)
	// 		}
	// 	}
	// })
}

// Commit commits the transitions in the StateCommitter.
// 1. For the block write cache, it commits the transitions to the cache.
// 2. For the eth storage, it updates the tries without committing the transitions to the DB
func (this *StateCommitter) Precommit(txs []uint64) *StateCommitter {
	this.Finalize(txs)
	this.SyncPrecommit()
	this.AsyncPrecommit()
	return this
}

// Only the global write cache needs to be synchronized before the next precommit or commit.
func (this *StateCommitter) SyncPrecommit() {
	slice.ParallelForeach(this.writers, len(this.writers),
		func(i int, writer *stgcommon.Writer[*univalue.Univalue]) {
			// Eth storage only serves user API enquiries. It has nothing to do with the
			// transitions execution. So we do not need to precommit it synchronously.
			if !common.IsType[*ethstorage.EthStorageWriter](*writer) {
				(*writer).Precommit(true)
			}
		})
}

// Only the global write cache needs to be synchronized before the next precommit or commit.
func (this *StateCommitter) AsyncPrecommit() {
	slice.ParallelForeach(this.writers, len(this.writers),
		func(_ int, writer *stgcommon.Writer[*univalue.Univalue]) {
			if !common.IsType[*cache.ExecutionCacheWriter](*writer) {
				(*writer).Precommit(false)
			}
		})
}

// Commit commits the transitions to different stores.
// This is mainly for degugging purposes. The two functions should
// be called separately, one for synchronous commit and one for asynchronous commit.
func (this *StateCommitter) Commit(blockNum uint64) *StateCommitter {
	this.SyncCommit(blockNum)
	this.AsyncCommit(blockNum)
	return this
}

// Only the global write cache needs to be synchronized before the next precommit.
func (this *StateCommitter) SyncCommit(blockNum uint64) {
	slice.ParallelForeach(this.syncWriters, len(this.syncWriters),
		func(_ int, writer *stgcommon.Writer[*univalue.Univalue]) {
			(*writer).Commit(blockNum)
		})
}

// Only the global write cache needs to be synchronized before the next precommit.
func (this *StateCommitter) AsyncCommit(blockNum uint64) {
	slice.ParallelForeach(this.asyncWriters, len(this.asyncWriters),
		func(_ int, writer *stgcommon.Writer[*univalue.Univalue]) {
			(*writer).Commit(blockNum)
		})
}
