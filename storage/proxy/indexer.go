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

	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
)

// CacheIndexer is simpliest  of indexers. It does not index anything, just stores the transitions.
type CacheIndexer struct {
	buffer []*univalue.Univalue
	keys   []string
	values []intf.Type
}

func NewCacheIndexer(store *ReadCache) *CacheIndexer {
	return &CacheIndexer{
		buffer: []*univalue.Univalue{},
		keys:   []string{},
		values: []intf.Type{},
	}
}

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
func (this *CacheIndexer) Add(transitions []*univalue.Univalue) {
	this.buffer = append(this.buffer, transitions...)
}

func (this *CacheIndexer) Finalize(_ intf.CommittableStore) {
	slice.RemoveIf((*[]*univalue.Univalue)(&this.buffer), func(i int, v *univalue.Univalue) bool { return v.GetPath() == nil })

	this.keys = make([]string, len(this.buffer))
	this.values = slice.ParallelTransform(this.buffer, runtime.NumCPU(), func(i int, v *univalue.Univalue) intf.Type {
		this.keys[i] = *v.GetPath()
		if v.Value() != nil {
			return v.Value().(intf.Type)
		}
		return nil // A deletion
	})
}
func (this *CacheIndexer) Clear() {
	this.buffer = this.buffer[:0]
	this.keys = this.keys[:0]
	this.values = this.values[:0]
}
