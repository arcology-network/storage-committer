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
)

func (this *Path) HeaderSize() uint64 {
	return 5 * codec.UINT64_LEN // number of fields + 1
}

func (this *Path) Size() uint64 {
	committedEmpty := this.DeltaSet.Committed() != nil
	appendedEmpty := this.DeltaSet.Added() != nil
	removedEmpty := this.DeltaSet.Removed() != nil

	return this.HeaderSize() +
		common.IfThenDo1st(committedEmpty, func() uint64 { return codec.Strings(this.DeltaSet.Committed().Elements()).Size() }, 0) +
		common.IfThenDo1st(appendedEmpty, func() uint64 { return codec.Strings(this.DeltaSet.Added().Elements()).Size() }, 0) +
		common.IfThenDo1st(removedEmpty, func() uint64 { return codec.Strings(this.DeltaSet.Removed().Elements()).Size() }, 0) +
		1 // 1 byte for typeID
}

func (this *Path) Encode() []byte {
	buffer := make([]byte, this.Size()) //  no need to send the committed keys
	offset := codec.Encoder{}.FillHeader(buffer,
		[]uint64{
			common.IfThenDo1st(this.DeltaSet.Committed() != nil, func() uint64 { return codec.Strings(this.DeltaSet.Committed().Elements()).Size() }, 0),
			common.IfThenDo1st(this.DeltaSet.Added() != nil, func() uint64 { return codec.Strings(this.DeltaSet.Added().Elements()).Size() }, 0),
			common.IfThenDo1st(this.DeltaSet.Removed() != nil, func() uint64 { return codec.Strings(this.DeltaSet.Removed().Elements()).Size() }, 0),
			1,
		},
	)
	this.EncodeToBuffer(buffer[offset:])
	return buffer
}

func (this *Path) EncodeToBuffer(buffer []byte) int {
	offset := common.IfThenDo1st(this.DeltaSet.Committed() != nil, func() int {
		return (codec.Strings(this.DeltaSet.Committed().Elements()).EncodeToBuffer(buffer))
	}, 0)

	offset += common.IfThenDo1st(this.DeltaSet.Added() != nil, func() int {
		return codec.Strings(this.DeltaSet.Added().Elements()).EncodeToBuffer(buffer[offset:])
	}, 0)

	offset += common.IfThenDo1st(this.DeltaSet.Removed() != nil, func() int {
		return codec.Strings(this.DeltaSet.Removed().Elements()).EncodeToBuffer(buffer[offset:])
	}, 0)

	buffer[offset] = this.Type
	offset += 1

	return offset
}

func (*Path) Decode(buffer []byte) any {
	if len(buffer) == 0 {
		return NewPath()
	}

	path := NewPath().(*Path)
	path.DeltaSet.SetNilVal("")

	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	path.DeltaSet.InsertCommitted(codec.Strings{}.Decode(fields[0]).(codec.Strings))
	path.DeltaSet.InsertAdded(codec.Strings{}.Decode(fields[1]).(codec.Strings))
	path.DeltaSet.InsertRemoved(codec.Strings{}.Decode(fields[2]).(codec.Strings))
	path.Type = uint8(fields[3][0])
	return path
}

func (this *Path) Print() {
	fmt.Println("Committed: ", codec.Strings(this.DeltaSet.Committed().Elements()).ToHex())
	fmt.Println("Updated  ", codec.Strings(this.DeltaSet.Added().Elements()).ToHex())
	fmt.Println("Removed: ", codec.Strings(this.DeltaSet.Removed().Elements()).ToHex())
	fmt.Println("Type: ", codec.Strings(this.DeltaSet.Removed().Elements()).ToHex())
	fmt.Println()
}

func (this *Path) StorageEncode(_ string) []byte {
	buffer, _ := rlp.EncodeToBytes(this.Encode())
	return buffer
}

func (this *Path) StorageDecode(_ string, buffer []byte) any {
	var decoded []byte
	rlp.DecodeBytes(buffer, &decoded)
	return this.Decode(decoded)
}
