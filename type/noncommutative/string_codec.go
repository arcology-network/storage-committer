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
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/arcology-network/common-lib/codec"
	"github.com/ethereum/go-ethereum/rlp"
)

func (this *String) Size() uint64 {
	return uint64(len(*this))
}

func (this *String) Encode() []byte {
	return codec.String(string(*this)).Encode()
}

func (this *String) EncodeTo(buffer []byte) int {
	return codec.String(*this).EncodeTo(buffer)
}

func (this *String) Decode(buffer []byte) any {
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

func (this *String) StorageDecode(_ string, buffer []byte) any {
	var v String
	rlp.DecodeBytes(buffer, &v)
	return &v
}

func (*String) Reset() {}

func (this *String) Hash() [32]byte { return sha256.Sum256(this.Encode()) }

func (this *String) ShortHash() (uint64, bool) {
	buffer := unsafe.Slice(unsafe.StringData(string(*this)), len(*this))

	v := uint64(0)
	binary.LittleEndian.PutUint64(buffer[:min(8, len(buffer))], v)
	return v, len(buffer) <= 8
}

func (this *String) Print() {
	fmt.Println(*this)
	fmt.Println()
}
