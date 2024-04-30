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
	"math"

	async "github.com/arcology-network/common-lib/async"
	"github.com/arcology-network/common-lib/exp/slice"
)

// AsyncWriter is a struct that contains data strucuture and methods for writing data to cache asynchronously.
// It contains a pipeline that has a list of functions executing in order. Each function consumes the output of the previous function.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type AsyncWriter struct {
	*async.Pipeline[*CacheIndexer]
	*CacheIndexer
	store *ReadCache
}

func NewAsyncWriter(cache *ReadCache, version uint64) *AsyncWriter {
	idxer := NewCacheIndexer(cache, 0)
	pipe := async.NewPipeline(
		"object cache",
		4,
		10,
		func(idxer *CacheIndexer, buffer *[]*CacheIndexer) ([]*CacheIndexer, bool) {
			*buffer = append(*buffer, idxer) // Buffer the indexers until the final one is received
			if idxer.Version == math.MaxUint64 {
				return nil, false
			}
			v := slice.Move(buffer)
			return v, true
		},
		// Merge the indexers and update the cache at once.
		func(idxer *CacheIndexer, buffer *[]*CacheIndexer) ([]*CacheIndexer, bool) {
			if idxer.Version == math.MaxUint64 {
				*buffer = append(*buffer, idxer) // Buffer the indexers until the final one is received
				return nil, false
			}

			mergedIdxer := new(CacheIndexer).Merge(*buffer)      // Merge indexers
			cache.BatchSet(mergedIdxer.keys, mergedIdxer.values) // update the local cache with the new values in the indexer
			*buffer = (*buffer)[:0]                              // Clear the buffer
			return nil, false
		},
	)

	return &AsyncWriter{
		Pipeline:     pipe.Start(),
		CacheIndexer: idxer,
		store:        cache,
	}
}

// Send the data to the downstream processor, this is called for each generation.
// If there are multiple generations, this can be called multiple times before Await.
// Each generation
func (this *AsyncWriter) Precommit() {
	this.CacheIndexer.Finalize()                                    // Remove the nil transitions
	this.Pipeline.Push(this.CacheIndexer)                           // push the indexer to the processor stream
	this.CacheIndexer = NewCacheIndexer(this.store, math.MaxUint64) // Reset the indexer with a default version number
}

// Triggered by the block commit.
func (this *AsyncWriter) Commit(version uint64) {
	this.Pipeline.Push(&CacheIndexer{Version: version}) // commit all the indexers to the state db
	this.Pipeline.Await()
}
