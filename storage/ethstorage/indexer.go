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

package ethstorage

import (
	"github.com/arcology-network/common-lib/exp/associative"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/common-lib/storage/indexer"
	platform "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
type EthIndexer struct {
	*indexer.UnorderedIndexer[[20]byte, *univalue.Univalue, *associative.Pair[*Account, []*univalue.Univalue]]

	version       uint64 // Block number of the last update
	dirtyAccounts []*Account
	// dirtyVals     [][]interfaces.Type // Dirty accountCache are the accountCache that have been updated in the current cycle.
	// dirtyKeys     [][]string          // Dirty accountCache are the accountCache that have been updated in the current cycle.
	err error
}

func NewEthIndexer(store *EthDataStore, version uint64) *EthIndexer {
	idxer := (indexer.NewUnorderedIndexer(
		nil,
		func(v *univalue.Univalue) ([20]byte, bool) {
			// if !platform.IsEthPath(*v.GetPath()) {
			// 	return [20]byte{}, false
			// }
			addr, _ := hexutil.Decode(platform.GetAccountAddr(*v.GetPath()))
			return ethcommon.BytesToAddress(addr), true //platform.IsEthPath(*v.GetPath())
		},

		func(addr [20]byte, v *univalue.Univalue) *associative.Pair[*Account, []*univalue.Univalue] {
			return &associative.Pair[*Account, []*univalue.Univalue]{
				First:  store.Preload(addr[:]).(*Account),
				Second: []*univalue.Univalue{v},
			}
		},

		func(_ [20]byte, v *univalue.Univalue, pair **associative.Pair[*Account, []*univalue.Univalue]) {
			(**pair).Second = append((**pair).Second, v)
		},
	))

	return &EthIndexer{
		version:          version,
		UnorderedIndexer: idxer,
	}
}

func (this *EthIndexer) Add(v []*univalue.Univalue) {
	this.UnorderedIndexer.Add(v)
}

// Remove the nil transitions from the index, because they are set by
func (this *EthIndexer) Finalize() {
	this.ParallelForeachDo(func(_ [20]byte, v **associative.Pair[*Account, []*univalue.Univalue]) {
		slice.RemoveIf(&((**v).Second), func(_ int, v *univalue.Univalue) bool { return v.GetPath() == nil })
	})

	// Remove accounts that have no transitions left after cleanning up
	pairs := this.UnorderedIndexer.Values()
	slice.RemoveIf(&(pairs), func(_ int, v *associative.Pair[*Account, []*univalue.Univalue]) bool { return len(v.Second) == 0 })
}

// Merge indexers so they can be updated at once.
func (this *EthIndexer) Merge(idxers []*EthIndexer) *EthIndexer {
	slice.Remove(&idxers, nil) // Remove the nil elements

	this.dirtyAccounts = slice.ConcateDo(idxers,
		func(idxer *EthIndexer) uint64 { return uint64(len(idxer.dirtyAccounts)) },
		func(idxer *EthIndexer) []*Account { return idxer.dirtyAccounts })

	// this.dirtyVals = slice.ConcateDo(idxers,
	// 	func(idxer *EthIndexer) uint64 { return uint64(len(idxer.dirtyVals)) },
	// 	func(idxer *EthIndexer) [][]interfaces.Type { return idxer.dirtyVals })

	// this.dirtyKeys = slice.ConcateDo(idxers,
	// 	func(idxer *EthIndexer) uint64 { return uint64(len(idxer.dirtyKeys)) },
	// 	func(idxer *EthIndexer) [][]string { return idxer.dirtyKeys })

	return this
}
