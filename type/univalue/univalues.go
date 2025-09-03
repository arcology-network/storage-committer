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
	"crypto/sha256"
	"sort"
	"strings"

	"github.com/arcology-network/common-lib/exp/slice"
	stgcommon "github.com/arcology-network/storage-committer/common"
)

type Univalues []*Univalue

func (this Univalues) To(filter interface{}) Univalues {
	fun := filter.(interface{ From(*Univalue) *Univalue })
	// for i, v := range this {
	// 	this[i] = fun.From(v)
	// }

	slice.ParallelForeach(this, 8, func(i int, _ **Univalue) {
		this[i] = fun.From(this[i])
	})

	slice.Remove((*[]*Univalue)(&this), nil)
	return this
}

func (this Univalues) PathsContain(keyword string) Univalues {
	return slice.CopyIf(this, func(_ int, v *Univalue) bool {
		return strings.Contains((*v.GetPath()), (keyword))
	})
}

// Debugging only
func (this Univalues) IfContains(target *Univalue) bool {
	for _, v := range this {
		if v.Equal(target) {
			return true
		}
	}
	return false
}

func (this Univalues) Keys() []string {
	keys := make([]string, len(this))
	for i, v := range this {
		keys[i] = *v.GetPath()
	}
	return keys
}

func (this Univalues) Values() []stgcommon.Type {
	vals := make([]stgcommon.Type, len(this))
	for i, v := range this {
		vals[i] = v.Value().(stgcommon.Type)
	}
	return vals
}

func (this Univalues) KVs() ([]string, []stgcommon.Type) {
	keys := make([]string, len(this))
	vals := make([]stgcommon.Type, len(this))
	for i, v := range this {
		keys[i] = *v.GetPath()
		if v.Value() == nil {
			vals[i] = nil
			continue
		}
		vals[i] = v.Value().(stgcommon.Type)
	}
	return keys, vals
}

// For debug only
func (this Univalues) Checksum() [32]byte {
	return sha256.Sum256(this.Encode())
}

func (this Univalues) Equal(other Univalues) bool {
	for i, v := range this {
		if !v.Equal(other[i]) {
			return false
		}
	}
	return true
}

func (this Univalues) Clone() Univalues {
	return slice.Clone(this)
}

func (this Univalues) SortByKey() Univalues {
	sort.Slice(this, func(i, j int) bool {
		if *this[i].GetPath() != *this[j].GetPath() {
			return (*this[i].GetPath()) < (*this[j].GetPath())
		}
		return this[i].GetTx() < this[j].GetTx()
	})
	return this
}

func (this Univalues) SortByDepth() Univalues {
	depths := make([]int, len(this))
	for i, v := range this {
		depths[i] = strings.Count(*v.GetPath(), "/")
	}

	slice.SortBy1st(depths, ([]*Univalue)(this), func(i, j int) bool {
		return i < j
	})
	return this
}

func (this Univalues) Sort() Univalues {
	sorter := func(i, j int) bool {
		if this[i].keyHash != this[j].keyHash {
			return this[i].keyHash < this[j].keyHash
		}

		if flag := bytes.Compare(this[i].pathBytes, this[j].pathBytes); flag != 0 {
			return flag < 0
		}

		if this[i].tx != this[j].tx {
			return this[i].tx < this[j].tx
		}

		if this[i].sequence != this[j].sequence {
			return this[i].sequence < this[j].sequence
		}
		return (this[i]).Less(this[j])
	}

	sort.Slice(this, sorter)
	// for i := 0; i < len(this); i++ {
	// 	// this[i].se = this[i].value
	// 	jobSeqIDs[i] = this[i].sequence
	// }
	return this
}

func (this Univalues) SortByTx() {
	sort.Slice(this, func(i, j int) bool {
		// if this[i].tx == this[j].tx {
		// 	if this[i].isExpanded == this[j].isExpanded && this[j].isExpanded {
		// 		panic("Two univalues with the same tx and both are substituted. This should not happen.")
		// 	}
		// 	return this[i].isExpanded // Impossible to have two substituted univalues with the same tx.
		// }
		return this[i].tx < this[j].tx
	})
}

func Sorter(univals []*Univalue) []*Univalue {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}
