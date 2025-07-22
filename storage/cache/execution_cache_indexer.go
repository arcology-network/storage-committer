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
package cache

import (
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/type/univalue"
)

// ExecutionCacheIndexer is the simpliest of all indexers. It does not index anything, just stores the transitions.
type ExecutionCacheIndexer struct {
	version int64
	buffer  []*univalue.Univalue
	filter  func(tran *univalue.Univalue) bool
}

func NewExecutionCacheIndexer(_ *intf.ReadOnlyStore, version int64, filter func(tran *univalue.Univalue) bool) *ExecutionCacheIndexer {
	return &ExecutionCacheIndexer{
		filter:  common.IfThen(filter == nil, func(tran *univalue.Univalue) bool { return true }, filter),
		version: version,
		buffer:  []*univalue.Univalue{},
	}
}

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
func (this *ExecutionCacheIndexer) Import(transitions []*univalue.Univalue) {
	this.buffer = append(this.buffer, transitions...)
}

// Remove nil transitions due to conflicts.
func (this *ExecutionCacheIndexer) Finalize() {
	slice.RemoveIf((*[]*univalue.Univalue)(&this.buffer), func(i int, v *univalue.Univalue) bool { return v.GetPath() == nil })
}
