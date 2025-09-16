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
	softdeltaset "github.com/arcology-network/common-lib/exp/softdeltaset"
	"github.com/ethereum/go-ethereum/rlp"
	// performance "github.com/arcology-network/common-lib/mhasher"
)

func (this *Path) HeaderSize() uint64 {
	return 5 * codec.UINT64_LEN // number of fields + 1
}

func (this *Path) Size() uint64 {
	return this.HeaderSize() +
		8 + // TotalSize
		1 + // isBlockBound
		uint64(this.DeltaSet.Size()) +
		1 // 1 byte for element type ID
}

func (this *Path) Encode() []byte {
	buffer := make([]byte, this.Size()) //  no need to send the isCommitted keys
	this.EncodeTo(buffer)
	return buffer
}

func (this *Path) EncodeTo(buffer []byte) int {
	offset := codec.Encoder{}.FillHeader(buffer,
		[]uint64{
			8,
			1,
			uint64(this.DeltaSet.Size()),
			1,
		},
	)

	offset += codec.Uint64(this.TotalSize).EncodeTo(buffer[offset:])
	offset += codec.Bool(this.isBlockBound).EncodeTo(buffer[offset:])
	this.DeltaSet.EncodeTo(buffer[offset:])
	offset += 1

	return offset
}

func (*Path) Decode(buffer []byte) any {
	if len(buffer) == 0 {
		return NewPath()
	}

	path := NewPath().(*Path)
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	path.TotalSize = uint64(codec.Uint64(0).Decode(fields[0]).(codec.Uint64))
	path.isBlockBound = bool(codec.Bool(false).Decode(fields[1]).(codec.Bool))
	path.DeltaSet = path.DeltaSet.Decode(fields[2]).(*softdeltaset.DeltaSet[string])
	path.ElemType = uint8(fields[3][0])
	return path
}

func (this *Path) Print() {
	fmt.Println("TotalSize: ", this.TotalSize)
	fmt.Println("isBlockBound: ", this.isBlockBound)
	fmt.Println("Committed: ", codec.Strings(this.DeltaSet.Committed().Elements()).ToHex())
	fmt.Println("Staged Added: ", codec.Strings(this.DeltaSet.Added().Elements()).ToHex())
	fmt.Println("Staged Removed: ", codec.Strings(this.DeltaSet.Removed().Elements()).ToHex())
	fmt.Println("Type: ", this.TypeID())
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
