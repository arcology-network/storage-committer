package noncommutative

import (
	"bytes"
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Bigint big.Int

func NewBigint(v int64) interface{} {
	var value big.Int
	value.SetInt64(v)
	this := Bigint(value)
	return &this
}

func (this *Bigint) IsSelf(key interface{}) bool { return true }
func (this *Bigint) TypeID() uint8               { return uint8(ccurlcommon.NonCommutative{}.Bigint()) }

// func (this *Bigint) Latest() interface{}         { return this }
func (this *Bigint) Equal(other interface{}) bool {
	return bytes.Equal((*big.Int)(this).Bytes(), (*big.Int)(other.(*Bigint)).Bytes())
}

func (this *Bigint) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func (this *Bigint) Deepcopy() interface{} {
	v := big.Int(*this)
	return Bigint(*new(big.Int).Set(&v))
}

func (this *Bigint) Size() uint32 {
	v := codec.Bigint(*this)
	return v.Size()
}

func (this *Bigint) Get() (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *Bigint) Delta() interface{} {
	return this
}

func (this *Bigint) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if value != nil {
		*this = Bigint(*(value.(*big.Int)))
	}
	return this, 0, 1, 0, nil
}

func (this *Bigint) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	vec := v.([]ccurlcommon.UnivalueInterface)
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
		return nil
	}
	return this
}
