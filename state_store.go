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
	writecache "github.com/arcology-network/storage-committer/storage/writecache"
	"github.com/arcology-network/storage-committer/univalue"
	"github.com/cespare/xxhash/v2"
)

// Buffer is simpliest  of indexers. It does not index anything, just stores the transitions.
type StateStore struct {
	*writecache.ShardedWriteCache
	store     intf.ReadOnlyDataStore
	committer *stgcomm.StateCommitter
}

// New creates a new StateCommitter instance.
func NewStateStore(store intf.ReadOnlyDataStore) *StateStore {
	return &StateStore{
		store: store,
		ShardedWriteCache: writecache.NewShardedWriteCache(
			store,
			16,
			1,
			func(k string) uint64 {
				return xxhash.Sum64String(k)
			},
		),
		committer: stgcomm.NewStateCommitter(store),
	}
}

func (this *StateStore) Store() intf.ReadOnlyDataStore      { return this.store }
func (this *StateStore) Committer() *stgcomm.StateCommitter { return this.committer }

// The committer will commit the transactions to different stores registered with the committer.
func (this *StateStore) Precommit(tx []uint32) [32]byte {
	return this.committer.Precommit(tx)
}

func (this *StateStore) Commit(blockNum uint64) *StateStore {
	this.committer.Commit(blockNum)
	return this
}

func (this *StateStore) Import(trans univalue.Univalues) *StateStore {
	this.committer.Import(trans)
	return this
}

func (this *StateStore) Clear() {
	this.ShardedWriteCache.Clear()
}
