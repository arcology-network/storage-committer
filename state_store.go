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
	intf "github.com/arcology-network/storage-committer/interfaces"
	stgcomm "github.com/arcology-network/storage-committer/storage/committer"
	proxy "github.com/arcology-network/storage-committer/storage/proxy"
	writecache "github.com/arcology-network/storage-committer/storage/writecache"
	"github.com/arcology-network/storage-committer/univalue"
	"github.com/cespare/xxhash/v2"
	//  "github.com/arcology-network/storage-committer/storage/proxy"
)

// Buffer is simpliest  of indexers. It does not index anything, just stores the transitions.
type StateStore struct {
	// *writecache.ShardedWriteCache
	*writecache.WriteCache // execution cache
	*stgcomm.StateCommitter
	backend *proxy.StorageProxy
}

// New creates a new StateCommitter instance.
func NewStateStore(backend *proxy.StorageProxy) *StateStore {
	store := &StateStore{
		backend: backend,
		WriteCache: writecache.NewWriteCache(
			backend,
			16,
			1,
			func(k string) uint64 {
				return xxhash.Sum64String(k)
			},
		),
	}
	store.StateCommitter = stgcomm.NewStateCommitter(backend, store.GetWriters())
	return store
}

func (this *StateStore) Backend() *proxy.StorageProxy    { return this.backend }
func (this *StateStore) Cache() *writecache.WriteCache   { return this.WriteCache }
func (this *StateStore) Import(trans univalue.Univalues) { this.StateCommitter.Import(trans) }
func (this *StateStore) Preload(key []byte) interface{}  { return this.backend.Preload(key) }
func (this *StateStore) Clear()                          { this.WriteCache.Clear() }

func (this *StateStore) GetWriters() []intf.AsyncWriter[*univalue.Univalue] {
	return append([]intf.AsyncWriter[*univalue.Univalue]{
		writecache.NewExecutionCacheWriter(this.WriteCache, -1)},
		this.backend.GetWriters()...)
}
