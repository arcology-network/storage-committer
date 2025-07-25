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

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/ethereum/go-ethereum/rlp"
)

func (this *Bigint) Size() uint64 {
	v := codec.Bigint(*this)
	return v.Size()
}

func (this *Bigint) Encode() []byte {
	v := codec.Bigint(*this)
	return v.Encode()
}

func (this *Bigint) EncodeTo(buffer []byte) int {
	v := codec.Bigint(*this)
	return v.EncodeTo(buffer)
}

func (this *Bigint) Decode(buffer []byte) any {
	if len(buffer) == 0 {
		return this
	}
	this = (*Bigint)((&codec.Bigint{}).Decode(buffer).(*codec.Bigint))
	return this
}

// func (this *Bigint) Encode() []byte {
// 	return this.Encode()
// }

// func (this *Bigint) DecodeCompact(bytes []byte) any {
// 	return this.Decode(bytes)
// }

func (this *Bigint) StorageEncode(_ string) []byte {
	buffer, _ := rlp.EncodeToBytes((*big.Int)(this))
	return buffer
}

func (this *Bigint) StorageDecode(_ string, buffer []byte) any {
	rlp.DecodeBytes(buffer, this)
	return this
}

func (this *Bigint) Reset() {}

func (this *Bigint) Hash() [32]byte { return sha256.Sum256(this.Encode()) }

func (this *Bigint) ShortHash() (uint64, bool) {
	v := big.Int(*this)
	if v.IsUint64() {
		return v.Uint64(), true
	}
	return 0, false
}

func (this *Bigint) Print() {
	fmt.Println(*this)
	fmt.Println()
}
