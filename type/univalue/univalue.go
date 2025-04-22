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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/common"
	"github.com/cespare/xxhash"
)

// THe univalue is a combination of a value and a property field that contains the access information about the value.
type Univalue struct {
	Property
	value any
	cache []byte
}

func NewUnivalue(tx uint64, key string, reads, writes uint32, deltaWrites uint32, T any, source any) *Univalue {
	univ := &Univalue{
		Property{
			vType:         common.IfThenDo1st(T != nil, func() uint8 { return T.(intf.Type).TypeID() }, uint8(reflect.Invalid)),
			tx:            tx,
			path:          &key,
			keyHash:       xxhash.Sum64String(key),
			reads:         reads,
			writes:        writes,
			deltaWrites:   deltaWrites,
			sizeInStorage: 0,
			preexists:     common.IfThenDo1st(source != nil, func() bool { return (&Property{}).CheckPreexist(key, source) }, false),
		},
		T,
		[]byte{},
	}

	if source != nil {
		if v, err := source.(intf.ReadOnlyStore).RetriveFromStorage(key, T); err != nil {
			univ.sizeInStorage = v.(intf.Type).MemSize()
		}
	}
	return univ
}

func (*Univalue) New(meta, value, cache any) *Univalue {
	return &Univalue{
		*meta.(*Property),
		value,
		cache.([]byte),
	}
}

func (*Univalue) Reset(this *Univalue) {
	this.Property.Reset()
	this.ClearCache()
	this.value = nil
}

func (this *Univalue) From(v *Univalue) any { return v }

// func (this *Univalue) IsHotLoaded() bool             { return this.reads > 1 }
func (this *Univalue) SetTx(txId uint64) { this.tx = txId }
func (this *Univalue) ClearCache()       { this.cache = this.cache[:0] }
func (this *Univalue) Value() any        { return this.value }
func (this *Univalue) SetValue(newValue any) *Univalue {
	if this.value != nil && reflect.TypeOf(this.value) != reflect.TypeOf(newValue) && newValue != nil {
		panic("Wrong type")
	}

	this.value = newValue
	return this
}

func (this *Univalue) GetCache() any { return this.cache }

func (this *Univalue) Init(tx uint64, key string, reads, writes, deltaWrites uint32, v any, dataSource ...any) *Univalue {
	this.vType = common.IfThenDo1st(v != nil, func() uint8 { return v.(intf.Type).TypeID() }, uint8(reflect.Invalid))
	this.tx = tx
	this.path = &key
	this.keyHash = xxhash.Sum64String(key)
	this.reads = reads
	this.writes = writes
	this.deltaWrites = deltaWrites
	this.value = v
	this.preexists = common.IfThenDo1st(len(dataSource) > 0, func() bool { return (&Property{}).CheckPreexist(key, dataSource[0]) }, false)

	this.sizeInStorage = 0
	if v, _ := dataSource[0].(intf.ReadOnlyStore).RetriveFromStorage(key, v); v != nil {
		this.sizeInStorage = v.(intf.Type).MemSize()
	}

	return this
}

func (this *Univalue) Reclaim() {
	if this.reclaimFunc != nil {
		this.reclaimFunc(this)
	}
}

// This performs the action on the value and returns the result.
// This function doesnn't make a deep copy of the original value.
// It should be used for read-only operations ONLY!!!.
func (this *Univalue) Do(tx uint64, path string, doer any) any {
	r, w, dw, ret := doer.(func(any) (uint32, uint32, uint32, any))(this)
	this.reads += r
	this.writes += w
	this.deltaWrites += dw
	return ret
}

func (this *Univalue) Get(tx uint64, path string, source any) any {
	if this.value != nil {
		tempV, r, w := this.value.(intf.Type).Get() //RW: Affiliated reads and writes
		this.reads += r

		// The whole copy mechansim is designed to avoid interference with maximum performance,
		// so deep copy is made only when necessary. The criteria for making a deep copy are:
		// 1. The value needs to be updated.
		// 2. The value is modified for the first time, when writes == 0 && deltaWrites == 0.
		// So in the following cases, if we record a write with a deep copy, it will effectivly stop making deep copy in the future.
		// This is problematic because it will change the value in the global object cache as well.
		if w > 0 {
			this.MakeDeepCopy(this.value)
		}

		this.writes += w
		return tempV
	}
	this.IncrementReads(1)
	return this.value
}

func (this *Univalue) CopyTo(writable any) {
	writeCache := writable.(interface {
		Read(uint64, string, any) (any, any, uint64)
		Write(uint64, string, any) (int64, error)
		Find(uint64, string, any) (any, any)
	})

	if this.writes == 0 && this.deltaWrites == 0 {
		writeCache.Read(this.tx, *this.GetPath(), this.value)
	} else {
		writeCache.Write(this.tx, *this.GetPath(), this.value)
	}

	_, univ := writeCache.Find(this.tx, *this.GetPath(), nil)
	if this == univ {
		return
	}

	univ.(*Univalue).IncrementReads(this.Reads())
	univ.(*Univalue).IncrementWrites(this.Writes())
	univ.(*Univalue).IncrementDeltaWrites(this.DeltaWrites())
}

func (this *Univalue) Set(tx uint64, path string, newV any, inCache bool, importer any) error { // update the value
	this.tx = tx

	// Delete an non-existing value or deleting an entry that has been deleted already.
	if this.value == nil && newV == nil {
		this.writes++
		return errors.New("Error: The value doesn't exists")
	}

	// Write a new value
	if this.value == nil {
		this.vType = newV.(intf.Type).TypeID()
		v, r, w, dw := newV.(intf.Type).CopyTo(newV)
		this.value = v
		this.writes += w
		this.reads += r
		this.deltaWrites += dw
		return nil
	}

	// To avoid interference with the value in the global object cache.
	this.MakeDeepCopy(newV)

	oldV := this.value.(intf.Type)
	v, r, w, dw, err := oldV.Set(newV, []any{path, *this.path, tx, importer}) // Update the current value
	this.value = v
	this.writes += w
	this.reads += r
	this.deltaWrites += dw

	if newV == nil && this.Value().(intf.Type).IsSelf(path) { // Delete the entry but keep the access record.
		this.vType = uint8(reflect.Invalid)
		this.value = newV // Delete the value
		this.writes++
		this.isDeleted = true // This is a delete operation
	}
	return err
}

// Making a deep copy may be necessary to avoid interference with
// the value in the global object cache.
func (this *Univalue) MakeDeepCopy(newV any) {
	// writes == 0 && deltaWrites == 0 means the value has been modified already.
	// this.value == nil, this is a new value assignment, so we don't need to make a deep copy.
	// typedV == nil, this is a delete operation, so we don't need to make a deep copy.
	// In cascading write cache, the values' access info will stripped off, so it wouldn't introduce interference.
	if this.writes == 0 && this.deltaWrites == 0 && this.value != nil { // Make a deep copy if has't done so
		this.value = this.value.(intf.Type).Clone()
	}
}

// Check & Merge attributes
func (this *Univalue) ApplyDelta(vec []*Univalue) error {
	// vec := v.([]*Univalue)

	/* Precheck & Merge attributes*/
	for i := 0; i < len(vec); i++ {
		this.PrecheckAttributes(vec[i])
		this.writes += vec[i].Writes()
		this.reads += vec[i].Reads()
		this.deltaWrites += vec[i].DeltaWrites()
	}

	// Apply transitions
	typedVals := slice.Transform(vec, func(_ int, v *Univalue) intf.Type {
		if v.Value() != nil {
			return v.Value().(intf.Type)
		}
		return nil
	})

	var err error
	if this.Value() != nil {
		if this.value, _, err = this.Value().(intf.Type).ApplyDelta(typedVals); err != nil {
			return err
		}
	}
	return nil
}

func (this *Univalue) IsReadOnly() bool       { return (this.writes == 0 && this.deltaWrites == 0) }
func (this *Univalue) IsWriteOnly() bool      { return (this.reads == 0 && this.deltaWrites == 0) }
func (this *Univalue) IsDeltaWriteOnly() bool { return (this.reads == 0 && this.writes == 0) }
func (this *Univalue) IsDeleteOnly() bool {
	return this.isDeleted && this.reads == 0 && this.deltaWrites == 0 // Cannot just use value == nil, because it may be a new value.
}

func (this *Univalue) IsNilInitOnly() bool {
	return this.Value() == nil && !this.isDeleted && this.reads == 0 && this.deltaWrites == 0
}

// Commutative write is no longer treated as a conflict with read.
// Write without read happens when a new value is created.
func (this *Univalue) IsCommutativeInitOnly() bool {
	return this.Value() != nil &&
		this.Value().(intf.Type).IsCommutative() &&
		this.Value().(intf.Type).IsNumeric() &&
		this.Reads() == 0
}

func (this *Univalue) PrecheckAttributes(other *Univalue) {
	if other.reads == 0 && other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Read/Write/Deltawrite all zero!!")
	}

	if other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Value type mismatched!") // Read only variable should never be here.
	}

	if this.GetTx() != other.GetTx() &&
		this.preexists &&
		this.Value() != nil &&
		this.Value().(intf.Type).IsCommutative() &&
		this.Reads() > 0 &&
		this.IsDeltaWriteOnly() == other.IsDeltaWriteOnly() {
		this.Print()
		fmt.Println("================================================================")
		other.Print()
		panic("Error: The composite attribute must match in different transitions")
	}

	if this.Value() == nil && this.IsDeltaWriteOnly() {
		panic("Error: A deleted value cann't be composite")
	}

	if !this.preexists && this.IsDeltaWriteOnly() {
		panic("Error: A new value cann't be composite")
	}
}

func (this *Univalue) Clone() any {
	v := &Univalue{
		this.Property.Clone(),
		common.IfThenDo1st(this.value != nil, func() any { return this.value.(intf.Type).Clone() }, this.value),
		slice.Clone(this.cache),
	}
	return v
}

func (this *Univalue) Less(other *Univalue) bool {
	if (this.value == nil || other.value == nil) && (this.value != other.value) {
		return this.value == nil
	}

	if this.writes != other.writes {
		return this.writes > other.writes
	}

	if this.reads != other.reads {
		return this.reads > other.reads
	}

	if this.deltaWrites != other.deltaWrites {
		return this.deltaWrites > other.deltaWrites
	}

	if (!this.preexists || !other.preexists) && (this.preexists != other.preexists) {
		return this.preexists
	}
	return true
}

func (this *Univalue) Checksum() [32]byte {
	return sha256.Sum256(this.Encode())
}

func (this *Univalue) Print() {
	spaces := " " //fmt.Sprintf("%"+strconv.Itoa(len(strings.Split(*this.path, "/"))*1)+"v", " ")
	fmt.Print(spaces+"tx: ", this.tx)
	fmt.Print(spaces+"reads: ", this.reads)
	fmt.Print(spaces+"writes: ", this.writes)
	fmt.Print(spaces+"DeltaWrites: ", this.deltaWrites)
	fmt.Print(spaces+"persistent: ", this.persistent)
	fmt.Print(spaces+"preexists: ", this.preexists)

	path := *this.path
	if index := strings.Index(path, "container/"); index != -1 {
		path = path[:index] + "container/" + hex.EncodeToString([]byte(path[index:]))
	}

	fmt.Print(spaces+"path: ", path, "      ")
	common.IfThenDo(this.value != nil, func() { this.value.(intf.Type).Print() }, func() { fmt.Print("nil") })
	fmt.Println()
}

func (this *Univalue) Equal(other *Univalue) bool {
	if this.value == nil && other.Value() == nil {
		return true
	}

	if (this.value == nil && other.Value() != nil) || (this.value != nil && other.Value() == nil) {
		return false
	}

	vFlag := this.value.(intf.Type).Equal(other.Value().(intf.Type))
	return this.tx == other.GetTx() &&
		*this.path == *other.GetPath() &&
		this.reads == other.Reads() &&
		this.writes == other.Writes() &&
		vFlag &&
		this.preexists == other.Preexist()
}
