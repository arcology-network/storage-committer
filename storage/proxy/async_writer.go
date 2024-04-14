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
	store   *ReadCache
	version uint64
}

func NewAsyncWriter(cache *ReadCache, version uint64) *AsyncWriter {
	idxer := NewCacheIndexer(cache, 0)
	pipe := async.NewPipeline(
		4,
		10,
		func(idxers ...*CacheIndexer) (*CacheIndexer, bool) {
			if len(idxers) == 0 || idxers[0] == nil {
				return nil, true
			}
			return idxers[0], false // Buffer the indexers until the final indexer is received
		},
		// Merge the indexers and update the cache at once.
		func(idxers ...*CacheIndexer) (*CacheIndexer, bool) {
			mergedIdxer := new(CacheIndexer).Merge(idxers)       // Merge indexers
			cache.BatchSet(mergedIdxer.keys, mergedIdxer.values) // update the local cache with the new values in the indexer
			return nil, true
		},
	)

	return &AsyncWriter{
		Pipeline:     pipe.Start(),
		CacheIndexer: idxer,
		store:        cache,
		version:      version,
	}
}

// Add adds a list of transitions to the indexer. If the list is empty, the indexer is finalized and pushed to the processor stream.
// The processor stream is a list of functions that will be executed in order, consuming the output of the previous function.
// func (this *AsyncWriter) Add(univ []*univalue.Univalue) *AsyncWriter {
// 	if len(univ) == 0 {
// 		this.CacheIndexer.Finalize()
// 		this.Pipeline.Push(this.CacheIndexer) // push the indexer to the processor stream
// 	} else {
// 		this.CacheIndexer.Add(univ)
// 	}
// 	return this
// }

// Send the data to the downstream processor, this is called for each generation.
// If there are multiple generations, this can be called multiple times before Await.
// Each generation
func (this *AsyncWriter) Feed() *AsyncWriter {
	this.CacheIndexer.Finalize()                                  // Remove the nil transitions
	this.Pipeline.Push(this.CacheIndexer)                         // push the indexer to the processor stream
	this.CacheIndexer = NewCacheIndexer(this.store, this.version) // Reset the indexer
	return this
}

// Triggered by the block commit.
func (this *AsyncWriter) Write() {
	this.Pipeline.Push(nil) // commit all the indexers to the state db
	this.Pipeline.Await()
}
