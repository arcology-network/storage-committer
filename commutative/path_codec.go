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

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/ethereum/go-ethereum/rlp"

	// performance "github.com/arcology-network/common-lib/mhasher"
	orderedset "github.com/arcology-network/common-lib/container/set"
)

func (this *Path) HeaderSize() uint32 {
	return 3 * codec.UINT32_LEN // number of fields + 1
}

func (this *Path) Size() uint32 {
	return this.HeaderSize() +
		common.IfThenDo1st(this.value != nil, func() uint32 { return this.value.Size() }, 0) +
		common.IfThenDo1st(this.delta != nil, func() uint32 { return this.delta.Size() }, 0)
}

func (this *Path) Encode() []byte {
	buffer := make([]byte, this.Size()) //  no need to send the committed keys
	offset := codec.Encoder{}.FillHeader(buffer,
		[]uint32{
			common.IfThenDo1st(this.value != nil, func() uint32 { return this.value.Size() }, 0),
			common.IfThenDo1st(this.delta != nil, func() uint32 { return this.delta.Size() }, 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:])
	return buffer
}

func (this *Path) EncodeToBuffer(buffer []byte) int {
	offset := common.IfThenDo1st(this.value != nil, func() int { return this.value.EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(this.delta != nil, func() int { return this.delta.EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *Path) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}

	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	return &Path{
		value: orderedset.NewOrderedSet(codec.Strings{}.Decode(fields[0]).(codec.Strings)),
		delta: NewPathDelta([]string{}, []string{}).Decode(fields[1]).(*PathDelta),
	}
}

func (this *Path) Print() {
	// fmt.Println("Keys: ", this.committedKeys)
	fmt.Println("Added: ", this.delta.addDict.Keys())
	fmt.Println("Removed: ", this.delta.delDict.Keys())
	fmt.Println()
}

func (this *Path) StorageEncode(_ bool) []byte {
	buffer, _ := rlp.EncodeToBytes(this.Encode())
	return buffer
}

func (this *Path) StorageDecode(_ bool, buffer []byte) interface{} {
	var decoded []byte
	rlp.DecodeBytes(buffer, &decoded)
	return this.Decode(decoded)
}
