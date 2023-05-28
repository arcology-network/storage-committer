package noncommutative

import (
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type Int64 codec.Int64

func NewInt64(v int64) interface{} {
	var this Int64 = Int64(v)
	return &this
}

func (this *Int64) MemSize() uint32                                            { return 8 }
func (this *Int64) IsSelf(key interface{}) bool                                { return true }
func (this *Int64) TypeID() uint8                                              { return INT64 }
func (this *Int64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }
func (this *Int64) Clone() interface{}                                         { return common.New(*this) }
func (this *Int64) Equal(other interface{}) bool                               { return *this == *(other.(*Int64)) }
func (this *Int64) Get() (interface{}, uint32, uint32)                         { return int64(*this), 1, 0 }

func (this *Int64) IsNumeric() bool     { return true }
func (this *Int64) IsCommutative() bool { return false }

func (this *Int64) Value() interface{} { return (this) }
func (this *Int64) Delta() interface{} { return (this) }
func (this *Int64) DeltaSign() bool    { return true } // delta sign
func (this *Int64) Min() interface{}   { return nil }
func (this *Int64) Max() interface{}   { return nil }

func (this *Int64) SetValue(v interface{}) { *this = *(v.(*Int64)) }
func (this *Int64) SetDelta(v interface{}) { *this = *(v.(*Int64)) }
func (this *Int64) SetDeltaSign(v interface{}) {
	(*this) *= (common.IfThen(v.(bool), Int64(1), Int64(-1)))
}
func (this *Int64) SetMin(v interface{}) {}
func (this *Int64) SetMax(v interface{}) {}

func (this *Int64) New(_, delta, _, _, _ interface{}) interface{} {
	return common.IfThenDo1st(delta != nil && delta.(*Int64) != nil, func() interface{} { return delta.(*Int64).Clone() }, interface{}(this))
}

func (this *Int64) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if value != nil {
		*this = Int64(*(value.(*Int64)))
	}
	return this, 0, 1, 0, nil
}

func (this *Int64) ApplyDelta(v interface{}) (interfaces.Type, int, error) {
	vec := v.([]interfaces.Univalue)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Int64)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*Int64), nil)
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
