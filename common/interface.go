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

package storagecommon

// type Platform interface { // value type
// 	IsSysPath(string) bool
// 	// Eth10Account() string
// }

type Type interface { // value type
	TypeID() uint8
	Equal(any) bool
	Clone() any

	IsNumeric() bool
	IsCommutative() bool
	IsBounded() bool

	Value() any // Get() - read/write count
	Delta() any
	DeltaSign() bool
	CloneDelta() any
	Min() any
	Max() any
	New(any, any, any, any, any) any

	SetValue(any)
	IsDeltaApplied() bool
	ResetDelta()
	SetDelta(any)
	SetDeltaSign(any)
	SetMin(any)
	SetMax(any)

	Get() (any, uint32, uint32) // Value, reads and writes, no deltawrites.
	Set(any, any) (any, uint32, uint32, uint32, error)
	CopyTo(any) (any, uint32, uint32, uint32)
	ApplyDelta([]Type) (Type, int, error)
	IsSelf(any) bool

	MemSize() uint64 // Size in memory
	Size() uint64    // Encoded size
	Encode() []byte
	EncodeToBuffer([]byte) int
	Decode([]byte) any

	StorageEncode(string) []byte
	StorageDecode(string, []byte) any

	Preload(string, any)

	AffliatedDeletes() []string // Keys that are deleted when this key is deleted.

	Hash(func([]byte) []byte) []byte
	Reset()
	Print()
}

type AsyncWriter[T any] interface {
	Import([]T)
	Precommit()
	Commit(uint64)
}

type ReadOnlyStore interface {
	IfExists(string) bool                        // Check if the key exists in the source, which can be a cache or a storage.
	RetriveFromStorage(string, any) (any, error) // Check if the key is in the persistent storage.
	Retrive(string, any) (any, error)
	Preload([]byte) any
}

type Hasher func(Type) []byte
