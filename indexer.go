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
package storagecommitter

import (
	"math"

	indexer "github.com/arcology-network/common-lib/storage/indexer"

	interfaces "github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
)

// An index by path, transitions have the same path will be put together in a list
// This index will be used for apply transitions on the original state. So all the transitions
// should be put into this index.
func PathIndexer(store interfaces.Datastore) *indexer.UnorderedIndexer[string, *univalue.Univalue, []*univalue.Univalue] {
	return indexer.NewUnorderedIndexer(
		nil,

		func(v *univalue.Univalue) (string, bool) {
			return *v.GetPath(), true
		},

		func(k string, v *univalue.Univalue) []*univalue.Univalue {
			if v.Value() != nil {
				v.Value().(intf.Type).Preload(k, store)
			}
			return []*univalue.Univalue{v}
		},

		func(_ string, v *univalue.Univalue, vals *[]*univalue.Univalue) { *vals = append(*vals, v) },
	)
}

// An index by tx number, transitions have the same tx number will be put together in a list.
// This index will be used to remove the transitions generated by the conflicting transactions.
// So, the immutable transitions should not be put into this index.
func TxIndexer(store interfaces.Datastore) *indexer.UnorderedIndexer[uint32, *univalue.Univalue, []*univalue.Univalue] {
	return indexer.NewUnorderedIndexer(
		nil,

		func(v *univalue.Univalue) (uint32, bool) {
			if !v.Persistent() {
				return v.GetTx(), true
			}
			return math.MaxUint32, false
		},
		func(_ uint32, v *univalue.Univalue) []*univalue.Univalue { return []*univalue.Univalue{v} },
		func(_ uint32, v *univalue.Univalue, vals *[]*univalue.Univalue) { *vals = append(*vals, v) },
	)
}

// // An index by account address, transitions have the same Eth account address will be put together in a list
// // This is for ETH storage, concurrent container related sub-paths won't be put into this index.
// func EthIndexer(store interfaces.Datastore) *indexer.UnorderedIndexer[[20]byte, *univalue.Univalue, *associative.Pair[*ethstg.Account, []*univalue.Univalue]] {
// 	return indexer.NewUnorderedIndexer(
// 		nil,
// 		func(v *univalue.Univalue) ([20]byte, bool) {
// 			if !platform.IsEthPath(*v.GetPath()) {
// 				return [20]byte{}, false
// 			}
// 			addr, _ := hexutil.Decode(platform.GetAccountAddr(*v.GetPath()))
// 			return ethcommon.BytesToAddress(addr), platform.IsEthPath(*v.GetPath())
// 		},

// 		func(addr [20]byte, v *univalue.Univalue) *associative.Pair[*ethstg.Account, []*univalue.Univalue] {
// 			return &associative.Pair[*ethstg.Account, []*univalue.Univalue]{
// 				First:  store.Preload(addr[:]).(*ethstg.Account),
// 				Second: []*univalue.Univalue{v},
// 			}
// 		},

// 		func(_ [20]byte, v *univalue.Univalue, pair **associative.Pair[*ethstg.Account, []*univalue.Univalue]) {
// 			(**pair).Second = append((**pair).Second, v)
// 		},
// 	)
// }
