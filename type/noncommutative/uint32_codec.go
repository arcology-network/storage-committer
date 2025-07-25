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

func (this *Uint64) Size() uint64 {
	return 8 // 8 bytes
}
func (this *Uint64) Encode() []byte {
	return codec.Uint64(*this).Encode()
}

func (this *Uint64) EncodeTo(buffer []byte) int {
	return codec.Uint64(*this).EncodeTo(buffer)
}

func (*Uint64) Decode(bytes []byte) any {
	this := Uint64(codec.Uint64(0).Decode(bytes).(codec.Uint64))
	return &this
}

func (this *Uint64) Reset()                        {}
func (this *Uint64) Hash() [32]byte                { return sha256.Sum256(this.Encode()) }
func (this *Uint64) ShortHash() (uint64, bool)     { return uint64(*this) ^ (1 << 63), true }
func (this *Uint64) StorageEncode(_ string) []byte { return codec.Uint64(*this).Encode() }
func (this *Uint64) StorageDecode(_ string, buffer []byte) any {
	return new(Uint64).Decode(buffer).(*Uint64)
}

func (this *Uint64) Print() {
	fmt.Println(*this)
	fmt.Println()
}
