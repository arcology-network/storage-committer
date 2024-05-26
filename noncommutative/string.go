package noncommutative

import (
	"strings"

	"github.com/arcology-network/common-lib/common"
	intf "github.com/arcology-network/storage-committer/interfaces"
)

type String string

func NewString(v string) intf.Type {
	var this String = String(v)
	return &this
}

func (this *String) MemSize() uint32                                            { return uint32(len(*this)) }
func (this *String) IsSelf(key interface{}) bool                                { return true }
func (this *String) TypeID() uint8                                              { return STRING }
func (this *String) Equal(other interface{}) bool                               { return *this == *(other.(*String)) }
func (this *String) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }
func (this *String) Get() (interface{}, uint32, uint32)                         { return string(*this), 1, 0 }

func (this *String) IsNumeric() bool     { return false }
func (this *String) IsCommutative() bool { return false }
func (this *String) IsBounded() bool     { return false }

func (this *String) Value() interface{} { return this }
func (this *String) Delta() interface{} { return this }
func (this *String) DeltaSign() bool    { return true } // delta sign
func (this *String) Min() interface{}   { return nil }
func (this *String) Max() interface{}   { return nil }

func (this *String) CloneDelta() interface{}         { return this.Clone() }
func (this *String) SetValue(v interface{})          { this.SetDelta(v) }
func (this *String) Preload(_ string, _ interface{}) {}

func (this *String) IsDeltaApplied() bool       { return true }
func (this *String) ResetDelta()                { this.SetDelta(common.New[String]("")) }
func (this *String) SetDelta(v interface{})     { *this = (*v.(*String)) }
func (this *String) SetDeltaSign(v interface{}) {}
func (this *String) SetMin(v interface{})       {}
func (this *String) SetMax(v interface{})       {}

func (this *String) New(_, delta, _, _, _ interface{}) interface{} {
	return common.IfThenDo1st(delta != nil && delta.(*String) != nil, func() interface{} { return delta.(*String).Clone() }, interface{}(this))
}

func (this *String) Clone() interface{} {
	value := strings.Clone(string(*this))
	return (*String)(&value)
}

func (this *String) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if value != nil {
		*this = String(*(value.(*String)))
	}
	return this, 0, 1, 0, nil
}

func (this *String) ApplyDelta(typedVals []intf.Type) (intf.Type, int, error) {
	// vec := v.([]*univalue.Univalue)
	for _, v := range typedVals {
		// v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*String)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*String), nil)
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
