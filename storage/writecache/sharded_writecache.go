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

// WriteCache is a read-only data store used for caching.

package cache

import (
	"runtime"

	slice "github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
)

const (
	NUM_SHARDS = 32
)

// ShardedWriteCache is a lockless data strucuture that wraps multiple WriteCache instances together, each of
// which is responsible for a subset of the data. It can be updated in parallel when a transaction generation
// is completed. But it isn't thread-safe.
type ShardedWriteCache struct {
	backend intf.ReadOnlyDataStore
	caches  [NUM_SHARDS]*WriteCache
	hasher  func(string) uint64
}

func NewShardedWriteCache(backend intf.ReadOnlyDataStore, perPage int, numPages int, hasher func(string) uint64, args ...interface{}) *ShardedWriteCache {
	writeCache := &ShardedWriteCache{
		backend: backend,
		hasher:  hasher,
	}

	for i := 0; i < len(writeCache.caches); i++ {
		writeCache.caches[i] = NewWriteCache(backend, perPage, numPages, args...)
	}
	return writeCache
}

func (this *ShardedWriteCache) ReadOnlyDataStore() intf.ReadOnlyDataStore { return this.backend }
func (this *ShardedWriteCache) Cache() [NUM_SHARDS]*WriteCache            { return this.caches }

func (this *ShardedWriteCache) NewUnivalue(k string) *univalue.Univalue {
	return this.caches[this.hasher(k)].NewUnivalue()
}

func (this *ShardedWriteCache) GetOrNew(tx uint32, path string, T any) (*univalue.Univalue, bool) {
	return this.caches[this.hasher(path)%NUM_SHARDS].GetOrNew(tx, path, T)
}

func (this *ShardedWriteCache) Read(tx uint32, path string, T any) (interface{}, interface{}, uint64) {
	return this.caches[this.hasher(path)%NUM_SHARDS].Read(tx, path, T)
}

func (this *ShardedWriteCache) Write(tx uint32, path string, value interface{}) (int64, error) {
	return this.caches[this.hasher(path)%NUM_SHARDS].Write(tx, path, value)
}

func (this *ShardedWriteCache) InCache(path string) (interface{}, bool) {
	return this.caches[this.hasher(path)%NUM_SHARDS].InCache(path)
}

func (this *ShardedWriteCache) Retrive(path string, T any) (interface{}, error) {
	return this.caches[this.hasher(path)%NUM_SHARDS].Retrive(path, T)
}

func (this *ShardedWriteCache) RetriveFromStorage(path string, T any) (interface{}, error) {
	return this.caches[this.hasher(path)%NUM_SHARDS].RetriveFromStorage(path, T)
}

func (this *ShardedWriteCache) IfExists(path string) bool {
	return this.caches[this.hasher(path)%NUM_SHARDS].IfExists(path)
}

func (this *ShardedWriteCache) Insert(transitions []*univalue.Univalue) *ShardedWriteCache {
	univalue.Univalues(transitions).SortByDepth() // To ensure that the parent  is inserted before the child

	// Precalculate the shard ID of each transition
	shards := slice.ParallelTransform(transitions, runtime.NumCPU(), func(i int, v *univalue.Univalue) uint64 {
		return this.hasher(*(v).GetPath())
	})

	// Insert each transition into the appropriate cache
	slice.ParallelForeach(this.caches[:], runtime.NumCPU(), func(num int, shard **WriteCache) {
		for i := 0; i < len(transitions); i++ {
			if shards[i] == uint64(num) {
				this.caches[num].set(transitions[i])
			}
		}
	})
	return this
}

// Reset the writecache to the initial state for the next round of processing.
func (this *ShardedWriteCache) Precommit(args ...interface{}) [NUM_SHARDS]byte {
	this.Insert(args[0].([]*univalue.Univalue))
	return [NUM_SHARDS]byte{}
}

func (this *ShardedWriteCache) Clear() *ShardedWriteCache {
	slice.ParallelForeach(this.caches[:], runtime.NumCPU(), func(i int, wcache **WriteCache) {
		(*wcache).Clear()
	})
	return this
}

func (this *ShardedWriteCache) Equal(other *ShardedWriteCache) bool {
	for i := 0; i < len(this.caches); i++ {
		if !this.caches[i].Equal(other.caches[i]) {
			return false
		}
	}
	return true
}

func (this *ShardedWriteCache) KVs() ([][]string, [][]intf.Type) {
	keySet, valueSet := make([][]string, len(this.caches)), make([][]intf.Type, len(this.caches))
	for i := 0; i < len(this.caches); i++ {
		keySet[i], valueSet[i] = this.caches[i].KVs()
	}
	return keySet, valueSet
}

func (this *ShardedWriteCache) Export(preprocs ...func([]*univalue.Univalue) []*univalue.Univalue) []*univalue.Univalue {
	valueSet := make([][]*univalue.Univalue, len(this.caches))
	for i := 0; i < len(this.caches); i++ {
		valueSet[i] = this.caches[i].Export()
	}
	return slice.Flatten(valueSet)
}
