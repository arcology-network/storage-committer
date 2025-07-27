/*
 *   Copyright (c) 2025 Arcology Network

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

package statestore

import (
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/type/commutative"
	"github.com/arcology-network/storage-committer/type/univalue"
)

// SubstitueWildcards substitutes the wildcards in the transitions with the actual transitions.
func (this *StateCommitter) SubstitueWildcards(transitions []*univalue.Univalue) []*univalue.Univalue {
	wildcards := slice.MoveIf(&transitions, func(_ int, v *univalue.Univalue) bool {
		flag, _ := v.IsWildcard()
		return flag
	})

	substitued := []*univalue.Univalue{}
	for i := range wildcards {
		parentPath := common.GetParentPath(*wildcards[i].GetPath())           // Get the parent path of the wildcard, so that it can be used for cascade delete.
		v, err := this.Store().ReadStorage(parentPath, new(commutative.Path)) // Preload the path, so that it can be used for cascade delete.
		if v == nil || err != nil {
			continue
		}

		// Get the sub paths of the path. including the child paths.
		pathStrs := v.(stgcommon.Type).GetCascadeSub(parentPath, this.Store())
		trans := slice.Transform(pathStrs, func(_ int, substitued string) *univalue.Univalue {
			newTran := wildcards[i].Clone().(*univalue.Univalue)
			newTran.SetPath(&substitued) // Set the path to the sub path.
			newTran.SetValue(nil)        // Set the value to nil, so that it won't be indexed
			return newTran
		})
		substitued = append(substitued, trans...)
	}
	transitions = append(transitions, substitued...) // Append the substituted transitions to the original transitions.
	return transitions
}
