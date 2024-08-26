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
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func (this *Uint64) HeaderSize() uint32 {
	return 5 * codec.UINT32_LEN //static size only , no header needed,
}

func (this *Uint64) Size() uint32 {
	return this.HeaderSize() +
		common.IfThen(this.value != 0, uint32(8), 0) +
		common.IfThen(this.delta != 0, uint32(8), 0) +
		common.IfThen(this.min != 0, uint32(8), 0) +
		common.IfThen(this.max != math.MaxUint64, uint32(8), 0)
}

func (this *Uint64) Encode() []byte {
	buffer := make([]byte, this.Size())
	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			common.IfThen(this.value != 0, uint32(8), 0),
			common.IfThen(this.delta != 0, uint32(8), 0),
			common.IfThen(this.min != 0, uint32(8), 0),
			common.IfThen(this.max != math.MaxUint64, uint32(8), 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:])
	return buffer
}

func (this *Uint64) EncodeToBuffer(buffer []byte) int {
	offset := common.IfThenDo1st(this.value != 0, func() int { return codec.Uint64(this.value).EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(this.delta != 0, func() int { return codec.Uint64(this.delta).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(this.min != 0, func() int { return codec.Uint64(this.min).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(this.max != math.MaxUint64, func() int { return codec.Uint64(this.max).EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *Uint64) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)

	this.value = uint64(codec.Uint64(0).Decode(fields[0]).(codec.Uint64))
	this.delta = uint64(codec.Uint64(0).Decode(fields[1]).(codec.Uint64))
	this.min = uint64(codec.Uint64(0).Decode(fields[2]).(codec.Uint64))
	this.max = uint64(codec.Uint64(math.MaxUint64).Decode(fields[3]).(codec.Uint64))
	return this
}

func (this *Uint64) Print() {
	fmt.Println(" Value: ", this.value, "Delta: ", this.delta)
}

func (this *Uint64) StorageEncode(_ string) []byte {
	var buffer []byte
	if this.IsBounded() {
		v := []*big.Int{new(big.Int).SetUint64(this.value), new(big.Int).SetUint64(this.min), new(big.Int).SetUint64(this.max)}
		buffer, _ = rlp.EncodeToBytes(v)
	} else {
		buffer, _ = rlp.EncodeToBytes(this.value)
	}
	return buffer
}

func (*Uint64) StorageDecode(_ string, buffer []byte) interface{} {
	this := NewUnboundedUint64().(*Uint64)

	arr := make([]*big.Int, 3)
	err := rlp.DecodeBytes(buffer, &arr)
	if err != nil {
		var value big.Int
		if err = rlp.DecodeBytes(buffer, &value); err == nil {
			this.value = value.Uint64()
		}
	} else {
		this.value = arr[0].Uint64()
		this.min = arr[1].Uint64()
		this.max = arr[2].Uint64()
	}
	return this
}
