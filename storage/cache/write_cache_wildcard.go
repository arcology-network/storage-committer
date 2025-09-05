/*
 *   Copyright (c) 2025 Arcology Network

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

package cache

import (
	"bytes"

	"github.com/arcology-network/storage-committer/type/univalue"
)

// PreloadMatched preloads the paths that match the wildcard delete path that are about to be deleted by the
// the current write operation.
func (this *WriteCache) MatchWildcard(path string, T any) (bool, *univalue.Univalue) {
	for _, wildcardPath := range this.committedDel {
		if len(path) < len(wildcardPath.Second) {
			continue
		}

		if bytes.Equal([]byte(path[:len(wildcardPath.Second)]), []byte(wildcardPath.Second)) {
			univ := this.LoadFromCommitted(0, path, T) // Preload the path from the backend
			univ.SetValue(nil)                         // To indicate t the path has been deleted by the wildcard
			univ.IncrementWrites(1)
			return true, univ
		}
	}
	return false, nil
}

// WildcardsToUnivalue converts wildcard paths to Univalue for exporting.
func (this *WriteCache) WildcardsToUnivalue() []*univalue.Univalue {
	univs := make([]*univalue.Univalue, 0)
	for _, wildcardPath := range this.committedDel {
		newV := univalue.NewUnivalue(wildcardPath.First, wildcardPath.Second+"*", 0, 1, 0, nil, nil)
		newV.SetPreexist(true) // Mark as pre-existing, so it pass through the filter.
		univs = append(univs, newV)
	}
	return univs
}
