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
	cache "github.com/arcology-network/common-lib/storage/cache"
	policy "github.com/arcology-network/common-lib/storage/policy"
	stgtype "github.com/arcology-network/storage-committer/common"

	// intf "github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/common"
	"github.com/cespare/xxhash/v2"
)

// ReadCache is a wrapper around cache.ReadCache with some extra methods provided
// by the intf.Datastore interface to work with the storage-committer.
type LiveCache struct {
	*cache.ReadCache[string, stgtype.Type] // Provide Readonly interface
	// queue                                  chan *associative.Pair[[]string, []stgtype.Type]
}

func NewReadCache(store intf.ReadOnlyStore) *LiveCache {
	return &LiveCache{
		cache.NewReadCache[string, stgtype.Type](
			4096, // 4096 shards to avoid lock contention
			func(v stgtype.Type) bool {
				return v == nil
			},
			func(k string) uint64 {
				return uint64(xxhash.Sum64String(k))
			},
			policy.NewCachePolicy(0, 0),
		),
		// make(chan *associative.Pair[[]string, []stgtype.Type], 16),
	}
}

func (this *LiveCache) CacheChecksum() [32]byte {
	encoders := func(k string, v intf.Type) ([]byte, []byte) {
		return []byte(k), v.Encode()
	}

	less := func(k0, k1 string) bool {
		return k0 < k1
	}
	return this.Checksum(less, encoders)
}
