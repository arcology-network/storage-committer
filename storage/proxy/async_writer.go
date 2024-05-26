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

// AsyncWriter is a struct that contains data strucuture and methods for writing data to cache.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type AsyncWriter struct {
	*CacheIndexer
	store  *ObjectCache
	buffer []*CacheIndexer
}

func NewAsyncWriter(cache *ObjectCache, version int64) *AsyncWriter {
	return &AsyncWriter{
		CacheIndexer: NewCacheIndexer(cache, version),
		store:        cache,
		buffer:       make([]*CacheIndexer, 0),
	}
}

// Send the data to the downstream processor, this is called for each generation.
// If there are multiple generations, this can be called multiple times before Await.
// Each generation
func (this *AsyncWriter) Precommit() {
	this.CacheIndexer.Finalize()                         // Remove the nil transitions
	this.buffer = append(this.buffer, this.CacheIndexer) // Append the indexer to the buffer
	this.CacheIndexer = NewCacheIndexer(this.store, -1)  // Reset the indexer with a default version number
}

// Triggered by the block commit.
func (this *AsyncWriter) Commit(version uint64) {
	mergedIdxer := new(CacheIndexer).Merge(this.buffer)       // Merge indexers
	this.store.BatchSet(mergedIdxer.keys, mergedIdxer.values) // update the local cache with the new values in the indexer
	this.buffer = make([]*CacheIndexer, 0)
}
