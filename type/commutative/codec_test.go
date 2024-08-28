/*
 *   Copyright (c) 2023 Arcology Network

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

package commutative

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/holiman/uint256"
)

func TestCommutativeCodec(t *testing.T) {
	/* Noncommutative Path Test*/
	v := NewBoundedU256(uint256.NewInt(1), uint256.NewInt(400))
	v.SetValue(*uint256.NewInt(37))

	buffer := v.StorageEncode("")
	output := (&U256{}).StorageDecode("", buffer)

	if !reflect.DeepEqual(v, output) {
		fmt.Println("Error: Missmatched")
	}

	v = NewBoundedUint64(uint64(1), uint64(400))
	v.SetValue(uint64(37))

	buffer = v.StorageEncode("")
	output = (&Uint64{}).StorageDecode("", buffer)

	if !reflect.DeepEqual(v, output) {
		fmt.Println("Error: Missmatched")
	}
}
