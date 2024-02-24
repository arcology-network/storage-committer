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

package importer

import (
	"github.com/arcology-network/common-lib/common"
	indexed "github.com/arcology-network/common-lib/container/array"
	intf "github.com/arcology-network/concurrenturl/interfaces"
	platform "github.com/arcology-network/concurrenturl/platform"
	"github.com/arcology-network/concurrenturl/storage"
)

// indexer  avoids having duplicate addresses in the account list and dictionary.
type AccountIndexer struct {
	platform *platform.Platform
	store    intf.Datastore
	dict     *indexed.IndexedArray[*DeltaSequence, string, *AccountUpdate]
}

// Newindexer creates a new indexer instance.
func NewAccountIndexer(
	store intf.Datastore,
	platform *platform.Platform,
	keygetter func(*DeltaSequence) string,
	inserter func(*DeltaSequence, *AccountUpdate) *AccountUpdate) *AccountIndexer {
	return &AccountIndexer{
		platform: platform,
		store:    store,
		dict:     indexed.NewIndexedArray[*DeltaSequence, string, *AccountUpdate](keygetter, inserter, nil),
	}
}

// Add the transaction to the account dictionary.
func (this *AccountIndexer) Add(transitions []*DeltaSequence) {
	if !common.IsType[*storage.EthDataStore](this.store) {
		return
	}

	for _, tran := range transitions {
		this.dict.Insert(tran)
	}

	// array.Foreach(transitions, func(i int, v **univalue.Univalue) {
	// 	addrString := platform.GetAccountAddr(*(*v).GetPath())
	// 	acctTriplet, ok := this.Index[addrString]
	// 	if !ok { // Does not exist
	// 		addr, _ := hexutil.Decode(platform.GetAccountAddr(addrString))
	// 		acctTriplet = &AccountUpdate{ // Create a new triplet
	// 			addrString,
	// 			this.store.(*storage.EthDataStore).NewAccount(addr), // Create an Eth account using the default configuration.
	// 			[]*DeltaSequence{*v}}
	// 		this.Index[addrString] = acctTriplet // Add the triplet to the dictionary
	// 		return
	// 	}
	// 	acctTriplet.Third = append(acctTriplet.Third, *v) // Add the transition to the triplet
	// })
}

// Remove the transitions that are marked for removal by the WhiteList function.
// Remove the account if it has no transitions. The results will be used for updating the trie.
func (this *AccountIndexer) Organize() {
	// array.ParallelForeach(this.Accounts, runtime.NumCPU(), func(i int, triplet **AccountUpdate) {
	// 	array.RemoveIf[*DeltaSequence](&((*triplet).Third), func(i int, seq *DeltaSequence) bool {
	// 		return seq.Finalized() == nil
	// 	})
	// })

	// // Remove the accounts that have no transitions left after white-listing.
	// array.RemoveIf(&this.Accounts, func(_ int, triple *AccountUpdate) bool {
	// 	return len(triple.Third) == 0
	// })
}
