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
package proxy

import (
	"runtime"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
)

// Buffer is simpliest  of indexers. It does not index anything, just stores the transitions.
type Buffer []*univalue.Univalue

func NewIndexer(store intf.Datastore) *Buffer {
	return (*Buffer)(common.Reference([]*univalue.Univalue{}))
}

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
func (this *Buffer) Add(transitions []*univalue.Univalue) { *this = append(*this, transitions...) }

func (this *Buffer) Get() interface{} {
	keys := make([]string, len(*this))
	tVals := slice.ParallelTransform(*this, runtime.NumCPU(), func(i int, v *univalue.Univalue) intf.Type {
		keys[i] = *v.GetPath()
		if v.Value() != nil {
			return v.Value().(intf.Type)
		}
		return nil // A deletion
	})
	return []interface{}{keys, tVals}
}

func (this *Buffer) Finalize(_ intf.CommittableStore) {
	slice.RemoveIf((*[]*univalue.Univalue)(this), func(i int, v *univalue.Univalue) bool { return v.GetPath() == nil })
}
func (this *Buffer) Clear() { *this = (*this)[:0] }
