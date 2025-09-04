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
	"crypto/sha256"
	"fmt"

	"github.com/arcology-network/common-lib/codec"
)

func (this *Uint32) Size() uint64 {
	return 8 // 8 bytes
}
func (this *Uint32) Encode() []byte {
	return codec.Uint32(*this).Encode()
}

func (this *Uint32) EncodeTo(buffer []byte) int {
	return codec.Uint32(*this).EncodeTo(buffer)
}

func (*Uint32) Decode(bytes []byte) any {
	this := Uint32(codec.Uint32(0).Decode(bytes).(codec.Uint32))
	return &this
}

func (this *Uint32) Reset()                        {}
func (this *Uint32) Hash() [32]byte                { return sha256.Sum256(this.Encode()) }
func (this *Uint32) ShortHash() (uint64, bool)     { return uint64(*this) ^ (1 << 63), true }
func (this *Uint32) StorageEncode(_ string) []byte { return codec.Uint32(*this).Encode() }
func (this *Uint32) StorageDecode(_ string, buffer []byte) any {
	return new(Uint32).Decode(buffer).(*Uint32)
}

func (this *Uint32) Print() {
	fmt.Println(*this)
	fmt.Println()
}
