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
	"errors"

	stgcommon "github.com/arcology-network/storage-committer/common"
)

type Any[T any] struct {
	value T
	min   T
	max   T
}

func (this *Any[T]) New(_, delta, _, _, _ any) any { return nil }

func (this *Any[T]) MemSize() uint64                            { return 0 }
func (this *Any[T]) CanApply(v any) bool                        { return false } // If the input has the same type as this, return true
func (this *Any[T]) TypeID() uint8                              { return stgcommon.UNKNOWN }
func (this *Any[T]) CopyTo(v any) (any, uint32, uint32, uint32) { return v, 0, 1, 0 }

// create a new path
func (this *Any[T]) Clone() any           { return nil }
func (this *Any[T]) Equal(other any) bool { return true }

func (this *Any[T]) IsNumeric() bool     { return false }
func (this *Any[T]) IsCommutative() bool { return false }
func (this *Any[T]) IsBounded() bool     { return false }

func (this *Any[T]) Value() any             { return this.value }
func (this *Any[T]) Delta() (any, bool)     { return this.value, true }
func (this *Any[T]) DeltaSign() bool        { return true } // delta sign
func (this *Any[T]) Limits() (any, any)     { return nil, nil }
func (this *Any[T]) SetLimits(min, max any) { this.min = min.(T); this.max = max.(T) }

func (this *Any[T]) CloneDelta() (any, bool) { return nil, true }
func (this *Any[T]) SetValue(v any)          { this.SetDelta(v, true) }
func (this *Any[T]) Preload(_ string, _ any) {}

func (this *Any[T]) IsDeltaApplied() bool       { return true }
func (this *Any[T]) ResetDelta()                {}
func (this *Any[T]) SetDelta(v any, _ bool)     {}
func (this *Any[T]) Get() (any, uint32, uint32) { return this.value, 1, 0 }
func (this *Any[T]) Set(val any, _ any) (any, uint32, uint32, uint32, error) {
	this.value = val.(T)
	return nil, 0, 1, 0, errors.New("Implementation of Set is not defined for Any type")
}

func (this *Any[T]) ApplyDelta(typedVals []stgcommon.Type) (stgcommon.Type, int, error) {
	for _, v := range typedVals {

		if this == nil && v != nil { // New value
			this.value = v.(*Any[T]).value
		}

		if this != nil && v != nil {
			this.Set(v, nil)
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	return this, len(typedVals), nil
}

func (this *Any[T]) Size() uint64                     { return 0 }
func (this *Any[T]) Encode() []byte                   { return nil }
func (this *Any[T]) EncodeToBuffer(buffer []byte) int { return 0 }
func (this *Any[T]) Decode(buffer []byte) any         { return nil }

func (this *Any[T]) StorageEncode(_ string) []byte             { return nil }
func (this *Any[T]) StorageDecode(_ string, buffer []byte) any { return nil }
func (this *Any[T]) Hash(hasher func([]byte) []byte) []byte    { return nil }
func (this *Any[T]) ShortHash() (uint64, bool)                 { return 0, false } // For fast comparison only.
func (this *Any[T]) Print()                                    {}
