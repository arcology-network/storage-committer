package noncommutative

import (
	"bytes"
	"math/big"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/evm/rlp"
)

// type Bigint codec.Bigint

type Bigint big.Int //has anonymous camera

func NewBigint(v int64) interface{} {
	var value big.Int
	value.SetInt64(v)
	this := Bigint(value)
	return &this
}

func (this *Bigint) MemSize() uint32                                            { return uint32((*big.Int)(this).BitLen()) }
func (this *Bigint) IsSelf(key interface{}) bool                                { return true }
func (this *Bigint) TypeID() uint8                                              { return BIGINT }
func (this *Bigint) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *Bigint) Equal(other interface{}) bool {
	return bytes.Equal((*big.Int)(this).Bytes(), (*big.Int)(other.(*Bigint)).Bytes())
}

func (this *Bigint) Clone() interface{} {
	v := big.Int(*this)
	return (*Bigint)(new(big.Int).Set(&v))
}

func (this *Bigint) IsNumeric() bool     { return true }
func (this *Bigint) IsCommutative() bool { return false }
func (this *Bigint) IsBounded() bool     { return false }

func (this *Bigint) Value() interface{} { return (this) }
func (this *Bigint) Delta() interface{} { return (this) }
func (this *Bigint) DeltaSign() bool    { return true } // delta sign
func (this *Bigint) Min() interface{}   { return nil }
func (this *Bigint) Max() interface{}   { return nil }

func (this *Bigint) CloneDelta() interface{} { return this.Clone() }

func (this *Bigint) SetValue(v interface{}) { this.SetDelta(v) }

func (this *Bigint) IsDeltaApplied() bool       { return true }
func (this *Bigint) ResetDelta()                { this.SetDelta(big.NewInt(0)) }
func (this *Bigint) SetDelta(v interface{})     { (*big.Int)(this).Set((*big.Int)(v.(*Bigint))) }
func (this *Bigint) SetDeltaSign(v interface{}) {}
func (this *Bigint) SetMin(v interface{})       {}
func (this *Bigint) SetMax(v interface{})       {}

func (this *Bigint) Get() (interface{}, uint32, uint32) { return (*big.Int)(this), 1, 0 }

func (this *Bigint) FromRawType(v interface{}) interface{} {
	if common.IsType[*big.Int](v) {
		v = (*codec.Bigint)(v.(*big.Int))
	}
	return v
}

func (this *Bigint) New(_, delta, _, _, _ interface{}) interface{} {
	return common.IfThenDo1st(delta != nil && delta.(*Bigint) != nil, func() interface{} { return delta.(*Bigint).Clone() }, interface{}(this))
}

func (this *Bigint) Set(value interface{}, _ interface{}) (interface{}, uint32, uint32, uint32, error) {
	if value != nil {
		*this = Bigint(*(value.(*big.Int)))
	}
	return this, 0, 1, 0, nil
}

func (this *Bigint) ApplyDelta(v interface{}) (interfaces.Type, int, error) {
	vec := v.([]interfaces.Univalue)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Bigint)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*Bigint), nil)
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil, 0, nil
	}
	return this, len(vec), nil
}

func (this *Bigint) StorageEncode() []byte {
	buffer, _ := rlp.EncodeToBytes(*this)
	return buffer
}

func (this *Bigint) StorageDecode(buffer []byte) interface{} {
	rlp.DecodeBytes(buffer, this)
	return this
}
