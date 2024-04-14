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
	version uint64
	buffer  []*univalue.Univalue
	keys    []string
	values  []intf.Type
}

func NewCacheIndexer(store *ReadCache, version uint64) *CacheIndexer {
	return &CacheIndexer{
		version: version,
		buffer:  []*univalue.Univalue{},
		keys:    []string{},
		values:  []intf.Type{},
	}
}

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
func (this *CacheIndexer) Add(transitions []*univalue.Univalue) {
	this.buffer = append(this.buffer, transitions...)
}

func (this *CacheIndexer) Finalize() {
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

// Merge indexers so they can be updated at once.
func (this *CacheIndexer) Merge(idxers []*CacheIndexer) *CacheIndexer {
	slice.Remove(&idxers, nil)

	this.buffer = slice.ConcateDo(idxers,
		func(idxer *CacheIndexer) uint64 { return uint64(len(idxer.buffer)) },
		func(idxer *CacheIndexer) []*univalue.Univalue { return idxer.buffer })

	this.keys = slice.ConcateDo(idxers,
		func(idxer *CacheIndexer) uint64 { return uint64(len(idxer.keys)) },
		func(idxer *CacheIndexer) []string { return idxer.keys })

	this.values = slice.ConcateDo(idxers,
		func(idxer *CacheIndexer) uint64 { return uint64(len(idxer.values)) },
		func(idxer *CacheIndexer) []intf.Type { return idxer.values })

	return this
}
