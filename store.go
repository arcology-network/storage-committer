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

package storagecommitter

import (
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	writecache "github.com/arcology-network/storage-committer/storage/writecache"
)

type StateStore struct {
	*StateCommitter                       // Provide committable interface
	*writecache.WriteCache                // Provide Readonly interface
	store                  intf.Datastore // Backend storage for the write cache and committer.
}

func NewStateStore(store intf.Datastore, platform *platform.Platform) *StateStore {
	return &StateStore{
		StateCommitter: NewStorageCommitter(store),
		WriteCache:     writecache.NewWriteCache(store, 1, 1, platform),
		store:          store,
	}
}
