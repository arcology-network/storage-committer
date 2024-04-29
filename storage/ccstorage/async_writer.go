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

package ccstorage

import (
	async "github.com/arcology-network/common-lib/async"
	slice "github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/storage-committer/univalue"
)

// AsyncWriter is a struct that contains data strucuture and methods for writing data to concurrent storage asynchronously.
// It contains a pipeline that has a list of functions executing in order. Each function consumes the output of the previous function.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type AsyncWriter struct {
	*async.Pipeline[*CCIndexer]
	*CCIndexer
	store   *DataStore
	version uint64
}

func NewAsyncWriter(store *DataStore, version uint64) *AsyncWriter {
	// store := reader.(*DataStore)
	pipe := async.NewPipeline(
		"ccstorage",
		4,
		10,
		// Buffer the indexers in the pipeline, until an empty indexer is received.
		func(idxer *CCIndexer, buffer *[]*CCIndexer) ([]*CCIndexer, bool) {
			*buffer = append(*buffer, idxer)
			if idxer == nil {
				return slice.Move(buffer), true
			}
			return nil, false
		},

		// db and cache writer
		func(idxer *CCIndexer, buffer *[]*CCIndexer) ([]*CCIndexer, bool) {
			if idxer != nil {
				*buffer = append(*buffer, idxer)
				return nil, false
			}
			mergedIdxer := new(CCIndexer).Merge(*buffer)

			err := store.db.BatchSet(mergedIdxer.keyBuffer, mergedIdxer.encodedBuffer)
			store.cache.BatchSet(mergedIdxer.keyBuffer, mergedIdxer.valueBuffer) // update the local cache

			*buffer = (*buffer)[:0]
			return nil, err == nil
		},
	)

	return &AsyncWriter{
		Pipeline:  pipe.Start(),
		CCIndexer: NewCCIndexer(store, 0),
		store:     store,
		version:   version,
	}
}

func (this *AsyncWriter) Import(trans []*univalue.Univalue) {
	this.CCIndexer.Import(trans)
}

// Send the data to the downstream processor. This can be called multiple times
// before calling Await to commit the data to the state db.
func (this *AsyncWriter) Precommit() {
	this.CCIndexer.Finalize()          // Remove the nil transitions
	this.Pipeline.Push(this.CCIndexer) // push the indexer to the processor stream
	this.CCIndexer = NewCCIndexer(this.store, this.version)
}

// Await commits the data to the state db.
func (this *AsyncWriter) Commit(version uint64) {
	this.Pipeline.Push(nil) // commit all th indexers to the state db
	this.Pipeline.Await()
}

// Await commits the data to the state db.
func (this *AsyncWriter) Close() { this.Pipeline.Close() }
