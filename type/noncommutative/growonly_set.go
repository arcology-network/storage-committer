/*
*   Copyright (c) 2025 Arcology Network

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
	"github.com/arcology-network/common-lib/codec"
	stgcommon "github.com/arcology-network/storage-committer/common"
)

type GrowonlySet struct {
	Any[[][]byte]
}

func NewGrowonlySet() stgcommon.Type {
	return &GrowonlySet{
		Any: Any[[][]byte]{
			value: make([][]byte, 0),
		},
	}
}

func (this *GrowonlySet) Clone() any              { return codec.Byteset(this.value).Clone() }
func (this *GrowonlySet) CloneDelta() (any, bool) { return this.Clone(), true }
func (this *GrowonlySet) ResetDelta()             { this.value = [][]byte{} }
func (this *GrowonlySet) SetDelta(v any, _ bool)  { this.value = v.([][]byte) }

func (this *GrowonlySet) Set(v any, _ any) (any, uint32, uint32, uint32, error) {
	if v == nil {
		this.value = nil
		return this, 0, 1, 0, nil
	}

	this.value = append(this.value, v.([]byte))
	return this, 0, 1, 0, nil
}

// Size() uint64 // Encoded size
func (this *GrowonlySet) Size() uint64   { return codec.Byteset(this.value).Size() }
func (this *GrowonlySet) Encode() []byte { return codec.Byteset(this.value).Encode() }
func (this *GrowonlySet) EncodeToBuffer(buffer []byte) int {
	return codec.Byteset(this.value).EncodeToBuffer(buffer)
}

func (this *GrowonlySet) Decode(buffer []byte) any {
	set := GrowonlySet{}
	set.value = codec.Byteset{}.Decode(buffer).(codec.Byteset)
	return &set
}

func (this *GrowonlySet) StorageEncode(_ string) []byte             { return this.Encode() }
func (this *GrowonlySet) StorageDecode(_ string, buffer []byte) any { return this.Decode(buffer) }
func (this *GrowonlySet) Hash(hasher func([]byte) []byte) []byte {
	return codec.Byteset(this.value).Hash(hasher)
}

func (this *GrowonlySet) ShortHash() (uint64, bool) { return 0, false } // For fast comparison only.
func (this *GrowonlySet) Print()                    { codec.Byteset(this.value).Print() }
