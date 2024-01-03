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

	common "github.com/arcology-network/common-lib/common"
	committercommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/univalue"
)

type Arbitrator struct {
	groupIDs    []uint32
	transitions []*univalue.Univalue
}

func (this *Arbitrator) Insert(groupIDs []uint32, newTrans []*univalue.Univalue) int {
	this.transitions = append(this.transitions, newTrans...)
	this.groupIDs = append(this.groupIDs, groupIDs...)
	return len(this.groupIDs)
}

func (this *Arbitrator) Detect(groupIDs []uint32, newTrans []*univalue.Univalue) []*Conflict {
	if this.Insert(groupIDs, newTrans) == 0 {
		return []*Conflict{}
	}

	// t0 := time.Now()
	univalue.Univalues(newTrans).Sort(groupIDs)

	ranges := common.FindAllIndics(newTrans, func(lhv, rhv *univalue.Univalue) bool {
		return *lhv.GetPath() == *rhv.GetPath()
	})

	conflicts := []*Conflict{}
	for i := 0; i < len(ranges)-1; i++ {
		if ranges[i]+1 == ranges[i+1] {
			continue // Only one entry
		}

		offset := int(1)
		if newTrans[ranges[i]].Writes() == 0 {
			if newTrans[ranges[i]].IsConcurrentWritable() { // Delta write only
				offset = common.LocateFirstIf(newTrans[ranges[i]+1:ranges[i+1]],
					func(v *univalue.Univalue) bool {
						return !v.IsConcurrentWritable()
					})
			} else { // Read only
				offset = common.LocateFirstIf(newTrans[ranges[i]+1:ranges[i+1]],
					func(v *univalue.Univalue) bool {
						return v.Writes() > 0 || v.DeltaWrites() > 0
					})
			}
			offset = common.IfThen(offset < 0, ranges[i+1]-ranges[i], offset+1) // offset == -1 means no conflict found
		}

		if ranges[i]+offset == ranges[i+1] {
			continue
		}

		conflictTxs := []uint32{}
		common.Foreach(newTrans[ranges[i]+offset:ranges[i+1]], func(v **univalue.Univalue, _ int) {
			conflictTxs = append(conflictTxs, (*v).GetTx())
		})

		conflicts = append(conflicts,
			&Conflict{
				key:     *newTrans[ranges[i]].GetPath(),
				self:    newTrans[ranges[i]].GetTx(),
				groupID: groupIDs[ranges[i]+offset : ranges[i+1]],
				txIDs:   conflictTxs,
				Err:     errors.New(committercommon.WARN_ACCESS_CONFLICT),
			},
		)

		if len(conflicts) > 0 {
			if newTrans[ranges[i]].Writes() == 0 {
				if newTrans[ranges[i]].IsConcurrentWritable() { // Delta write only
					offset = common.LocateFirstIf(newTrans[ranges[i]+1:ranges[i+1]],
						func(v *univalue.Univalue) bool {
							return !v.IsConcurrentWritable()
						})
				} else { // Read only
					offset = common.LocateFirstIf(newTrans[ranges[i]+1:ranges[i+1]],
						func(v *univalue.Univalue) bool {
							return v.Writes() > 0 || v.DeltaWrites() > 0
						})
				}
				offset = common.IfThen(offset < 0, ranges[i+1]-ranges[i], offset+1) // offset == -1 means no conflict found
			}
		}

		dict := common.MapFromArray(conflictTxs, true) //Conflict dict
		trans := common.CopyIf(newTrans[ranges[i]+offset:ranges[i+1]], func(v *univalue.Univalue) bool { return (*dict)[v.GetTx()] })

		if outOfLimits := (&Accumulator{}).CheckMinMax(trans); outOfLimits != nil {
			conflicts = append(conflicts, outOfLimits...)
		}
	}

	// if len(conflicts) > 0 {
	// 	fmt.Println("range: ", ranges)
	// }
	return conflicts
}
