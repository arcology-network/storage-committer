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
	"sort"
	"strings"

	"github.com/arcology-network/common-lib/exp/associative"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/storage-committer/type/commutative"
	"github.com/arcology-network/storage-committer/type/univalue"
)

// PreloadMatched preloads the paths that match the wildcard delete path that are about to be deleted by the
// the current write operation.
func (this *WriteCache) MatchWildcard(path string, T any) (bool, *univalue.Univalue) {
	for _, wildcardPath := range this.wildcardDel {
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
	for _, wildcardPath := range this.wildcardDel {
		newV := univalue.NewUnivalue(wildcardPath.First, wildcardPath.Second+"*", 0, 1, 0, nil, nil)
		newV.SetPreexist(true) // Mark as pre-existing, so it pass through the filter.
		univs = append(univs, newV)
	}
	return univs
}

func (this *WriteCache) HandleWildcard(tx uint64, path string, newVal any, args ...any) (bool, uint64) {
	if flag, cleanPath := univalue.IsWildcard(path); flag { // Check if the path is a wildcard path
		this.wildcardDel = append(this.wildcardDel, &associative.Pair[uint64, string]{First: 0, Second: cleanPath})

		sort.SliceStable(this.wildcardDel, func(i, j int) bool {
			return this.wildcardDel[i].Second < this.wildcardDel[j].Second
		})

		slice.UniqueIf(this.wildcardDel, func(v0, v1 *associative.Pair[uint64, string]) bool {
			return v0.Second == v1.Second
		})

		for k, v := range this.kvDict {
			if strings.HasPrefix(k, cleanPath) { // Remove all the paths that match the wildcard path
				if k == cleanPath && this.platform.IsSysPath(k) { // We cannot delete the system paths
					continue
				}

				v.SetValue(nil) // Set the value to nil to indicate the path has been deleted
				v.SetSubstituted(true)
				v.IncrementWrites(1) // Increment the write count
			}
		}

		pathMeta, _, _ := this.FindForRead(tx, cleanPath, newVal, nil) // Read the clean path to ensure it exists in the cache
		return true, pathMeta.(*commutative.Path).TotalSize
	}
	return false, 0 // Return true to indicate that the path is valid
}
