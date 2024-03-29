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
	"errors"
	"runtime"

	indexer "github.com/arcology-network/common-lib/storage/indexer"
	platform "github.com/arcology-network/storage-committer/platform"
	ccstg "github.com/arcology-network/storage-committer/storage/ccstorage"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
	proxy "github.com/arcology-network/storage-committer/storage/proxy"
	"github.com/arcology-network/storage-committer/univalue"

	"github.com/arcology-network/common-lib/common"
	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
	importer "github.com/arcology-network/storage-committer/importer"
	interfaces "github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/interfaces"
)

// StateCommitter represents a storage committer.
type StateCommitter struct {
	// backends Committable
	store    interfaces.Datastore
	platform *platform.Platform

	byPath *indexer.UnorderedIndexer[string, *univalue.Univalue, []*univalue.Univalue]
	byTxID *indexer.UnorderedIndexer[uint32, *univalue.Univalue, []*univalue.Univalue]

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

// NewStorageCommitter creates a new StateCommitter instance.
func NewStorageCommitter(store interfaces.Datastore) *StateCommitter {
	plat := platform.NewPlatform()

	return &StateCommitter{
		store:    store,
		platform: plat, //[]stgcommcommon.FilteredTransitionsInterface{&importer.NonceFilter{}, &importer.BalanceFilter{}},

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
func (this *StateCommitter) Store() interfaces.Datastore { return this.store }
func (this *StateCommitter) SetStore(store interfaces.Datastore) {
	this.store = store
}

// Import imports the given transitions into the StateCommitter.
func (this *StateCommitter) Import(transitions []*univalue.Univalue, args ...interface{}) *StateCommitter {
	this.byPath.Add(transitions)
	this.byTxID.Add(transitions)
	this.ethIndex.Add(transitions)
	this.ccIndex.Add(transitions)
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
	// Mark the transitions that are not in the whitelist
	this.Whitelist(txs)

	// Finalize all the transitions for both the ETH storage and the concurrent container transitions
	this.byPath.ParallelForeachDo(func(_ string, v *[]*univalue.Univalue) {
		// Remove conflicting ones.
		slice.RemoveIf(v, func(_ int, val *univalue.Univalue) bool { return val.GetPath() == nil })
		if len(*v) > 0 {
			importer.DeltaSequence(*v).Finalize(this.store) // Finalize the transitions and flag the merged ones.
		}
	})

	var ethRootHash [32]byte
	common.ParallelExecute(
		func() {
			this.Store().(*proxy.StoreProxy).CCStore().Precommit(this.ccIndex) // Container store
		},

		func() {
			ethRootHash = this.Store().(*proxy.StoreProxy).EthStore().Precommit(this.ethIndex)
		},
	)
	return ethRootHash // Write to the DB buffer
}

// Commit commits the transitions in the StateCommitter.
func (this *StateCommitter) Commit(blockNum uint64) *StateCommitter {
	var ethStgErr, ccStgErr error
	common.ParallelExecute(
		func() {
			trans := slice.TransformIf(this.byPath.Values(), func(_ int, v []*univalue.Univalue) (bool, *univalue.Univalue) {
				if len(v) > 0 {
					return true, v[0]
				}
				return false, nil
			})
			this.CommitToCache(blockNum, trans)
		},
		func() { ccStgErr = this.Store().(*proxy.StoreProxy).CCStore().Commit(blockNum) },   // To container store
		func() { ethStgErr = this.Store().(*proxy.StoreProxy).EthStore().Commit(blockNum) }, // To ETH store
	)
	this.Err = errors.Join(ethStgErr, ccStgErr)
	this.Clear()
	return this
}

// Update the cache
func (this *StateCommitter) CommitToCache(blockNum uint64, trans []*univalue.Univalue) {
	keys := make([]string, len(trans))
	typedVals := slice.ParallelTransform(trans, runtime.NumCPU(), func(i int, v *univalue.Univalue) intf.Type {
		keys[i] = *v.GetPath()
		if v.Value() != nil {
			return v.Value().(intf.Type)
		}
		return nil // A deletion
	})
	this.Store().(*proxy.StoreProxy).RefreshCache(blockNum, keys, typedVals) // Update the cache
}

// Clear clears the StateCommitter.
func (this *StateCommitter) Clear() {
	this.byPath.Clear()
	this.byTxID.Clear()

	this.ethIndex.Clear()
	this.ccIndex.Clear()
}
