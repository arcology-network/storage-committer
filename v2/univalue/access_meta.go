package univalue

import (
	"reflect"

	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/type/noncommutative"
)

type Unimeta struct {
	vType       uint8
	tx          uint32
	path        *string
	reads       uint32
	writes      uint32
	deltaWrites uint32
	preexists   bool
	reserved    interface{}
	reclaimFunc func(interface{})
}

func NewUnimeta(tx uint32, key string, reads, writes uint32, deltaWrites uint32, args ...interface{}) *Unimeta {
	return &Unimeta{
		vType:       (&Unimeta{}).GetTypeID(args[0]),
		tx:          tx,
		path:        &key,
		reads:       reads,
		writes:      writes,
		deltaWrites: deltaWrites,
	}
}

func (this *Unimeta) GetTx() uint32        { return this.tx }
func (this *Unimeta) GetPath() *string     { return this.path }
func (this *Unimeta) SetPath(path *string) { this.path = path }
func (this *Unimeta) ClearPath()           { *this.path = (*this.path)[:0] }
func (this *Unimeta) TypeID() uint8        { return this.vType }

func (this *Unimeta) Reads() uint32       { return this.reads }
func (this *Unimeta) Writes() uint32      { return this.writes } // Exist in cache as a failed read
func (this *Unimeta) DeltaWrites() uint32 { return this.deltaWrites }

func (this *Unimeta) Preexist() bool { return this.preexists } // Exist in cache as a failed read

func (this *Unimeta) IncrementReads(reads uint32)       { this.reads += reads }
func (this *Unimeta) IncrementWrites(writes uint32)     { this.writes += writes }
func (this *Unimeta) IncrementDelta(deltaWrites uint32) { this.deltaWrites += deltaWrites }

func (this *Unimeta) CheckPreexist(key string, source interface{}) bool {
	return source.(ccurlcommon.IndexerInterface).RetriveShallow(key) != nil
}

func (this *Unimeta) Clone() Unimeta {
	return Unimeta{
		vType:       this.vType,
		tx:          this.tx,
		path:        this.path,
		reads:       this.reads,
		writes:      this.writes,
		preexists:   this.preexists,
		reclaimFunc: this.reclaimFunc,
	}
}

func (*Unimeta) GetTypeID(value interface{}) uint8 {
	if value != nil {
		switch value.(type) {
		case *noncommutative.Bigint:
			return ccurlcommon.NoncommutativeBigint
		case *noncommutative.Int64:
			return ccurlcommon.NoncommutativeInt64
		case *noncommutative.String:
			return ccurlcommon.NoncommutativeString
		case *noncommutative.Bytes:
			return ccurlcommon.NoncommutativeBytes
		case *commutative.Path:
			return ccurlcommon.CommutativeMeta
		case *commutative.U256: /* Commutatives */
			return ccurlcommon.CommutativeUint256
		case *commutative.Int64:
			return ccurlcommon.CommutativeInt64
		case *commutative.Uint64:
			return ccurlcommon.CommutativeUint64
		}
	}
	return uint8(reflect.Invalid)
}

func (Unimeta) IsCommutative(value interface{}) bool {
	if value != nil {
		switch value.(type) {
		case *commutative.Path:
			return true
		case *commutative.U256: /* Commutatives */
			return true
		case *commutative.Int64:
			return true
		case *commutative.Uint64:
			return true
		}
	}
	return false
}
