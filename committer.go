// Package storagecommitter provides functionality for committing storage changes to url2a datastore.
package storagecommitter

import (
	"github.com/arcology-network/common-lib/common"
	committercommon "github.com/arcology-network/concurrenturl/common"

	importer "github.com/arcology-network/concurrenturl/importer"
	interfaces "github.com/arcology-network/concurrenturl/interfaces"
)

// StorageCommitter represents a storage committer.
type StorageCommitter struct {
	importer    *importer.Importer
	imuImporter *importer.Importer // transitions that will take effect anyway regardless of execution failures or conflicts
	Platform    *committercommon.Platform
}

// NewStorageCommitter creates a new StorageCommitter instance.
func NewStorageCommitter(store interfaces.Datastore) *StorageCommitter {
	platform := committercommon.NewPlatform()
	return &StorageCommitter{
		importer:    importer.NewImporter(store, platform),
		imuImporter: importer.NewImporter(store, platform),
		Platform:    platform, //[]committercommon.FilteredTransitionsInterface{&importer.NonceFilter{}, &importer.BalanceFilter{}},
	}
}

// New creates a new StorageCommitter instance.
func (this *StorageCommitter) New(args ...interface{}) *StorageCommitter {
	return &StorageCommitter{
		Platform: committercommon.NewPlatform(),
	}
}

// Importer returns the importer of the StorageCommitter.
func (this *StorageCommitter) Importer() *importer.Importer { return this.importer }

// Init initializes the StorageCommitter with the given datastore.
func (this *StorageCommitter) Init(store interfaces.Datastore) {
	this.importer.Init(store)
	this.imuImporter.Init(store)
}

// Clear clears the StorageCommitter.
func (this *StorageCommitter) Clear() {
	this.importer.Store().Clear()
	this.importer.Clear()
	this.imuImporter.Clear()
}

// Import imports the given transitions into the StorageCommitter.
func (this *StorageCommitter) Import(transitions []interfaces.Univalue, args ...interface{}) *StorageCommitter {
	invTransitions := make([]interfaces.Univalue, 0, len(transitions))

	for i := 0; i < len(transitions); i++ {
		if transitions[i].Persistent() { // Peristent transitions are immune to conflict detection
			invTransitions = append(invTransitions, transitions[i]) //
			transitions[i] = nil                                    // mark the peristent transitions
		}
	}
	common.Remove(&transitions, nil) // Remove the Peristent transitions from the transition lists

	common.ParallelExecute(
		func() { this.imuImporter.Import(invTransitions, args...) },
		func() { this.importer.Import(transitions, args...) })
	return this
}

// Sort sorts the transitions in the StorageCommitter.
func (this *StorageCommitter) Sort() *StorageCommitter {
	common.ParallelExecute(
		func() { this.imuImporter.SortDeltaSequences() },
		func() { this.importer.SortDeltaSequences() })

	return this
}

// Finalize finalizes the transitions in the StorageCommitter.
func (this *StorageCommitter) Finalize(txs []uint32) *StorageCommitter {
	if txs != nil && len(txs) == 0 { // Commit all the transactions when txs == nil
		return this
	}

	common.ParallelExecute(
		func() { this.imuImporter.MergeStateDelta() },
		func() {
			this.importer.WhilteList(txs)   // Remove all the transitions generated by the conflicting transactions
			this.importer.MergeStateDelta() // Finalize states
		},
	)
	return this
}

// CopyToDbBuffer copies the transitions to the DB buffer.
func (this *StorageCommitter) CopyToDbBuffer() [32]byte {
	keys, values := this.importer.KVs()
	invKeys, invVals := this.imuImporter.KVs()

	keys, values = append(keys, invKeys...), append(values, invVals...)
	return this.importer.Store().Precommit(keys, values) // save the transitions to the DB buffer
}

// SaveToDB saves the transitions to the database.
func (this *StorageCommitter) SaveToDB() {
	store := this.importer.Store()
	store.Commit(0) // Commit to the state store
	this.Clear()
}

// Commit commits the transitions in the StorageCommitter.
func (this *StorageCommitter) Commit(txs []uint32) *StorageCommitter {
	if txs != nil && len(txs) == 0 {
		this.Clear()
		return this
	}
	this.Finalize(txs)
	this.CopyToDbBuffer() // Export transitions and save them to the DB buffer.
	this.SaveToDB()
	return this
}