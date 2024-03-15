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

package storage

import (
	"strings"

	"github.com/arcology-network/common-lib/exp/slice"
	commutative "github.com/arcology-network/storage-committer/commutative"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
)

type StoreSelector struct{}

func (this *StoreRouter) GetStorage(key string) intf.Datastore {
	if !strings.Contains(key, "/container") {
		return this.ethDataStore
	}
	return this.ccDataStore
}

func (this *StoreRouter) FilterLocalByType(vals *[]*univalue.Univalue) ([]*univalue.Univalue, []*univalue.Univalue) {
	localTrans := slice.MoveIf(vals, func(i int, v *univalue.Univalue) bool {
		return v.TypeID() == commutative.PATH // Move all the path metadata to the local storage
	})
	return *vals, localTrans
}

func (this *StoreRouter) FilterLocalByPath(key *[]string, vals *[]any) ([]string, []any) {
	localKeys, localVals := slice.MoveBothIf(key, vals, func(i int, str string, v any) bool {
		return strings.Contains(str, "/container")
	})
	return localKeys, localVals
}
