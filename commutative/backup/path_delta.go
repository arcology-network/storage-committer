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
	orderedset "github.com/arcology-network/common-lib/container/set"
)

type PathDelta struct {
	addDict *orderedset.OrderedSet
	delDict *orderedset.OrderedSet
}

func NewPathDelta(add []string, del []string) *PathDelta {
	return &PathDelta{
		orderedset.NewOrderedSet(add),
		orderedset.NewOrderedSet(del),
	}
}

func (this *PathDelta) IsEmpty() bool {
	return this.addDict.Length() == 0 && this.delDict.Length() == 0
}

func (this *PathDelta) Clone() interface{} {
	if this == nil {
		return this
	}
	return &PathDelta{
		this.addDict.Clone().(*orderedset.OrderedSet),
		this.delDict.Clone().(*orderedset.OrderedSet),
	}
}

func (this *PathDelta) Equal(other *PathDelta) bool {
	return this.addDict.Equal(other.addDict) &&
		this.delDict.Equal(other.delDict)
}

func (this *PathDelta) Add(subkey string, value interface{}, preexists bool) bool {
	return this.addKey(subkey, value, preexists) ||
		this.delKeys(subkey, value, preexists)
}

func (this *PathDelta) addKey(subkey string, value interface{}, preexists bool) bool {
	if this.delDict.Exists(subkey) { // Adding back a preexisting entry
		this.delDict.Delete(subkey) // Cancel out each other
		return true
	}

	if !preexists && value != nil {
		this.addDict.Insert(subkey) // won't be duplicate
		return true
	}
	return false
}

// Only the preexisting keys are in this buffer, or they will be cancel each other
func (this *PathDelta) delKeys(subkey string, value interface{}, preexists bool) bool {
	if value != nil {
		return false
	}

	if preexists { // Preexists and value == nil
		this.delDict.Insert(subkey)
		return true
	}
	return this.addDict.Delete(subkey) // Leave out the entry if it is in the added buffer
}

func (this *PathDelta) Added() []string   { return this.addDict.Keys() }
func (this *PathDelta) Removed() []string { return this.delDict.Keys() }
func (this *PathDelta) Touched() bool     { return len(this.Added()) > 0 || len(this.Removed()) > 0 }
