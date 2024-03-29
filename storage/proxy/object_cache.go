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
	cache "github.com/arcology-network/common-lib/cache"
	intf "github.com/arcology-network/storage-committer/interfaces"
)

// ReadCache is a read only cache that is used to store the read values from the storage.
// The cache updates itself when the update is called. The implementation isn't thread safe.
// So, it's the caller's responsibility to ensure that the cache is only accessed by one thread updating it.
// Each entry in the cache holds two values, the first value is the old value, and the second value is the new value.
// The new value will be set to the old value when the Finalize function is called.
type ObjectCache cache.ReadCache[string, intf.Type]

func (this *ObjectCache) Commit(placeholder uint64)     {}
func (this *ObjectCache) Precommit(args ...interface{}) {}
