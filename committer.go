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
	"fmt"
	"time"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	platform "github.com/arcology-network/concurrenturl/platform"
	"github.com/arcology-network/concurrenturl/storage"
	"github.com/arcology-network/concurrenturl/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	importer "github.com/arcology-network/concurrenturl/importer"
	interfaces "github.com/arcology-network/concurrenturl/interfaces"
)

// StateCommitter represents a storage committer.
type StateCommitter struct {
	store       interfaces.Datastore
	indexer     *storage.AccountIndexer // Account indexer is an index by unique account address.
	importer    *importer.Importer
	imuImporter *importer.Importer // transitions that will take effect anyway regardless of execution failures or conflicts
	Platform    *platform.Platform
}

// NewStorageCommitter creates a new StateCommitter instance.
func NewStorageCommitter(store interfaces.Datastore) *StateCommitter {
	keyGetter := func(seq *importer.DeltaSequence) string {
		return seq.Key
	}

	inserter := func(seq *importer.DeltaSequence, update *storage.AccountUpdate) *storage.AccountUpdate {
		if update == nil {
			addr, _ := hexutil.Decode(seq.Key)
			return &storage.AccountUpdate{
				Key:  seq.Key,
				Addr: ethcommon.BytesToAddress(addr),
				Seqs: []*importer.DeltaSequence{seq},
				Acct: store.(*storage.EthDataStore).PreloadAccount(addr),
			}
		}
		update.Seqs = append(update.Seqs, seq)
		return update
	}

	platform := platform.NewPlatform()
	return &StateCommitter{
		store:       store,
		importer:    importer.NewImporter(store, platform),
		imuImporter: importer.NewImporter(store, platform),
		Platform:    platform, //[]committercommon.FilteredTransitionsInterface{&importer.NonceFilter{}, &importer.BalanceFilter{}},
		indexer:     storage.NewAccountIndexer(store, platform, keyGetter, inserter),
	}
}

// New creates a new StateCommitter instance.
func (this *StateCommitter) New(args ...interface{}) *StateCommitter {
	return &StateCommitter{
		Platform: platform.NewPlatform(),
	}
}

// Importer returns the importer of the StateCommitter.
func (this *StateCommitter) Importer() *importer.Importer { return this.importer }

// Init initializes the StateCommitter with the given datastore.
func (this *StateCommitter) Init(store interfaces.Datastore) {
	this.importer.Init(store)
	this.imuImporter.Init(store)
}

// Clear clears the StateCommitter.
func (this *StateCommitter) Clear() {
	t0 := time.Now()
	this.importer.Store().Clear()
	this.importer.Clear()
	this.imuImporter.Clear()
	this.indexer.Clear() // Clear the account indexer
	fmt.Println("StateCommitter.Clear(): ", time.Since(t0))
}

// Import imports the given transitions into the StateCommitter.
func (this *StateCommitter) Import(transitions []*univalue.Univalue, args ...interface{}) *StateCommitter {
	// Move the Peristent transitions(nonce and gas fee) to another list.
	invTransitions := array.MoveIf(&transitions, func(i int, v *univalue.Univalue) bool {
		return v.Persistent()
	})
	array.Remove(&transitions, nil) // Remove the Peristent transitions from the transition lists

	var imuSeqs, seqs []*importer.DeltaSequence
	common.ParallelExecute(
		func() { imuSeqs = this.imuImporter.Import(invTransitions, args...) },
		func() { seqs = this.importer.Import(transitions, args...) })

	// Add to the indexer for the account index
	this.indexer.Add(append(seqs, imuSeqs...))
	return this
}

// Sort sorts the transitions in the StateCommitter.
func (this *StateCommitter) Sort() *StateCommitter {
	common.ParallelExecute(
		func() { this.imuImporter.SortDeltaSequences() },
		func() { this.importer.SortDeltaSequences() })
	return this
}

// Finalize finalizes the transitions in the StateCommitter.
func (this *StateCommitter) Finalize(txs []uint32) *StateCommitter {
	if txs != nil && len(txs) == 0 { // Commit all the transactions when txs == nil
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
	return this.importer.Store().Precommit(this.indexer.Updates()) // Write to the DB buffer
}

// Commit commits the transitions in the StateCommitter.
func (this *StateCommitter) Commit() *StateCommitter {
	store := this.importer.Store()
	store.Commit(0) // Commit to the state store
	this.Clear()
	return this
}
