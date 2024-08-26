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

package noncommutative

import (
	"bytes"
	"fmt"

	"github.com/arcology-network/common-lib/codec"
	"github.com/ethereum/go-ethereum/rlp"
)

func (this *String) Size() uint32 {
	return uint32(len(*this))
}

func (this *String) Encode() []byte {
	return codec.String(string(*this)).Encode()
}

func (this *String) EncodeToBuffer(buffer []byte) int {
	return codec.String(*this).EncodeToBuffer(buffer)
}

func (this *String) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}

	*this = String(codec.String("").Decode(bytes.Clone(buffer)).(codec.String))
	return this
}

func (this *String) StorageEncode(_ string) []byte {
	buffer, _ := rlp.EncodeToBytes(*this)
	return buffer
}

func (this *String) StorageDecode(_ string, buffer []byte) interface{} {
	var v String
	rlp.DecodeBytes(buffer, &v)
	return &v
}

func (*String) Reset() {}

func (this *String) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

func (this *String) Print() {
	fmt.Println(*this)
	fmt.Println()
}
