/*
 *   Copyright (c) 2023 Arcology Network

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
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/orderedmap"
	"github.com/arcology-network/common-lib/exp/slice"
	platform "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/storage"
	"github.com/arcology-network/storage-committer/univalue"

	importer "github.com/arcology-network/storage-committer/importer"
	interfaces "github.com/arcology-network/storage-committer/interfaces"
)

// StateCommitter represents a storage committer.
type StateCommitter struct {
	store    interfaces.Datastore
	Platform *platform.Platform
	byPath   *orderedmap.OrderedMap[string, *univalue.Univalue, *importer.DeltaSequence]

	acctIndex   *storage.AccountIndexer // Account acctIndex is an index by unique account address.
	importer    *importer.Importer
	imuImporter *importer.Importer // transitions that will take effect anyway regardless of execution failures or conflicts
}

// NewStorageCommitter creates a new StateCommitter instance.
func NewStorageCommitter(store interfaces.Datastore) *StateCommitter {
	plat := platform.NewPlatform()
	return &StateCommitter{
		store:    store,
		Platform: plat, //[]stgcommcommon.FilteredTransitionsInterface{&importer.NonceFilter{}, &importer.BalanceFilter{}},

		// An index by path, transitions have the same path will be put together in a list
		// This index will be used for apply transitions on the original state. So all the transitions
		// should be put into this index.
		byPath: orderedmap.NewOrderedMap[string, *univalue.Univalue, *importer.DeltaSequence](
			nil,
			1024,
			func(k string, v *univalue.Univalue) *importer.DeltaSequence {
				return importer.NewDeltaSequence(k, store)
			},
			func(k string, v *univalue.Univalue, seq **importer.DeltaSequence) {
				(**seq).Add(v)
			}),

		importer:    importer.NewImporter(store, plat),
		imuImporter: importer.NewImporter(store, plat),
		acctIndex:   storage.NewAccountIndexer(store, plat),
	}
}

// New creates a new StateCommitter instance.
func (this *StateCommitter) New(args ...interface{}) *StateCommitter {
	return &StateCommitter{
		Platform: platform.NewPlatform(),
	}
}

// Importer returns the importer of the StateCommitter.
func (this *StateCommitter) Store() interfaces.Datastore { return this.store }

// Init initializes the StateCommitter with the given datastore.
func (this *StateCommitter) Init(store interfaces.Datastore) {
	this.importer.Init(store)
	this.imuImporter.Init(store)
}

// Clear clears the StateCommitter.
func (this *StateCommitter) Clear() {
	this.importer.Store().Clear()
	this.importer.Clear()
	this.imuImporter.Clear()
	this.acctIndex.Clear() // Clear the account acctIndex
}

// Import imports the given transitions into the StateCommitter.
func (this *StateCommitter) Import(transitions []*univalue.Univalue, args ...interface{}) *StateCommitter {
	this.byPath.InsertDo(transitions, func(_ int, v *univalue.Univalue) string {
		return *(*v).GetPath()
	})

	// Move the Peristent transitions(nonce and gas fee) to another list.
	invTransitions := slice.MoveIf(&transitions, func(i int, v *univalue.Univalue) bool {
		return v.Persistent()
	})
	slice.Remove(&transitions, nil) // Remove the Peristent transitions from the transition lists

	var imuSeqs, seqs []*importer.DeltaSequence
	common.ParallelExecute(
		func() { imuSeqs = this.imuImporter.Import(invTransitions, args...) },
		func() { seqs = this.importer.Import(transitions, args...) })

	// Add to the acctIndex for the account index
	this.acctIndex.Add(seqs, imuSeqs)
	return this
}

// Finalize finalizes the transitions in the StateCommitter.
func (this *StateCommitter) Finalize(txs []uint32) *StateCommitter {
	this.byPath.ParallelForeachDo(func(_ string, v *importer.DeltaSequence) {
		v.Finalize()
	})

	// Sort delta transitions
	common.ParallelExecute(
		func() { this.imuImporter.SortDeltaSequences() },
		func() { this.importer.SortDeltaSequences() })

	// Commit all the transactions
	if txs != nil && len(txs) == 0 {
		return this
	}

	common.ParallelExecute(
		func() {
			this.imuImporter.MergeStateDelta()
		},
		func() {
			this.importer.WhiteList(txs)    // Remove all the transitions generated by the conflicting transactions
			this.importer.MergeStateDelta() // Finalize states
		},
	)
	return this
}

// Commit commits the transitions in the StateCommitter.
func (this *StateCommitter) Precommit(txs []uint32) [32]byte {
	if txs != nil && len(txs) == 0 {
		this.Clear()
		// panic("No transactions to commit")
		return [32]byte{}
	}
	this.Finalize(txs)
	return this.store.Precommit(this.acctIndex.Updates()) // Write to the DB buffer
}

// Commit commits the transitions in the StateCommitter.
func (this *StateCommitter) Commit() *StateCommitter {
	store := this.importer.Store()
	store.Commit(0) // Commit to the state store
	this.Clear()
	return this
}
