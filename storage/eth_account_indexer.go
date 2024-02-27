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

	"github.com/arcology-network/common-lib/common"
	indexed "github.com/arcology-network/common-lib/container/array"
	"github.com/arcology-network/common-lib/exp/array"
	"github.com/arcology-network/storage-committer/importer"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
)

// indexer  avoids having duplicate addresses in the account list and dictionary.
type AccountIndexer struct {
	platform *platform.Platform
	store    intf.Datastore
	dict     *indexed.IndexedArray[*importer.DeltaSequence, string, *AccountUpdate]
}

// Newindexer creates a new indexer instance.
func NewAccountIndexer(
	store intf.Datastore,
	platform *platform.Platform,
	keygetter func(*importer.DeltaSequence) string,
	inserter func(*importer.DeltaSequence, *AccountUpdate) *AccountUpdate) *AccountIndexer {
	return &AccountIndexer{
		platform: platform,
		store:    store,
		dict:     indexed.NewIndexedArray[*importer.DeltaSequence, string, *AccountUpdate](keygetter, inserter, nil),
	}
}

// Add the transaction to the account dictionary.
func (this *AccountIndexer) Add(transitions []*importer.DeltaSequence) {
	if !common.IsType[*EthDataStore](this.store) {
		return
	}

	for _, tran := range transitions {
		this.dict.Insert(tran)
	}
}

func (this *AccountIndexer) Updates() []*AccountUpdate {
	acctUpdates := this.dict.Elements()
	array.ParallelForeach(acctUpdates, runtime.NumCPU(), func(i int, update **AccountUpdate) {
		array.RemoveIf[*importer.DeltaSequence](&(*update).Seqs, func(i int, seq *importer.DeltaSequence) bool {
			return seq.Finalized() == nil
		})
	})
	return acctUpdates
}

func (this *AccountIndexer) Clear() {
	this.dict.Clear()
}
