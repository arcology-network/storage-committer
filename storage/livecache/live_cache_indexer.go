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
package livecache

import (
	"runtime"

	"github.com/arcology-network/common-lib/exp/slice"
	stgcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/type/univalue"
)

// LiveCacheIndexer is simpliest  of indexers. It does not index anything, just stores the transitions.
type LiveCacheIndexer struct {
	Version      int64
	buffer       []*univalue.Univalue
	importBuffer []*univalue.Univalue
	keys         []string
	values       []stgcommon.Type
	filter       func(*univalue.Univalue) bool
}

func NewLiveCacheIndexer(store *LiveCache, Version int64, filter func(*univalue.Univalue) bool) *LiveCacheIndexer {
	return &LiveCacheIndexer{
		Version:      Version,
		importBuffer: []*univalue.Univalue{},
		keys:         []string{},
		filter:       filter,
		values:       []stgcommon.Type{},
	}
}

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
func (this *LiveCacheIndexer) Import(transitions []*univalue.Univalue) {
	for i := range transitions {
		if this.filter(transitions[i]) {
			this.importBuffer = append(this.importBuffer, transitions[i])
		}
	}
}

func (this *LiveCacheIndexer) PreCommit() {
	this.buffer = this.importBuffer
	this.importBuffer = []*univalue.Univalue{}
}

func (this *LiveCacheIndexer) Finalize() {
	slice.RemoveIf((*[]*univalue.Univalue)(&this.buffer), func(i int, v *univalue.Univalue) bool { return v.GetPath() == nil })

	this.keys = make([]string, len(this.buffer))
	this.values = slice.ParallelTransform(this.buffer, runtime.NumCPU(), func(i int, v *univalue.Univalue) stgcommon.Type {
		this.keys[i] = *v.GetPath()
		if v.Value() != nil {
			return v.Value().(stgcommon.Type)
		}
		return nil // A deletion
	})
}

// Merge indexers so they can be updated at once. This is useful when working
// with multiple indexers at once.
func (this *LiveCacheIndexer) Merge(idxers []*LiveCacheIndexer) *LiveCacheIndexer {
	slice.Remove(&idxers, nil)

	this.buffer = slice.ConcateDo(idxers,
		func(idxer *LiveCacheIndexer) uint64 { return uint64(len(idxer.buffer)) },
		func(idxer *LiveCacheIndexer) []*univalue.Univalue { return idxer.buffer })

	// this.keys = slice.ConcateDo(idxers,
	// 	func(idxer *LiveCacheIndexer) uint64 { return uint64(len(idxer.keys)) },
	// 	func(idxer *LiveCacheIndexer) []string { return idxer.keys })

	// this.values = slice.ConcateDo(idxers,
	// 	func(idxer *LiveCacheIndexer) uint64 { return uint64(len(idxer.values)) },
	// 	func(idxer *LiveCacheIndexer) []stgcommon.Type { return idxer.values })

	return this
}
