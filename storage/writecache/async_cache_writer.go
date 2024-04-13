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
	"github.com/arcology-network/storage-committer/univalue"
)

// AsyncWriter is a struct that contains data strucuture and methods for writing data to cache asynchronously.
// It contains a pipeline that has a list of functions executing in order. Each function consumes the output of the previous function.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type AsyncWriter struct {
	*async.Pipeline[*WriteCacheIndexer]
	*WriteCacheIndexer
	cache *WriteCache
}

func NewAsyncWriter(cache *WriteCache, version uint64) *AsyncWriter {
	pipe := async.NewPipeline(
		4,
		10,
		// db writer
		func(idxers ...*WriteCacheIndexer) (*WriteCacheIndexer, bool) {
			cache.Insert(idxers[0].buffer) // update the write cache right away as soon as the indexer is received
			return nil, true
		},
	)

	return &AsyncWriter{
		Pipeline:          pipe.Start(),
		WriteCacheIndexer: NewWriteCacheIndexer(nil, version),
		cache:             cache,
	}
}

// Add adds a list of transitions to the indexer. If the list is empty, the indexer is finalized and pushed to the processor stream.
// The processor stream is a list of functions that will be executed in order, consuming the output of the previous function.
func (this *AsyncWriter) Add(univ []*univalue.Univalue) *AsyncWriter {
	if len(univ) == 0 {
		this.WriteCacheIndexer.Finalize()          // Remove the nil transitions because of conflict.
		this.Pipeline.Push(this.WriteCacheIndexer) // Send the indexer to the streamed processors.
	} else {
		this.WriteCacheIndexer.Add(univ) // Not all have been imported yet. Add them to the indexer
	}
	return this
}

// Called after each precommit to update the cache.
func (this *AsyncWriter) Feed() {
	this.WriteCacheIndexer.Finalize()          // Remove the nil transitions
	this.Pipeline.Push(this.WriteCacheIndexer) // push the indexer to the processor stream
}

// write cache updates itself every generation. It doesn't need to write to the database.
func (this *AsyncWriter) WriteToDB() {}
