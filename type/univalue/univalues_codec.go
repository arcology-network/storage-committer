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
	"fmt"
	"sort"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
)

func (this Univalues) Size() int {
	size := (len(this) + 1) * codec.UINT32_LEN
	for _, v := range this {
		size += int(v.Size())
	}
	return size
}

func (this Univalues) Sizes() []int {
	sizes := make([]int, len(this))
	for i, v := range this {
		sizes[i] = common.IfThenDo1st(v != nil, func() int { return int(v.Size()) }, 0)
	}
	return sizes
}

func (this Univalues) Encode(selector ...interface{}) []byte {
	lengths := make([]uint32, len(this))
	if len(lengths) == 0 {
		return []byte{}
	}

	slice.ParallelForeach(this, 6, func(i int, _ **Univalue) {
		if this[i] != nil {
			lengths[i] = this[i].Size()
		}
	})

	offsets := make([]uint32, len(this)+1)
	for i := 0; i < len(lengths); i++ {
		offsets[i+1] = offsets[i] + lengths[i]
	}

	headerLen := uint32((len(this) + 1) * codec.UINT32_LEN)
	buffer := make([]byte, headerLen+offsets[len(offsets)-1])
	codec.Uint32(len(this)).EncodeToBuffer(buffer)

	slice.ParallelForeach(this, 6, func(i int, _ **Univalue) {
		codec.Uint32(offsets[i]).EncodeToBuffer(buffer[(i+1)*codec.UINT32_LEN:])
		this[i].EncodeToBuffer(buffer[headerLen+offsets[i]:])
	})
	return buffer
}

func (Univalues) Decode(bytes []byte) interface{} {
	if len(bytes) == 0 {
		return Univalues{}
	}

	buffers := [][]byte(codec.Byteset{}.Decode(bytes).(codec.Byteset))
	univalues := make([]*Univalue, len(buffers))

	slice.ParallelForeach(buffers, 6, func(i int, _ *[]byte) {
		v := (&Univalue{}).Decode(buffers[i])
		univalues[i] = v.(*Univalue)
	})
	return Univalues(univalues)
}

func (Univalues) DecodeWithMempool(bytes []byte, get func() *Univalue, put func(interface{})) interface{} {
	if len(bytes) == 0 {
		return nil
	}

	buffers := [][]byte(codec.Byteset{}.Decode(bytes).(codec.Byteset))
	univalues := make([]*Univalue, len(buffers))

	slice.ParallelForeach(buffers, 6, func(i int, _ *[]byte) {
		v := get()
		v.reclaimFunc = put
		univalues[i] = v.Decode(buffers[i]).(*Univalue)
	})
	return Univalues(univalues)
}

// func (Univalues) DecodeV2(bytesset [][]byte, get func() interface{}, put func(interface{})) Univalues {
// 	univalues := make([]*Univalue, len(bytesset))
// 	for i := range bytesset {
// 		v := get().(*Univalue)
// 		v.reclaimFunc = put
// 		v.Decode(bytesset[i])
// 		univalues[i] = v
// 	}
// 	return Univalues(univalues)
// }

func (this Univalues) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}

func (this *Univalues) GobDecode(data []byte) error {
	v := this.Decode(data)
	*this = v.(Univalues)
	return nil
}

// Print the univalues if the satisfied the existing condition
func (this Univalues) Print(condition ...func(v *Univalue) bool) {
	sorted := slice.Clone(this)
	sort.Slice(sorted, func(i, j int) bool {
		return (*sorted[i].GetPath()) < (*sorted[j].GetPath())
	})

	for i, v := range sorted {
		if len(condition) > 0 && !condition[0](v) {
			continue
		}

		fmt.Print(i, ": ")
		v.Print()
	}
	fmt.Println(" --------------------  ")
}
