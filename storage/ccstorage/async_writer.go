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
	"github.com/arcology-network/storage-committer/univalue"
)

// AsyncWriter is a struct that contains data strucuture and methods for writing data to concurrent storage asynchronously.
// It contains a pipeline that has a list of functions executing in order. Each function consumes the output of the previous function.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type AsyncWriter struct {
	*async.Pipeline[intf.Indexer[*univalue.Univalue]]
	*CCIndexer
	store    *DataStore
	blockNum uint64
}

func NewAsyncWriter(store *DataStore) *AsyncWriter {
	blockNum := uint64(0) // TODO: get the block number from the block header
	idxer := NewCCIndexer(store)
	pipe := async.NewPipeline(
		4,
		10,
		// Precommitter
		func(idxer intf.Indexer[*univalue.Univalue]) (intf.Indexer[*univalue.Univalue], bool) {
			idxer.Finalize(store)
			return idxer, true
		},

		// db and cache writer
		func(idxer intf.Indexer[*univalue.Univalue]) (intf.Indexer[*univalue.Univalue], bool) {
			var err error
			idx := idxer.(*CCIndexer)
			if store.keyCompressor != nil {
				common.ParallelExecute(
					func() { err = store.db.BatchSet(idx.keyBuffer, idx.encodedBuffer) }, // Write data back
					func() { store.keyCompressor.Commit() })

			} else {
				err = store.db.BatchSet(idx.keyBuffer, idx.encodedBuffer)
			}

			store.cccache.BatchSet(idx.keyBuffer, idx.valueBuffer) // update the local cache
			return nil, err == nil
		},
	)

	return &AsyncWriter{
		Pipeline:  pipe.Start(),
		CCIndexer: idxer,
		store:     store,
		blockNum:  blockNum,
	}
}

func (this *AsyncWriter) Add(univ []*univalue.Univalue) *AsyncWriter {
	if len(univ) == 0 {
		this.CCIndexer.Finalize(nil)
		this.Pipeline.Push(this.CCIndexer) // push the indexer to the processor stream
	} else {
		this.CCIndexer.Add(univ)
	}
	return this
}
