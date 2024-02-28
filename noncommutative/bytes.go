package noncommutative

import (
	"bytes"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/interfaces"
)

//type Bytes []byte

type Bytes struct {
	placeholder bool //
	value       codec.Bytes
}

func NewBytes(v []byte) intf.Type {
	b := make([]byte, len(v))
	copy(b, v)
	return &Bytes{
		placeholder: true,
		value:       b,
	}
}

func (this *Bytes) Assign(v []byte) {
	this.value = v
}

func (this *Bytes) MemSize() uint32             { return uint32(1 + len(this.value)) }
func (this *Bytes) IsSelf(key interface{}) bool { return true }
func (this *Bytes) TypeID() uint8               { return BYTES }

func (this *Bytes) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

// create a new path
func (this *Bytes) Clone() interface{} {
	return &Bytes{
		placeholder: true,
		value:       slice.Clone(this.value),
	}
}

func (this *Bytes) Equal(other interface{}) bool {
	return bytes.Equal(this.value, other.(*Bytes).value)
}

func (this *Bytes) IsNumeric() bool     { return false }
func (this *Bytes) IsCommutative() bool { return false }
func (this *Bytes) IsBounded() bool     { return false }

func (this *Bytes) Value() interface{} { return this.value }
func (this *Bytes) Delta() interface{} { return this.value }
func (this *Bytes) DeltaSign() bool    { return true } // delta sign
func (this *Bytes) Min() interface{}   { return nil }
func (this *Bytes) Max() interface{}   { return nil }

func (this *Bytes) CloneDelta() interface{} { return codec.Bytes(slice.Clone(this.value)) }
func (this *Bytes) SetValue(v interface{})  { this.SetDelta(v) }

func (this *Bytes) IsDeltaApplied() bool       { return true }
func (this *Bytes) ResetDelta()                { this.SetDelta(codec.Bytes([]byte{})) }
func (this *Bytes) SetDelta(v interface{})     { copy(this.value, v.(codec.Bytes)) }
func (this *Bytes) SetDeltaSign(v interface{}) {}
func (this *Bytes) SetMin(v interface{})       {}
func (this *Bytes) SetMax(v interface{})       {}

func (this *Bytes) Get() (interface{}, uint32, uint32) { return []byte(this.value), 1, 0 }

func (this *Bytes) New(_, delta, _, _, _ interface{}) interface{} {
	v := common.IfThenDo1st(delta != nil && delta.(codec.Bytes) != nil, func() codec.Bytes { return delta.(codec.Bytes).Clone().(codec.Bytes) }, this.value)
	return &Bytes{
		true,
		v,
	}
}

func (this *Bytes) Set(value interface{}, _ interface{}) (interface{}, uint32, uint32, uint32, error) {
	if value != nil && this != value { // Avoid self copy.
		this.value = make([]byte, len(value.(*Bytes).value))
		copy(this.value, value.(*Bytes).value)
	}
	return this, 0, 1, 0, nil
}

func (this *Bytes) ApplyDelta(typedVals []intf.Type) (intf.Type, int, error) {
	// vec := v.([]*univalue.Univalue)
	for _, v := range typedVals {
		// v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Bytes)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*Bytes), nil)
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil, 0, nil
	}
	return this, len(typedVals), nil
}
