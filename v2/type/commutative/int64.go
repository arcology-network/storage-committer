package commutative

import (
	"fmt"

	codec "github.com/HPISTechnologies/common-lib/codec"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
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

func (this *Int64) TypeID() uint8 { return ccurlcommon.CommutativeInt64 }

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

func (this *Int64) ToAccess() interface{} {
	return this
}

func (this *Int64) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	if this.delta == 0 {
		return this, 1, 0
	}

	this.finalized = true
	return &Int64{
		finalized: true,
		value:     this.value + this.delta,
		delta:     0,
	}, 1, 1
}

func (this *Int64) Peek(source interface{}) interface{} {
	return &Int64{
		this.finalized,
		this.value + this.delta,
		0,
	}
}

func (this *Int64) Delta(source interface{}) interface{} {
	return &Int64{
		this.finalized,
		0,
		this.delta,
	}
}

func (this *Int64) Set(tx uint32, path string, v interface{}, source interface{}) (uint32, uint32, error) {
	this.delta += v.(*Int64).delta
	return 0, 1, nil
}

func (this *Int64) ApplyDelta(tx uint32, others []ccurlcommon.UnivalueInterface) ccurlcommon.TypeInterface {
	for _, other := range others {
		if other != nil && other.Value() != nil {
			this.Set(tx, "", other.Value().(*Int64), nil)
		}
	}

	this.value += this.delta
	this.delta = 0
	return this
}

func (this *Int64) Composite() bool { return !this.finalized }

func (this *Int64) Purge() {
	this.finalized = false
	this.delta = 0
}

func (this *Int64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
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

func (this *Int64) EncodeCompact() []byte {
	return codec.Int64(this.value).Encode()
}

func (this *Int64) DecodeCompact(bytes []byte) interface{} {
	value := NewInt64(0, 0)
	value.(*Int64).value = codec.Int64(0).Decode(bytes)
	return value
}

func (this *Int64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}

func (this *Int64) GetDelta() int64 {
	return this.delta
}
