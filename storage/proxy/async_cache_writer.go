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
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
)

// AsyncWriter is a struct that contains data strucuture and methods for writing data to cache asynchronously.
// It contains a pipeline that has a list of functions executing in order. Each function consumes the output of the previous function.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type AsyncWriter struct {
	*async.Pipeline[intf.Indexer[*univalue.Univalue]]
	*CacheIndexer
	blockNum uint64
}

func NewAsyncWriter(cache *ReadCache) *AsyncWriter {
	blockNum := uint64(0) // TODO: get the block number from the block header
	idxer := NewCacheIndexer(cache, 0)
	pipe := async.NewPipeline(
		4,
		10,
		func(idxer intf.Indexer[*univalue.Univalue]) bool { return idxer == nil },
		// db writer
		func(idxer intf.Indexer[*univalue.Univalue]) (intf.Indexer[*univalue.Univalue], bool) {
			var err error
			idx := idxer.(*CacheIndexer)

			cache.BatchSet(idx.keys, idx.values) // update the local cache
			return nil, err == nil
		},
	)

	return &AsyncWriter{
		Pipeline:     pipe.Start(),
		CacheIndexer: idxer,
		blockNum:     blockNum,
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

func (this *AsyncWriter) Await() { this.Pipeline.Await() }
