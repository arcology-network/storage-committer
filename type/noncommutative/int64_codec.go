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
	"math/big"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func (this *Int64) Size() uint64 {
	return 8 // 8 bytes
}
func (this *Int64) Encode() []byte {
	return codec.Int64(*this).Encode()
}

func (this *Int64) EncodeTo(buffer []byte) int {
	return codec.Int64(*this).EncodeTo(buffer)
}

func (*Int64) Decode(bytes []byte) any {
	this := Int64(codec.Int64(0).Decode(bytes).(codec.Int64))
	return &this
}

func (this *Int64) Reset() {}

func (this *Int64) Hash() [32]byte            { return sha256.Sum256(this.Encode()) }
func (this *Int64) ShortHash() (uint64, bool) { return uint64(*this) ^ (1 << 63), true }

func (this *Int64) Print() {
	fmt.Println(*this)
	fmt.Println()
}

func (this *Int64) StorageEncode(_ string) []byte {
	buffer, _ := rlp.EncodeToBytes(new(big.Int).SetInt64(int64(*this.Value().(*Int64))))
	return buffer
}

func (this *Int64) StorageDecode(_ string, buffer []byte) any {
	var v big.Int
	rlp.DecodeBytes(buffer, &v)
	return common.New(Int64(v.Uint64()))
}
