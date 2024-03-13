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

package storage

import (
	"runtime"

	indexed "github.com/arcology-network/common-lib/container/indexed"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/storage-committer/importer"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	ethcommon "github.com/ethereum/go-ethereum/common"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
)

// AccountIndexer is an index to put all the transitions under the same account together.
type AccountIndexer struct {
	platform *platform.Platform
	store    intf.Datastore
	dict     *indexed.IndexedSlice[*importer.DeltaSequence, string, *AccountUpdate] // A hybrid combining a slice and a map.
}

// Newindexer creates a new indexer instance.
func NewAccountIndexer(
	store intf.Datastore,
	platform *platform.Platform) *AccountIndexer {

	gkeyGetter := func(seq *importer.DeltaSequence) string {
		return seq.Account
	}

	ginserter := func(seq *importer.DeltaSequence, update *AccountUpdate) *AccountUpdate {
		if update == nil {
			addr, _ := hexutil.Decode(seq.Account)
			return &AccountUpdate{
				Key:  seq.Account,
				Addr: ethcommon.BytesToAddress(addr),
				Seqs: []*importer.DeltaSequence{seq},
				Acct: store.Preload(addr).(*Account),
			}
		}
		update.Seqs = append(update.Seqs, seq)
		return update
	}

	return &AccountIndexer{
		platform: platform,
		store:    store,
		dict:     indexed.NewIndexedSlice[*importer.DeltaSequence, string, *AccountUpdate](gkeyGetter, ginserter, nil),
	}
}

// Add the transaction to the account dictionary.
func (this *AccountIndexer) Add(transitionSets ...[]*importer.DeltaSequence) {
	for _, trans := range transitionSets {
		for _, tran := range trans {
			this.dict.Insert(tran)
		}
	}
}

func (this *AccountIndexer) Updates() []*AccountUpdate {
	acctUpdates := this.dict.Elements()
	slice.ParallelForeach(acctUpdates, runtime.NumCPU(), func(i int, update **AccountUpdate) {
		slice.RemoveIf[*importer.DeltaSequence](&(*update).Seqs, func(i int, seq *importer.DeltaSequence) bool {
			return seq.Finalized == nil
		})
	})
	return acctUpdates
}

func (this *AccountIndexer) Clear() {
	this.dict.Clear()
}
