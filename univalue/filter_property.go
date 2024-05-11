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

package univalue

import (
	"strings"

	common "github.com/arcology-network/storage-committer/common"
)

type PropertyFilter struct {
	*Univalue
	Err error
}

func (this PropertyFilter) From(v *Univalue) *Univalue {
	path := *v.GetPath()
	if len(path) < common.ETH10_ACCOUNT_FULL_LENGTH {
		return nil
	}

	if path = path[:common.ETH10_ACCOUNT_FULL_LENGTH]; strings.Contains(path, "/func/") {
		return v
	}
	return nil
}
