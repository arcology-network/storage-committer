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
	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/interfaces"
	writecache "github.com/arcology-network/storage-committer/storage/writecache"
	"github.com/arcology-network/storage-committer/univalue"
)

// WriteCache is a wrapper around writecache.WriteCache with some extra methods provided
// by the intf.Datastore interface to work with the storage-committer.
type WriteCache struct {
	*writecache.WriteCache // Provide Readonly interface
}

func NewWriteCache(store intf.Datastore) *WriteCache {
	return &WriteCache{writecache.NewWriteCache(store, 10, 10)}
}

func (this *WriteCache) Precommit(args ...interface{}) [32]byte {
	univs := []*univalue.Univalue(*(args[0].(*Buffer)))
	slice.RemoveIf(&univs, func(i int, v *univalue.Univalue) bool { return v.GetPath() == nil })
	return this.WriteCache.Precommit(univs)
}

func (this *WriteCache) Commit(placeHolder uint64) error { return nil }
