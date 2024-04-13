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
	async "github.com/arcology-network/common-lib/async"
	"github.com/arcology-network/storage-committer/univalue"
)

// AsyncWriter is a struct that contains data strucuture and methods for writing data to cache asynchronously.
// It contains a pipeline that has a list of functions executing in order. Each function consumes the output of the previous function.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type AsyncWriter struct {
	*async.Pipeline[*CacheIndexer]
	*CacheIndexer
	version uint64
}

func NewAsyncWriter(cache *ReadCache) *AsyncWriter {
	version := uint64(0) // TODO: get the block number from the block header
	idxer := NewCacheIndexer(cache, 0)
	pipe := async.NewPipeline(
		4,
		10,
		// Cache writer, update the cache as the indexers are received. This needs to be fixed
		// once the write cache is in use
		func(idxers ...*CacheIndexer) (*CacheIndexer, bool) {
			cache.BatchSet(idxers[0].keys, idxers[0].values) // update the local cache with the new values in the indexer
			return nil, true
		},
	)

	return &AsyncWriter{
		Pipeline:     pipe.Start(),
		CacheIndexer: idxer,
		version:      version,
	}
}

// Add adds a list of transitions to the indexer. If the list is empty, the indexer is finalized and pushed to the processor stream.
// The processor stream is a list of functions that will be executed in order, consuming the output of the previous function.
func (this *AsyncWriter) Add(univ []*univalue.Univalue) *AsyncWriter {
	if len(univ) == 0 {
		this.CacheIndexer.Finalize()
		this.Pipeline.Push(this.CacheIndexer) // push the indexer to the processor stream
	} else {
		this.CacheIndexer.Add(univ)
	}
	return this
}

func (this *AsyncWriter) WriteToDB() {
	this.Pipeline.Await()
}
