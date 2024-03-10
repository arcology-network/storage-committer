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

// AccountIndexer  avoids having duplicate addresses in the account list and dictionary.

package storage

import (
	"fmt"

	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/storage-committer/importer"
	"github.com/arcology-network/storage-committer/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

// AccountUpdate organizes all the transitions under the same account together.
type AccountUpdate struct {
	Key  string
	Addr ethcommon.Address
	Seqs []*importer.DeltaSequence
	Acct *Account
}

type AccountUpdates []*AccountUpdate

func (this AccountUpdates) KVs() ([]string, []interface{}) {
	univs := slice.ConcateDo(this,
		func(update *AccountUpdate) uint64 {
			return uint64(len(update.Seqs))
		},

		func(update *AccountUpdate) []*univalue.Univalue {
			return slice.Transform(update.Seqs, func(i int, seq *importer.DeltaSequence) *univalue.Univalue {
				return seq.Finalize()
			})
		})

	keys, values := make([]string, len(univs)), make([]interface{}, len(univs))
	for i, univ := range univs {
		keys[i], values[i] = *univ.GetPath(), univ.Value()
	}
	return keys, values
}

func (this AccountUpdate) Print() {
	fmt.Println("Key:", this.Key)
	fmt.Println("Addr:", this.Addr)

	univals := slice.Transform(this.Seqs, func(i int, seq *importer.DeltaSequence) *univalue.Univalue {
		return seq.Finalized
	})
	univalue.Univalues(univals).Print()
}

func (this AccountUpdates) Print() {
	for i := range this {
		this[i].Print()
		fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	}
}
