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
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
)

type AsyncCacheWriter struct {
	*async.Pipeline[intf.Indexer[*univalue.Univalue]]
	*CacheIndexer
	blockNum uint64
}

func NewAsyncWriter(cache *ReadCache) *AsyncCacheWriter {
	blockNum := uint64(0) // TODO: get the block number from the block header
	idxer := NewCacheIndexer(cache)
	pipe := async.NewPipeline(
		4,
		10,
		// // Precommitter
		// func(idxer intf.Indexer[*univalue.Univalue]) (intf.Indexer[*univalue.Univalue], bool) {
		// 	idxer.Finalize(cache)
		// 	return idxer, true
		// },

		// db and cache writer
		func(idxer intf.Indexer[*univalue.Univalue]) (intf.Indexer[*univalue.Univalue], bool) {
			var err error
			idx := idxer.(*CacheIndexer)

			cache.BatchSet(idx.keys, idx.values) // update the local cache
			return nil, err == nil
		},
	)

	return &AsyncCacheWriter{
		Pipeline:     pipe.Start(),
		CacheIndexer: idxer,
		blockNum:     blockNum,
	}
}

func (this *AsyncCacheWriter) Add(univ []*univalue.Univalue) *AsyncCacheWriter {
	if len(univ) == 0 {
		this.CacheIndexer.Finalize(nil)
		this.Pipeline.Push(this.CacheIndexer) // push the indexer to the processor stream
	} else {
		this.CacheIndexer.Add(univ)
	}
	return this
}
