package univalue

import "github.com/arcology-network/concurrenturl/interfaces"

type Unimeta struct {
	vType       uint8
	persistent  bool
	tx          uint32
	path        *string
	reads       uint32
	writes      uint32
	deltaWrites uint32
	preexists   bool
	reclaimFunc func(interface{})
}

func NewUnimeta(tx uint32, key string, reads, writes uint32, deltaWrites uint32, vType uint8, persistent, preexists bool) *Unimeta {
	return &Unimeta{
		vType:       vType,
		persistent:  persistent,
		tx:          tx,
		path:        &key,
		reads:       reads,
		writes:      writes,
		deltaWrites: deltaWrites,
	}
}

func (this *Unimeta) GetPersistent() bool  { return this.persistent }
func (this *Unimeta) SetPersistent(v bool) { this.persistent = v }

func (this *Unimeta) GetTx() uint32     { return this.tx }
func (this *Unimeta) SetTx(txId uint32) { this.tx = txId }

func (this *Unimeta) GetPath() *string     { return this.path }
func (this *Unimeta) SetPath(path *string) { this.path = path }
func (this *Unimeta) ClearPath()           { *this.path = (*this.path)[:0] }

func (this *Unimeta) TypeID() uint8 { return this.vType }

func (this *Unimeta) Reads() uint32       { return this.reads }
func (this *Unimeta) Writes() uint32      { return this.writes } // Exist in cache as a failed read
func (this *Unimeta) DeltaWrites() uint32 { return this.deltaWrites }

func (this *Unimeta) IncrementReads(reads uint32)             { this.reads += reads }
func (this *Unimeta) IncrementWrites(writes uint32)           { this.writes += writes }
func (this *Unimeta) IncrementDeltaWrites(deltaWrites uint32) { this.deltaWrites += deltaWrites }

func (this *Unimeta) IsReadOnly() bool { return this.Writes() == 0 && this.DeltaWrites() == 0 }
func (this *Unimeta) Preexist() bool   { return this.preexists } // Exist in cache as a failed read
func (this *Unimeta) Persistent() bool { return this.persistent }

func (this *Unimeta) CheckPreexist(key string, source interface{}) bool {
	return source.(interfaces.Importer).RetriveShallow(key) != nil
}

func (this *Unimeta) Equal(other *Unimeta) bool {
	return this.vType == other.vType &&
		this.tx == other.tx &&
		*this.path == *other.path &&
		this.reads == other.reads &&
		this.writes == other.writes &&
		this.deltaWrites == other.deltaWrites
}

func (this *Unimeta) Clone() Unimeta {
	return Unimeta{
		vType:       this.vType,
		tx:          this.tx,
		path:        this.path,
		reads:       this.reads,
		deltaWrites: this.deltaWrites,
		writes:      this.writes,
		preexists:   this.preexists,
		reclaimFunc: this.reclaimFunc,
	}
}
