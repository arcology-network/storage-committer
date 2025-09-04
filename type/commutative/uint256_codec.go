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
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func (this *U256) HeaderSize() uint64 {
	return 5 // Total number of fields + offsets of these fields
}

func (this *U256) Size() uint64 {
	return this.HeaderSize() +
		common.IfThen(this.value.Eq(&U256_ZERO), 0, uint64(32)) + // Values
		common.IfThen(this.delta.Eq(&U256_ZERO), 0, uint64(32)) + // delta
		common.IfThen(this.deltaPositive, 0, uint64(1)) + // delta sign
		common.IfThen(this.min.Eq(&U256_ZERO), 0, uint64(32)) + // Min
		common.IfThen(this.max.Eq(&U256_MAX), 0, uint64(32)) // Max
}

func (this *U256) Encode() []byte {
	buffer := make([]byte, this.Size())
	buffer[0] = common.IfThen(this.value.Eq(&U256_ZERO), 0, uint8(32))
	buffer[1] = common.IfThen(this.delta.Eq(&U256_ZERO), 0, uint8(32))
	buffer[2] = common.IfThen(this.deltaPositive, 0, uint8(1)) //only is the delta != 0
	buffer[3] = common.IfThen(this.min.Eq(&U256_ZERO), 0, uint8(32))
	buffer[4] = common.IfThen(this.max.Eq(&U256_MAX), 0, uint8(32))

	this.EncodeTo(buffer[5:])
	return buffer
}

func (this *U256) EncodeTo(buffer []byte) int {
	offset := common.IfThenDo1st(!this.value.Eq(&U256_ZERO), func() int { return codec.Uint64s(this.value[:]).EncodeTo(buffer) }, 0)
	offset += common.IfThenDo1st(!this.delta.Eq(&U256_ZERO), func() int { return codec.Uint64s(this.delta[:]).EncodeTo(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(!this.deltaPositive, func() int { return codec.Bool(this.deltaPositive).EncodeTo(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(!this.min.Eq(&U256_ZERO), func() int { return codec.Uint64s(this.min[:]).EncodeTo(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(!this.max.Eq(&U256_MAX), func() int { return codec.Uint64s(this.max[:]).EncodeTo(buffer[offset:]) }, 0)
	return offset
}

func (this *U256) Decode(buffer []byte) any {
	if len(buffer) == 0 {
		return this
	}
	this = NewUnboundedU256().(*U256)

	offset := 5
	if buffer[0] > 0 {
		copy(this.value[:], codec.Uint64s{}.Decode(buffer[offset:]).(codec.Uint64s))
		offset += int(buffer[0])
	}

	if buffer[1] > 0 {
		copy(this.delta[:], codec.Uint64s{}.Decode(buffer[offset:]).(codec.Uint64s))
		offset += int(buffer[1])
	}

	if buffer[2] > 0 {
		this.deltaPositive = bool(codec.Bool(true).Decode(buffer[offset:]).(codec.Bool))
		offset += int(buffer[2])
	}

	if buffer[3] > 0 {
		copy(this.min[:], codec.Uint64s{}.Decode(buffer[offset:]).(codec.Uint64s))
		offset += int(buffer[3])
	}

	if buffer[4] > 0 {
		copy(this.max[:], codec.Uint64s{}.Decode(buffer[offset:]).(codec.Uint64s))
	}

	return this
}

func (this *U256) StorageEncode(_ string) []byte {
	var buffer []byte
	if this.HasLimits() {
		buffer, _ = rlp.EncodeToBytes([]any{this.value, this.min, this.max})
	} else {
		buffer, _ = rlp.EncodeToBytes(this.value.ToBig())
	}
	return buffer
}

func (*U256) StorageDecode(_ string, buffer []byte) any {
	this := NewUnboundedU256().(*U256)

	var arr []any
	err := rlp.DecodeBytes(buffer, &arr)
	if err != nil {
		var v2 big.Int
		if err = rlp.DecodeBytes(buffer, &v2); err == nil {
			this.value.SetFromBig(&v2)
		}
	} else {
		this.value.SetBytes(arr[0].([]byte))
		this.min.SetBytes(arr[1].([]byte))
		this.max.SetBytes(arr[2].([]byte))
	}
	return this
}
