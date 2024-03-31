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
package storagecommitter

import (
	indexer "github.com/arcology-network/common-lib/storage/indexer"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	ccstg "github.com/arcology-network/storage-committer/storage/ccstorage"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	"github.com/arcology-network/storage-committer/univalue"

	"github.com/arcology-network/common-lib/exp/associative"
	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
	importer "github.com/arcology-network/storage-committer/importer"
)

// StateCommitter represents a storage committer.
// The main purpose of the StateCommitter is to commit the transitions to the different stores.
type StateCommitter struct {
	store          intf.Datastore
	platform       *platform.Platform
	WritableStores []*associative.Pair[intf.Indexer[*univalue.Univalue], []intf.WritableStore] // backends Committable
	byPath         *indexer.UnorderedIndexer[string, *univalue.Univalue, []*univalue.Univalue]
	byTxID         *indexer.UnorderedIndexer[uint32, *univalue.Univalue, []*univalue.Univalue]

	ethIndex interface {
		Add([]*univalue.Univalue)
		Clear()
	}
	ccIndex interface {
		Add([]*univalue.Univalue)
		Clear()
	}
	Err error
}

// NewStorageCommitter creates a new StateCommitter instance. The WritableStores are the stores that can be committed.
// A Committable store is a pair of an index and a store. The index is used to index the input transitions as they are
// received, and the store is used to commit the indexed transitions. Since multiple store can share the same index, each
// WritableStore is an indexer and a list of Committable stores.
func NewStorageCommitter(store intf.Datastore) *StateCommitter {
	return &StateCommitter{
		store:          store,
		platform:       platform.NewPlatform(),
		WritableStores: store.(*stgproxy.StorageProxy).Committable(),

		byPath: PathIndexer(store), // By storage path
		byTxID: TxIndexer(store),   // By tx ID

		ethIndex: ethstg.NewIndexer(store), // By eth account
		ccIndex:  ccstg.NewIndexer(store),  // By concurrent container
	}
}

// New creates a new StateCommitter instance.
func (this *StateCommitter) New(args ...interface{}) *StateCommitter {
	return &StateCommitter{
		platform: platform.NewPlatform(),
	}
}

// Importer returns the importer of the StateCommitter.
func (this *StateCommitter) Store() intf.Datastore { return this.store }
func (this *StateCommitter) SetStore(store intf.Datastore) {
	this.store = store
}

// Import imports the given transitions into the StateCommitter.
func (this *StateCommitter) Import(transitions []*univalue.Univalue, args ...interface{}) *StateCommitter {
	this.byPath.Add(transitions)
	this.byTxID.Add(transitions)

	for _, pair := range this.WritableStores {
		pair.First.Add(transitions)
	}
	return this
}

// Finalize finalizes the transitions in the StateCommitter.
func (this *StateCommitter) Whitelist(txs []uint32) *StateCommitter {
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

// Commit commits the transitions in the StateCommitter.
func (this *StateCommitter) Precommit(txs []uint32) [32]byte {
	this.Whitelist(txs) // Mark the transitions that are not in the whitelist

	// Finalize all the transitions for both the ETH storage and the concurrent container transitions
	this.byPath.ParallelForeachDo(func(_ string, v *[]*univalue.Univalue) {
		// Remove conflicting ones.
		slice.RemoveIf(v, func(_ int, val *univalue.Univalue) bool { return val.GetPath() == nil })
		if len(*v) > 0 {
			importer.DeltaSequence(*v).Finalize(this.store) // Finalize the transitions and flag the merged ones.
		}
	})

	// Commit the transitions to different stores
	for _, pair := range this.WritableStores {
		for _, store := range pair.Second {
			pair.First.Finalize()       // Remove the excluded transitions
			store.Precommit(pair.First) // Commit the transitions
		}
	}
	return [32]byte{} // Write to the DB buffer
}

func (this *StateCommitter) Commit(blockNum uint64) *StateCommitter {
	for _, pair := range this.WritableStores {
		for _, store := range pair.Second {
			store.Commit(blockNum)
		}
	}
	return this
}

// Clear clears the StateCommitter.
func (this *StateCommitter) Clear() {
	this.byPath.Clear()
	this.byTxID.Clear()

	for _, pair := range this.WritableStores {
		pair.First.Clear()
	}
}
