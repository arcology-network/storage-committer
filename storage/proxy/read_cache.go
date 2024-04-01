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
	cache "github.com/arcology-network/common-lib/storage/cache"
	policy "github.com/arcology-network/common-lib/storage/policy"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/cespare/xxhash/v2"
)

// ReadCache is a wrapper around cache.ReadCache with some extra methods provided
// by the intf.Datastore interface to work with the storage-committer.
type ReadCache struct {
	*cache.ReadCache[string, intf.Type] // Provide Readonly interface
}

func NewReadCache(store intf.Datastore) *ReadCache {
	return &ReadCache{cache.NewReadCache[string, intf.Type](
		4096, // 4096 shards to avoid lock contention
		func(v intf.Type) bool {
			return v == nil
		},
		func(k string) uint64 {
			return uint64(xxhash.Sum64String(k))
		},
		policy.NewCachePolicy(0, 0),
	)}
}

// Read cache does not have a precommit method equivalent to the write cache
func (this *ReadCache) Precommit(args ...interface{}) [32]byte {
	kvs := args[0].(*Buffer).Get().([]interface{})
	this.ReadCache.Commit(kvs[0].([]string), kvs[1].([]intf.Type))
	return [32]byte{}
}

func (this *ReadCache) Commit(placeHolder uint64) error { return nil }
