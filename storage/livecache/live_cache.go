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
package livecache

import (
	"runtime"

	"github.com/arcology-network/common-lib/exp/associative"
	"github.com/arcology-network/common-lib/exp/slice"
	cache "github.com/arcology-network/common-lib/storage/cache"
	stgtype "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/type/univalue"

	// intf "github.com/arcology-network/storage-committer/interfaces"

	"github.com/cespare/xxhash/v2"
)

// ReadCache is a wrapper around cache.ReadCache with some extra methods provided
// by the intf.Datastore interface to work with the storage-committer.
type LiveCache struct {
	*cache.ReadCache[string, *associative.Pair[stgtype.Type, *Usage]] // Provide Readonly interface
	*CacheUsage                                                       // Memory usage of the cache.
}

func NewLiveCache() *LiveCache {
	cache := &LiveCache{
		ReadCache: cache.NewReadCache[string, *associative.Pair[stgtype.Type, *Usage]](
			4096, // 4096 shards to avoid lock contention
			func(v *associative.Pair[stgtype.Type, *Usage]) bool {
				return v == nil
			},
			func(k string) uint64 {
				return uint64(xxhash.Sum64String(k))
			},
		),
	}

	cache.CacheUsage = NewCacheUsage(cache) // To keep track of the cache memory usage.
	return cache
}

func (this *LiveCache) CacheChecksum() [32]byte {
	encoders := func(k string, v *associative.Pair[stgtype.Type, *Usage]) ([]byte, []byte) {
		return []byte(k), v.First.Encode()
	}

	less := func(k0, k1 string) bool {
		return k0 < k1
	}
	return this.ReadCache.Checksum(less, encoders)
}

func (this *LiveCache) Delete(keys []string) {
	this.ReadCache.BatchSet(keys, make([]*associative.Pair[stgtype.Type, *Usage], len(keys)))
}

func (this *LiveCache) Get(key string) (stgtype.Type, bool) {
	v, ok := this.ReadCache.Get(key)
	if !ok {
		return nil, ok
	}
	return (*v).First, ok
}

func (this *LiveCache) GetRaw(key string) (*associative.Pair[stgtype.Type, *Usage], bool) {
	v, ok := this.ReadCache.Get(key)
	if !ok {
		return nil, ok
	}
	return *v, ok
}

// func (this *LiveCache) Commit(keys []string, values []stgtype.Type) {
// 	pairedVals := make([]*associative.Pair[stgtype.Type, *Usage], len(values))
// 	for i, _ := range keys {
// 		if values[i] != nil {
// 			continue
// 		}

// 		memSize := uint64(0)
// 		if values[i] != nil {
// 			memSize = values[i].MemSize()
// 		}

// 		pairedVals[i] = &associative.Pair[stgtype.Type, *Usage]{
// 			First: values[i],
// 			Second: &Usage{
// 				sizeInMem: memSize,
// 				visits:    1,
// 			},
// 		}
// 	}
// 	this.ReadCache.Commit(keys, pairedVals) // update the local cache with the new values in the indexer
// }

func (this *LiveCache) Commit(keys []string, values []stgtype.Type, univals []*univalue.Univalue) {
	// Prepare the space for the new values in the cache, some univalues may be deleted because of the memory limit.
	if _, err := this.CacheUsage.PrepareSpace(&univals, this); err != nil {
		return
	}

	pairedVals := slice.ParallelTransform(univals, runtime.NumCPU(), func(i int, v *univalue.Univalue) *associative.Pair[stgtype.Type, *Usage] {
		if v.Value() == nil {
			return nil
		}

		// The entry may already exist in the cache, update the visits.
		accumVisits := uint64(0)
		metav, _ := this.GetRaw(*v.GetPath())
		if metav != nil {
			accumVisits = metav.Second.visits
		}

		return &associative.Pair[stgtype.Type, *Usage]{
			First: v.Value().(stgtype.Type),
			Second: &Usage{
				sizeInMem: v.Value().(stgtype.Type).MemSize(),
				visits:    accumVisits,
			},
		}
	})

	this.UpdateStats(univals)
	this.ReadCache.Commit(keys, pairedVals) // update the local cache with the new values in the indexer
}
