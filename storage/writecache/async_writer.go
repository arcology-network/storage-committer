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

package cache

import (
	async "github.com/arcology-network/common-lib/async"
)

// AsyncWriter is a struct that contains data strucuture and methods for writing data to cache asynchronously.
// It contains a pipeline that has a list of functions executing in order. Each function consumes the output of the previous function.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type AsyncWriter struct {
	*async.Pipeline[*WriteCacheIndexer]
	*WriteCacheIndexer
	*WriteCache
}

func NewAsyncWriter(cache *WriteCache, version int64) *AsyncWriter {
	pipe := async.NewPipeline(
		"WriteCache",
		4,
		10,
		// db writer
		func(idxer *WriteCacheIndexer, _ *[]*WriteCacheIndexer) ([]*WriteCacheIndexer, bool) {
			cache.Insert(idxer.buffer) // update the write cache right away as soon as the indexer is received
			return nil, true
		},
	)

	return &AsyncWriter{
		Pipeline:          pipe.Start(),
		WriteCacheIndexer: NewWriteCacheIndexer(nil, int64(version)),
		WriteCache:        cache,
	}
}

// write cache updates itself every generation. It doesn't need to write to the database.
func (this *AsyncWriter) Precommit() {
	this.WriteCacheIndexer.Finalize()          // Remove the nil transitions
	this.Pipeline.Push(this.WriteCacheIndexer) // push the indexer to the processor stream
	this.Pipeline.Await()
	this.WriteCacheIndexer = NewWriteCacheIndexer(nil, -1)
}

// The generation cache is transient and will clear itself when all the transitions are committed to
// the database.
func (this *AsyncWriter) Commit(_ uint64) {
	this.WriteCache.Clear()
}
