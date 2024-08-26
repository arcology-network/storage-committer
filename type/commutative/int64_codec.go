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
	"math"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func (this *Int64) HeaderSize() uint32 {
	return 5 * codec.UINT32_LEN //static size only , no header needed,
}

func (this *Int64) Size() uint32 {
	return this.HeaderSize() +
		common.IfThen(this.value != 0, uint32(8), 0) +
		common.IfThen(this.delta != 0, uint32(8), 0) +
		common.IfThen(this.min != math.MinInt64, uint32(8), 0) +
		common.IfThen(this.max != math.MaxInt64, uint32(8), 0)
}

func (this *Int64) Encode() []byte {
	buffer := make([]byte, this.Size())
	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			common.IfThen(this.value != 0, uint32(8), 0),
			common.IfThen(this.delta != 0, uint32(8), 0),
			common.IfThen(this.min != math.MinInt64, uint32(8), 0),
			common.IfThen(this.max != math.MaxInt64, uint32(8), 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:])
	return buffer
}

func (this *Int64) EncodeToBuffer(buffer []byte) int {
	offset := common.IfThenDo1st(this.value != 0, func() int { return codec.Int64(this.value).EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(this.delta != 0, func() int { return codec.Int64(this.delta).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(this.min != math.MinInt64, func() int { return codec.Int64(this.min).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(this.max != math.MaxInt64, func() int { return codec.Int64(this.max).EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *Int64) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}

	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)

	this.value = int64(codec.Int64(0).Decode(fields[0]).(codec.Int64))
	this.delta = int64(codec.Int64(0).Decode(fields[1]).(codec.Int64))
	this.min = int64(codec.Int64(math.MinInt64).Decode(fields[2]).(codec.Int64))
	this.max = int64(codec.Int64(math.MaxInt64).Decode(fields[3]).(codec.Int64))
	return this
}

func (this *Int64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}

func (this *Int64) StorageEncode(_ string) []byte {
	var buffer []byte
	if this.IsBounded() {
		buffer, _ = rlp.EncodeToBytes([]interface{}{this.value, this.min, this.max})
	} else {
		buffer, _ = rlp.EncodeToBytes(this.value)
	}
	return buffer
}

func (*Int64) StorageDecode(_ string, buffer []byte) interface{} {
	var this *Int64
	var arr []interface{}
	err := rlp.DecodeBytes(buffer, &arr)
	if err != nil {
		var value int64
		if err = rlp.DecodeBytes(buffer, &value); err == nil {
			this.value = value
		}
	} else {
		this.value = arr[0].(int64)
		this.min = arr[1].(int64)
		this.max = arr[2].(int64)
	}
	return this
}
