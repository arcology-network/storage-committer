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

package ccstorage

import (
	"runtime"

	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/univalue"
)

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
type CCIndexer []*univalue.Univalue

func NewIndexer(store interfaces.Datastore) intf.Indexer[*univalue.Univalue] {
	return (*CCIndexer)(&[]*univalue.Univalue{})
}

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
func (this *CCIndexer) Add(transitions []*univalue.Univalue) {
	for _, v := range transitions {
		if v.GetPath() != nil || !platform.IsEthPath(*v.GetPath()) {
			*this = append(*this, v)
		}
	}
}

func (this *CCIndexer) Get() interface{} {
	keys := make([]string, len(*this))
	tVals := slice.ParallelTransform(*this, runtime.NumCPU(), func(i int, v *univalue.Univalue) interface{} {
		keys[i] = *v.GetPath()
		if v.Value() != nil {
			return v.Value().(intf.Type)
		}
		return nil // A deletion
	})
	return []interface{}{keys, tVals}
}

func (this *CCIndexer) Finalize() {
	slice.RemoveIf((*[]*univalue.Univalue)(this), func(_ int, v *univalue.Univalue) bool { return v.GetPath() == nil }) // Remove the transitions that are marked
}

func (this *CCIndexer) Clear() {
	*this = (*this)[:0]
}
