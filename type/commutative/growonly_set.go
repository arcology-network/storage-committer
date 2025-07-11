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
package commutative

import (
	"crypto/sha256"
	"reflect"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommon "github.com/arcology-network/storage-committer/common"
)

type GrowOnlySet[T any] struct {
	value          []T
	sizer          func(T) uint64
	encodeToBuffer func(T, []byte) int
	decorder       func([]byte) T
	equal          func(T, T) bool
}

func NewGrowOnlyByteSet(v []byte) *GrowOnlySet[[]byte] {
	return &GrowOnlySet[[]byte]{
		value:          [][]byte{v}, // Initialize with the provided byte slice
		sizer:          func(v []byte) uint64 { return uint64(len(v)) },
		encodeToBuffer: func(v []byte, bf []byte) int { return copy(bf, v) },
		decorder:       func(v []byte) []byte { return v },
		equal:          func(v0, v1 []byte) bool { return reflect.DeepEqual(v0, v1) },
	}
}

func NewGrowOnlySet[T any](
	sizer func(T) uint64,
	encodeToBuffer func(T, []byte) int,
	decorder func([]byte) T,
	equal func(T, T) bool, args ...T) stgcommon.Type {
	this := &GrowOnlySet[T]{
		value:          make([]T, 0),
		sizer:          sizer,
		encodeToBuffer: encodeToBuffer,
		decorder:       decorder,
		equal:          equal,
	}
	this.value = append(this.value, args...)
	return this
}

func (this *GrowOnlySet[T]) New(_, delta, _, _, _ any) any {
	return &GrowOnlySet[T]{
		value:          []T{}, // Initialize with an empty slice
		sizer:          this.sizer,
		encodeToBuffer: this.encodeToBuffer,
		decorder:       this.decorder,
		equal:          this.equal,
	}
}

func (this *GrowOnlySet[T]) Clone() any {
	if this == nil {
		return nil
	}
	return slice.Clone(this.value) // Clone the slice to avoid modifying the original
}

func (this *GrowOnlySet[T]) Equal(other any) bool {
	if len(this.value) != len(other.(*GrowOnlySet[T]).value) {
		return false
	}

	// Order must be the same, so we can compare directly
	for i := range this.value {
		if !this.equal(this.value[i], other.(*GrowOnlySet[T]).value[i]) {
			return false
		}
	}
	return true
}

func (this *GrowOnlySet[T]) IsNumeric() bool     { return false }
func (this *GrowOnlySet[T]) IsCommutative() bool { return true }
func (this *GrowOnlySet[T]) IsBounded() bool     { return false }

func (this *GrowOnlySet[T]) MemSize() uint64                            { return 0 }
func (this *GrowOnlySet[T]) CanApply(v any) bool                        { return false } // If the input has the same type as this, return true
func (this *GrowOnlySet[T]) TypeID() uint8                              { return GROWONLY_SET }
func (this *GrowOnlySet[T]) CopyTo(v any) (any, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *GrowOnlySet[T]) Value() any         { return this.value }
func (this *GrowOnlySet[T]) Delta() (any, bool) { return this.value, true }
func (this *GrowOnlySet[T]) DeltaSign() bool    { return true } // delta sign
func (this *GrowOnlySet[T]) Limits() (any, any) { return nil, nil }
func (this *GrowOnlySet[T]) SetLimits(_, _ any) {}

func (this *GrowOnlySet[T]) SetValue(v any) { this.SetDelta(v, true) }
func (this *GrowOnlySet[T]) ResetDelta()    { this.value = []T{} }
func (this *GrowOnlySet[T]) SetDelta(v any, _ bool) {
	this.value = v.([]T)
}

func (this *GrowOnlySet[T]) Preload(_ string, _ any) {}
func (this *GrowOnlySet[T]) IsDeltaApplied() bool    { return true }
func (this *GrowOnlySet[T]) CloneDelta() (any, bool) { return this.Clone(), true }

func (this *GrowOnlySet[T]) Get() (any, uint32, uint32) { return this.value, 1, 0 }
func (this *GrowOnlySet[T]) Set(v any, _ any) (any, uint32, uint32, uint32, error) {
	if v == nil {
		this.value = nil
		return this, 0, 1, 0, nil
	}

	vals := v.(*GrowOnlySet[T]).Value().([]T)
	this.value = append(this.value, vals...)
	return this, 0, 1, 0, nil
}

func (this *GrowOnlySet[T]) ApplyDelta(otherSets []stgcommon.Type) (stgcommon.Type, int, error) {
	for _, v := range otherSets {
		if this == nil && v != nil { // New value
			this = &GrowOnlySet[T]{}
			this.value = v.(*GrowOnlySet[T]).value
		}

		if this != nil && v != nil {
			this.Set(v, nil)
		}
	}
	return this, len(otherSets), nil
}

// Size() uint64 // Encoded size
func (this *GrowOnlySet[T]) HeaderSize() uint64 { return uint64(len(this.value)+1) * codec.UINT64_LEN } // Assuming each element is 8 bytes, adjust as necessary

func (this *GrowOnlySet[T]) Size() uint64 {
	return this.HeaderSize() + slice.Accumulate(this.value, uint64(0), func(_ int, v T) uint64 { return this.sizer(v) }) // Add header size
}

func (this *GrowOnlySet[T]) Encode() []byte {
	if this == nil {
		return []byte{} // Return empty byte slice if nil
	}
	buff := make([]byte, this.Size())
	this.EncodeToBuffer(buff)
	return buff
}

func (this *GrowOnlySet[T]) EncodeToBuffer(buffer []byte) int {
	offset := codec.Encoder{}.FillHeader(
		buffer,
		slice.Transform(this.value, func(_ int, v T) uint64 { return this.sizer(v) }),
	)

	for _, v := range this.value {
		offset += this.encodeToBuffer(v, buffer[offset:]) // Encode each element to the buffer
	}
	return offset
}

func (this *GrowOnlySet[T]) Decode(buffer []byte) any {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	this.value = make([]T, len(fields)) // Initialize the value slice with the number of fields
	for i, field := range fields {
		this.value[i] = this.decorder(field) // Decode each element
	}
	return this
}

func (this *GrowOnlySet[T]) StorageEncode(_ string) []byte             { return this.Encode() }
func (this *GrowOnlySet[T]) StorageDecode(_ string, buffer []byte) any { return this.Decode(buffer) }
func (this *GrowOnlySet[T]) Hash() [32]byte {
	buff := make([]byte, slice.Accumulate(this.value, uint64(0), func(_ int, v T) uint64 { return this.sizer(v) }))
	this.EncodeToBuffer(buff)
	return sha256.Sum256(buff) // Use sha256 to hash the encoded value
}

func (this *GrowOnlySet[T]) ShortHash() (uint64, bool) { return 0, false } // For fast comparison only.
func (this *GrowOnlySet[T]) Print()                    {}
