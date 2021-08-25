package commutative

import (
	"fmt"

	codec "github.com/HPISTechnologies/common-lib/codec"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/common"
)

type Int64 struct {
	finalized bool
	value     int64
	delta     int64
}

func NewInt64(value int64, delta int64) interface{} {
	return &Int64{
		false,
		value,
		delta,
	}
}

func (this *Int64) Deepcopy() interface{} {
	return &Int64{
		this.finalized,
		this.value,
		this.delta,
	}
}

func (this *Int64) Value() interface{} {
	return this.value
}

func (this *Int64) Delta() interface{} {
	return this.delta
}

func (this *Int64) ToAccess() interface{} {
	return this
}

func (this *Int64) TypeID() uint8 {
	return ccurlcommon.CommutativeInt64
}

func (this *Int64) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	if tx == ccurlcommon.SYSTEM {
		return this, 0, 0
	}
	this.finalized = true
	if this.delta == 0 {
		return this, 1, 0
	}

	this.value += this.delta
	this.delta = 0
	return this.value, 1, 1
}

func (this *Int64) Transitional(source interface{}) interface{} {
	return this.delta
}

func (this *Int64) Set(tx uint32, path string, v interface{}, source interface{}) (uint32, uint32, error) {
	this.delta += v.(*Int64).delta
	return 0, 1, nil
}

func (this *Int64) Merge(tx uint32, other interface{}) {
	if this != other {
		this.Set(tx, "", other, nil)
	}
}

func (this *Int64) Composite() bool { return !this.finalized }
func (this *Int64) Finalize()       { this.Get(0, "", nil) }

func (this *Int64) Purge() {
	this.finalized = false
	this.delta = 0
}

func (this *Int64) GobEncode() ([]byte, error) {

	return this.Encode(), nil
}
func (this *Int64) GobDecode(data []byte) error {
	myInt64 := this.Decode(data).(*Int64)
	this.delta = myInt64.delta
	this.finalized = myInt64.finalized
	this.value = myInt64.value
	return nil
}

func (this *Int64) Encode() []byte {
	return codec.Byteset{
		codec.Int64(this.value).Encode(),
		codec.Int64(this.delta).Encode(),
	}.Encode()
}

func (Int64) Decode(bytes []byte) interface{} {
	fields := codec.Byteset{}.Decode(bytes)
	return NewInt64(codec.Int64(0).Decode(fields[0]), codec.Int64(0).Decode(fields[1]))
}

func (this *Int64) EncodeStripped() []byte {
	return codec.Int64(this.value).Encode()
}

func (this *Int64) DecodeStripped(bytes []byte) interface{} {
	value := NewInt64(0, 0)
	value.(*Int64).value = codec.Int64(0).Decode(bytes)
	return value
}

func (this *Int64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}
