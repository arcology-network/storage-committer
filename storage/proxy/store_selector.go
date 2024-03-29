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
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
)

type StoreSelector struct{}

func (this *StoreProxy) GetStorage(key string) intf.Datastore {
	if platform.IsEthPath(key) {
		return this.ethDataStore
	}
	return this.ccDataStore
}

// func (this *StoreProxy) FilterLocalByType(vals *[]*univalue.Univalue) ([]*univalue.Univalue, []*univalue.Univalue) {
// 	localTrans := slice.MoveIf(vals, func(i int, v *univalue.Univalue) bool {
// 		return v.TypeID() == commutative.PATH // Move all the path metadata to the local storage
// 	})
// 	return *vals, localTrans
// }

// func (this *StoreProxy) FilterLocalByPath(key *[]string, vals *[]any) ([]string, []any) {
// 	localKeys, localVals := slice.MoveBothIf(key, vals, func(i int, str string, v any) bool {
// 		return strings.Contains(str, "/container")
// 	})
// 	return localKeys, localVals
// }
