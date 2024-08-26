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

package univalue

import (
	"bytes"
	"reflect"

	codec "github.com/arcology-network/common-lib/codec"
)

func (this *Property) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Property) HeaderSize() uint32 {
	return uint32(10 * codec.UINT32_LEN)
}

func (this *Property) Size() uint32 {
	return this.HeaderSize() + // uint32(9*codec.UINT32_LEN) +
		uint32(1) + // codec.Uint8(this.vType).Size() +
		uint32(4) + // codec.Uint32(uint32(this.tx)).Size() +
		uint32(len(*this.path)) + // codec.String(*this.path).Size() +
		uint32(4) + // codec.Uint32(this.reads).Size() +
		uint32(4) + // codec.Uint32(this.writes).Size() +
		uint32(4) + // codec.Uint32(this.deltaWrites).Size() +
		uint32(1) + //+  codec.Bool(this.preexists).Size() +
		uint32(1) + //+  codec.Bool(this.persistent).Size() +
		uint32(len(this.msg))
}

func (this *Property) FillHeader(buffer []byte) int {
	return codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			uint32(codec.Uint8(this.vType).Size()),
			codec.Uint32(this.tx).Size(),
			codec.String(*this.path).Size(),
			codec.Uint32(this.reads).Size(),
			codec.Uint32(this.writes).Size(),
			codec.Uint32(this.deltaWrites).Size(),
			codec.Bool(this.preexists).Size(),
			codec.Bool(this.persistent).Size(),
			codec.String(this.msg).Size(),
		},
	)
}

func (this *Property) EncodeToBuffer(buffer []byte) int {
	offset := this.FillHeader(buffer)
	offset += codec.Uint8(this.vType).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.tx).EncodeToBuffer(buffer[offset:])
	offset += codec.String(*this.path).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.reads).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.writes).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.deltaWrites).EncodeToBuffer(buffer[offset:])
	offset += codec.Bool(this.preexists).EncodeToBuffer(buffer[offset:])
	offset += codec.Bool(this.persistent).EncodeToBuffer(buffer[offset:])
	offset += codec.String(this.msg).EncodeToBuffer(buffer[offset:])

	return offset
}

func (this *Property) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	if len(fields) == 1 {
		return this
	}

	this.vType = uint8(reflect.Kind(codec.Uint8(1).Decode(fields[0]).(codec.Uint8)))
	this.tx = uint32(codec.Uint32(0).Decode(fields[1]).(codec.Uint32))
	key := string(codec.String("").Decode(bytes.Clone(fields[2])).(codec.String))
	this.path = &key
	this.reads = uint32(codec.Uint32(1).Decode(fields[3]).(codec.Uint32))
	this.writes = uint32(codec.Uint32(1).Decode(fields[4]).(codec.Uint32))
	this.deltaWrites = uint32(codec.Uint32(1).Decode(fields[5]).(codec.Uint32))
	this.preexists = bool(codec.Bool(false).Decode(fields[6]).(codec.Bool))
	this.persistent = bool(codec.Bool(true).Decode(fields[7]).(codec.Bool))
	this.msg = string(codec.String("").Decode(bytes.Clone(fields[8])).(codec.String))
	return this
}

func (this *Property) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}

func (this *Property) GobDecode(data []byte) error {
	v := this.Decode(data).(*Property)
	this.vType = v.vType
	this.path = v.path
	this.preexists = v.preexists
	this.persistent = v.persistent
	this.tx = v.tx
	this.reads = v.reads
	this.writes = v.writes
	this.deltaWrites = v.deltaWrites
	this.msg = v.msg
	return nil
}
