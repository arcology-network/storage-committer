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
	common "github.com/arcology-network/common-lib/common"
	intf "github.com/arcology-network/storage-committer/interfaces"
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

func NewAsyncWriter(reader intf.ReadOnlyDataStore) *AsyncWriter {
	store := reader.(*DataStore)
	version := uint64(0) // TODO: get the block number from the block header

	pipe := async.NewPipeline(
		4,
		10,
		// Buffer the indexers in the pipeline, until an empty indexer is received.
		func(indexers ...*CCIndexer) (*CCIndexer, bool) {
			return indexers[0], indexers[0] == nil
		},

		// db and cache writer
		func(indexers ...*CCIndexer) (*CCIndexer, bool) {
			mergedIdxer := new(CCIndexer).Merge(indexers)

			var err error
			if store.keyCompressor != nil {
				common.ParallelExecute(
					func() { err = store.db.BatchSet(mergedIdxer.keyBuffer, mergedIdxer.encodedBuffer) }, // Write data back
					func() { store.keyCompressor.Commit() })
			} else {
				err = store.db.BatchSet(mergedIdxer.keyBuffer, mergedIdxer.encodedBuffer)
			}

			store.cache.BatchSet(mergedIdxer.keyBuffer, mergedIdxer.valueBuffer) // update the local cache
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

// Send the data to the downstream processor. This can be called multiple times
// before calling Await to commit the data to the state db.
func (this *AsyncWriter) Feed() {
	this.CCIndexer.Finalize()          // Remove the nil transitions
	this.Pipeline.Push(this.CCIndexer) // push the indexer to the processor stream
	this.CCIndexer = NewCCIndexer(this.store, this.version)
}

// Await commits the data to the state db.
func (this *AsyncWriter) Write() {
	this.Pipeline.Push(nil) // commit all th indexers to the state db
	this.Pipeline.Await()
	this.version++
}
