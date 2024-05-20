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
)

// AsyncWriter is a struct that contains data strucuture and methods for writing data to cache asynchronously.
// It contains a pipeline that has a list of functions executing in order. Each function consumes the output of the previous function.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type AsyncWriter struct {
	*async.Pipeline[*CacheIndexer]
	*CacheIndexer
	store *ReadCache
}

func NewAsyncWriter(cache *ReadCache, version int64) *AsyncWriter {
	idxer := NewCacheIndexer(cache, version)
	pipe := async.NewPipeline(
		"object cache",
		14,
		10,
		func(idxer *CacheIndexer, buffer *async.Slice[*CacheIndexer]) ([]*CacheIndexer, bool, bool) {
			if buffer.Append(idxer); idxer.Version < 0 {
				return nil, false, false
			}
			return buffer.MoveToSlice(), true, true
		},
		// Merge the indexers and update the cache at once.
		func(idxer *CacheIndexer, buffer *async.Slice[*CacheIndexer]) ([]*CacheIndexer, bool, bool) {
			if idxer.Version < 0 {
				buffer.Append(idxer)
				return nil, false, false
			}

			mergedIdxer := new(CacheIndexer).Merge(buffer.MoveToSlice()) // Merge indexers
			cache.BatchSet(mergedIdxer.keys, mergedIdxer.values)         // update the local cache with the new values in the indexer
			buffer.Clear()                                               // Clear the buffer
			return nil, false, true
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
	this.CacheIndexer.Finalize()                        // Remove the nil transitions
	this.Pipeline.Push(this.CacheIndexer)               // push the indexer to the processor stream
	this.CacheIndexer = NewCacheIndexer(this.store, -1) // Reset the indexer with a default version number
}

// Triggered by the block commit.
func (this *AsyncWriter) Commit(version uint64) {
	this.Pipeline.Push(&CacheIndexer{Version: int64(version)}) // commit all the indexers to the state db
	this.Pipeline.Await()
}
