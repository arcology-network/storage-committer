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

package importer

import (
	"errors"
	"sort"

	"github.com/arcology-network/common-lib/exp/array"
	committercommon "github.com/arcology-network/concurrenturl/common"
	intf "github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/concurrenturl/univalue"
)

type Accumulator struct{}

// check if the value is out of limits. It sorts the transitions by the delta sign and type.
// The neative delta is in the front of the array to make sure it has sufficient balance for the positive deltas.
// The underflow is always checked first before the overflow.
func (this *Accumulator) CheckMinMax(transitions []*univalue.Univalue) []*Conflict {
	if len(transitions) <= 1 ||
		(transitions)[0].Value() == nil ||
		!(transitions)[0].Value().(intf.Type).IsCommutative() ||
		!(transitions)[0].Value().(intf.Type).IsNumeric() {
		return nil
	}

	array.RemoveIf(&transitions, func(_ int, v *univalue.Univalue) bool {
		return v.IsReadOnly()
	})

	if len(transitions) <= 1 {
		return nil
	}

	sort.SliceStable(transitions, func(i, j int) bool {
		lhv := transitions[i].Value().(intf.Type)
		rhv := transitions[i].Value().(intf.Type)
		return lhv.DeltaSign() != rhv.DeltaSign() && !lhv.DeltaSign()
	})

	negatives, positives := this.Categorize(transitions)

	// check underflow first
	underflowed := this.isOutOfLimits(*(transitions)[0].GetPath(), negatives)
	if underflowed != nil {
		underflowed.Err = errors.New(committercommon.WARN_OUT_OF_LOWER_LIMIT)
	}

	// check overflow
	overflowed := this.isOutOfLimits(*(transitions)[0].GetPath(), positives)
	if overflowed != nil {
		overflowed.Err = errors.New(committercommon.WARN_OUT_OF_UPPER_LIMIT)
	}

	if overflowed == nil && underflowed == nil {
		return nil
	}

	conflicts := []*Conflict{}
	if underflowed != nil {
		conflicts = append(conflicts, underflowed)
	}

	if overflowed != nil {
		conflicts = append(conflicts, overflowed)
	}
	return conflicts
}

// categorize transitions into two groups, one is negative, the other is positive.
func (*Accumulator) Categorize(transitions []*univalue.Univalue) ([]*univalue.Univalue, []*univalue.Univalue) {
	offset, _ := array.FindFirstIf(transitions, func(v *univalue.Univalue) bool { return v.Value().(intf.Type).DeltaSign() })

	if offset < 0 {
		offset = len(transitions)
	}
	return transitions[:offset], transitions[offset:]
}

// check if the value is out of limits
func (this *Accumulator) isOutOfLimits(k string, transitions []*univalue.Univalue) *Conflict {
	if len(transitions) <= 1 {
		return nil
	}

	initialv := transitions[0].Value().(intf.Type).Clone().(intf.Type)

	typedVals := array.Append(transitions, func(_ int, v *univalue.Univalue) intf.Type {
		return v.Value().(intf.Type)
	})

	_, length, err := initialv.ApplyDelta(typedVals[1:])
	if err == nil {
		return nil
	}

	txIDs := []uint32{}
	array.Foreach(transitions[length+1:], func(_ int, v **univalue.Univalue) { txIDs = append(txIDs, (*v).GetTx()) })

	return &Conflict{
		key:   k,
		txIDs: txIDs,
	}
}
