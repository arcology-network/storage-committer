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

package statestore

import (
	// "github.com/arcology-network/concurrenturl/commutative"
	intf "github.com/arcology-network/storage-committer/common"
	stgcommon "github.com/arcology-network/storage-committer/common"
	committer "github.com/arcology-network/storage-committer/storage/committer"
	"github.com/arcology-network/storage-committer/type/commutative"

	cache "github.com/arcology-network/storage-committer/storage/cache"
	proxy "github.com/arcology-network/storage-committer/storage/proxy"
	"github.com/arcology-network/storage-committer/type/univalue"
	"github.com/cespare/xxhash/v2"
)

// Buffer is simpliest  of indexers. It does not index anything, just stores the transitions.
type StateStore struct {
	// *cache.ShardedWriteCache
	*cache.WriteCache // execution cache
	*committer.StateCommitter
	backend *proxy.StorageProxy
}

// New creates a new StateCommitter instance.
func NewStateStore(backend *proxy.StorageProxy) *StateStore {
	store := &StateStore{
		backend: backend,
		WriteCache: cache.NewWriteCache(
			backend,
			16,
			1,
			func(k string) uint64 {
				return xxhash.Sum64String(k)
			},
		),
	}
	store.StateCommitter = committer.NewStateCommitter(store.WriteCache, store.GetWriters())

	// Commit initial transitions to the store if any.
	initTrans := []*univalue.Univalue{
		univalue.NewUnivalue(stgcommon.SYSTEM, stgcommon.GAS_PREPAYERS, 0, 1, 0, commutative.NewPath(), nil),
	}

	for _, tran := range initTrans {
		tran.SkipConflictCheck(true) // Skip conflict check for initial transitions
	}

	committer := committer.NewStateCommitter(store, store.GetWriters())
	committer.Import(initTrans)
	committer.Precommit([]uint64{stgcommon.SYSTEM})
	committer.Commit(0)
	return store
}

func (this *StateStore) Backend() *proxy.StorageProxy    { return this.backend }
func (this *StateStore) Cache() *cache.WriteCache        { return this.WriteCache }
func (this *StateStore) Import(trans univalue.Univalues) { this.StateCommitter.Import(trans) }
func (this *StateStore) Preload(key []byte) any          { return this.backend.Preload(key) }
func (this *StateStore) Clear()                          { this.WriteCache.Clear() }

func (this *StateStore) GetWriters() []intf.AsyncWriter[*univalue.Univalue] {
	return append([]intf.AsyncWriter[*univalue.Univalue]{
		cache.NewExecutionCacheWriter(this.WriteCache, -1)},
		this.backend.GetWriters()...)
}
