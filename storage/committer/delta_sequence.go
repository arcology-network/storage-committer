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

//
// delta_sequence.go provides types and methods for managing sequences of state deltas
// in the Arcology Network's storage committer. It defines DeltaSequence and DeltaSequences,
// which represent ordered collections of state transitions (Univalue objects), and provides
// utilities for sorting, finalizing, and extracting values and keys from these sequences.
//

package statestore

import (
	"sort"

	"github.com/arcology-network/common-lib/exp/slice"
	stgcommon "github.com/arcology-network/storage-committer/common"
	univalue "github.com/arcology-network/storage-committer/type/univalue"
)

type DeltaSequence []*univalue.Univalue

func (this DeltaSequence) sort() DeltaSequence {
	if len(this) <= 1 {
		return this
	}

	sort.SliceStable(this, func(i, j int) bool {
		if this[i].GetTx() == stgcommon.SYSTEM {
			return true
		}

		if this[j].GetTx() == stgcommon.SYSTEM {
			return false
		}
		return this[i].GetTx() < this[j].GetTx()
	})
	return this
}

func (this DeltaSequence) Finalize(store stgcommon.ReadOnlyStore) *univalue.Univalue {
	trans := []*univalue.Univalue(this)
	slice.RemoveIf(&trans, func(_ int, v *univalue.Univalue) bool {
		return v.GetPath() == nil
	})

	if len(this) == 0 {
		return nil
	}

	this.sort()

	// Use the first transition as the base value to apply the delta sets.
	if err := this[0].ApplyDelta(this[1:]); err != nil {
		panic(err)
	}

	// Remove the transition to indicate that the delta sequence has been finalized
	for i := 1; i < len(this); i++ {
		this[i].Property.SetPath(nil)
	}

	this = this[:1]
	return this[0]
}

func (this DeltaSequence) Finalized() *univalue.Univalue { return this[0] }

type DeltaSequences []DeltaSequence

func (this DeltaSequences) Finalized() []stgcommon.Type {
	return slice.Transform(this, func(_ int, v DeltaSequence) stgcommon.Type {
		return v[0].Value().(stgcommon.Type)
	})
}

func (this DeltaSequences) Keys() []*string {
	return slice.Transform(this, func(_ int, v DeltaSequence) *string {
		return v[0].GetPath()
	})
}
