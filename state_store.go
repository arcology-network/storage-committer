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
	stgcomm "github.com/arcology-network/storage-committer/committer"
	intf "github.com/arcology-network/storage-committer/interfaces"
	proxy "github.com/arcology-network/storage-committer/storage/proxy"
	writecache "github.com/arcology-network/storage-committer/storage/writecache"
	"github.com/arcology-network/storage-committer/univalue"
	"github.com/cespare/xxhash/v2"
	//  "github.com/arcology-network/storage-committer/storage/proxy"
)

// Buffer is simpliest  of indexers. It does not index anything, just stores the transitions.
type StateStore struct {
	// *writecache.ShardedWriteCache
	*writecache.WriteCache
	*stgcomm.StateCommitter
	store *proxy.StorageProxy
}

// New creates a new StateCommitter instance.
func NewStateStore(store *proxy.StorageProxy) *StateStore {
	return &StateStore{
		store: store,
		WriteCache: writecache.NewWriteCache(
			store,
			16,
			1,
			func(k string) uint64 {
				return xxhash.Sum64String(k)
			},
		),
		StateCommitter: stgcomm.NewStateCommitter(store),
	}
}

func (this *StateStore) Store() *proxy.StorageProxy      { return this.store }
func (this *StateStore) Cache() *writecache.WriteCache   { return this.WriteCache }
func (this *StateStore) Import(trans univalue.Univalues) { this.StateCommitter.Import(trans) }
func (this *StateStore) Preload(key []byte) interface{}  { return this.store.Preload(key) }
func (this *StateStore) Clear()                          { this.WriteCache.Clear() }

func (this *StateStore) GetWriters() []intf.AsyncWriter[*univalue.Univalue] {
	return append([]intf.AsyncWriter[*univalue.Univalue]{
		writecache.NewAsyncWriter(this.WriteCache, 0)},
		this.store.GetWriters()...)
}
