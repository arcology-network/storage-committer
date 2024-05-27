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

type Property struct {
	vType       uint8
	persistent  bool
	tx          uint32
	path        *string
	reads       uint32
	writes      uint32
	deltaWrites uint32
	preexists   bool
	msg         string
	reclaimFunc func(interface{})
}

func NewProperty(tx uint32, key string, reads, writes uint32, deltaWrites uint32, vType uint8, persistent, preexists bool) *Property {
	return &Property{
		vType:       vType,
		persistent:  persistent, // Won't be affected by conflict status
		tx:          tx,
		path:        &key,
		reads:       reads,
		writes:      writes,
		deltaWrites: deltaWrites,
	}
}

func (this *Property) Reset() {
	this.vType = 0
	this.persistent = false // Won't be affected by conflict status
	this.tx = 0
	this.path = nil
	this.reads = 0
	this.writes = 0
	this.deltaWrites = 0
	this.reclaimFunc = nil
}

func (this *Property) Merge(other *Property) {
	this.reads += other.reads
	this.writes += other.writes
	this.deltaWrites += other.deltaWrites
	this.persistent = this.persistent || other.persistent
}

func (this *Property) GetMsg() string       { return this.msg }
func (this *Property) SetMsg(msg string)    { this.msg = msg }
func (this *Property) AppendMsg(msg string) { this.msg = this.msg + "\n" + msg }

func (this *Property) GetPersistent() bool  { return this.persistent }
func (this *Property) SetPersistent(v bool) { this.persistent = v }

func (this *Property) GetTx() uint32     { return this.tx }
func (this *Property) SetTx(txId uint32) { this.tx = txId }

func (this *Property) GetPath() *string     { return this.path }
func (this *Property) SetPath(path *string) { this.path = path }
func (this *Property) ClearPath()           { *this.path = (*this.path)[:0] }

func (this *Property) TypeID() uint8 { return this.vType }

func (this *Property) Reads() uint32       { return this.reads }
func (this *Property) Writes() uint32      { return this.writes } // Exist in cache as a failed read
func (this *Property) DeltaWrites() uint32 { return this.deltaWrites }

func (this *Property) IncrementReads(reads uint32)             { this.reads += reads }
func (this *Property) IncrementWrites(writes uint32)           { this.writes += writes }
func (this *Property) IncrementDeltaWrites(deltaWrites uint32) { this.deltaWrites += deltaWrites }

func (this *Property) IsReadOnly() bool { return this.Writes() == 0 && this.DeltaWrites() == 0 }
func (this *Property) Preexist() bool   { return this.preexists } // Exist in cache as a failed read
func (this *Property) Persistent() bool { return this.persistent }

func (this *Property) CheckPreexist(key string, source interface{}) bool {
	return source.(interface{ IfExists(string) bool }).IfExists(key)
}

func (this *Property) Equal(other *Property) bool {
	return this.vType == other.vType &&
		this.tx == other.tx &&
		*this.path == *other.path &&
		this.reads == other.reads &&
		this.writes == other.writes &&
		this.deltaWrites == other.deltaWrites &&
		this.msg == other.msg
}

func (this *Property) Clone() Property {
	return Property{
		vType:       this.vType,
		tx:          this.tx,
		path:        this.path,
		reads:       this.reads,
		deltaWrites: this.deltaWrites,
		writes:      this.writes,
		preexists:   this.preexists,
		reclaimFunc: this.reclaimFunc,
		msg:         this.msg,
	}
}
