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
	"github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
)

type AsyncCommitter struct {
	*async.Triple[[]*univalue.Univalue, intf.Indexer[*univalue.Univalue], intf.Indexer[*univalue.Univalue], error]
	*CCIndexerV2
	store *DataStore
}

func NewAsyncCommitter(store interfaces.Datastore) *AsyncCommitter {
	idxer := NewCCIndexerV2(store)
	pipe := async.NewTriple(
		func(vals []*univalue.Univalue) (intf.Indexer[*univalue.Univalue], bool) {
			if len(vals) != 0 {
				idxer.Add(vals)
				return nil, false
			}

			// Finalize the indexer after whitelisting
			idxer.Finalize(store)
			return idxer, true
		},

		func(idxer intf.Indexer[*univalue.Univalue]) (intf.Indexer[*univalue.Univalue], bool) {
			idxer.Get()
			return idxer, true
		},

		func(idxer intf.Indexer[*univalue.Univalue]) (error, bool) {
			idx := idxer.(*CCIndexerV2)
			var err error
			if store.(*DataStore).keyCompressor != nil {
				common.ParallelExecute(
					func() { err = store.(*DataStore).db.BatchSet(idx.keyBuffer, idx.encodedBuffer) }, // Write data back
					func() { store.(*DataStore).keyCompressor.Commit() })

			} else {
				err = store.(*DataStore).db.BatchSet(idx.keyBuffer, idx.encodedBuffer)
			}

			store.(*DataStore).cccache.BatchSet(idx.keyBuffer, idx.valueBuffer) // update the local cache
			return err, true
		},
		64,
		10,
		1000,
	)

	return &AsyncCommitter{
		Triple:      pipe,
		CCIndexerV2: idxer,
		store:       store.(*DataStore),
	}
}
