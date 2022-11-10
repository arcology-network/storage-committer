package commutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
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

func (this *Int64) HeaderSize() uint32 {
	return 0 // 8 bytes, static sizes only , no header needed,
}

func (this *Int64) Size() uint32 {
	return this.HeaderSize() + // No need to encode this.finalized
		codec.Bool(this.finalized).Size() +
		codec.Int64(this.value).Size() +
		codec.Int64(this.delta).Size()
}

func (this *Int64) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Int64) EncodeToBuffer(buffer []byte) {
	codec.Bool(this.finalized).EncodeToBuffer(buffer)
	codec.Int64(this.value).EncodeToBuffer(buffer[1 : codec.UINT64_LEN+1])
	codec.Int64(this.delta).EncodeToBuffer(buffer[codec.UINT64_LEN+1:])
}

func (this *Int64) Decode(buffer []byte) interface{} {
	this = NewInt64(int64(
		codec.Int64(0).Decode(buffer[1:codec.UINT64_LEN+1]).(codec.Int64)),
		int64(codec.Int64(0).Decode(buffer[codec.UINT64_LEN+1:]).(codec.Int64))).(*Int64)

	this.finalized = bool(codec.Bool(this.finalized).Decode(buffer).(codec.Bool))
	return this
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

func (this *Int64) ApplyDelta(tx uint32, v interface{}) ccurlcommon.TypeInterface {
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Int64)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(tx, "", v.(*Int64), nil)
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil
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

func (this *Int64) EncodeCompact() []byte {
	return codec.Int64(this.value).Encode()
}

func (this *Int64) DecodeCompact(bytes []byte) interface{} {
	value := NewInt64(0, 0)
	value.(*Int64).value = int64(codec.Int64(0).Decode(bytes).(codec.Int64))
	return value
}

func (this *Int64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}

func (this *Int64) GetDelta() interface{} {
	return this.delta
}
