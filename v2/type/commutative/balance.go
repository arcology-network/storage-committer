package commutative

import (
	"errors"
	"fmt"
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Balance struct {
	finalized bool
	value     *big.Int
	delta     *big.Int
}

func NewBalance(initialV *big.Int, deltaV *big.Int) interface{} {
	return &Balance{
		false,
		initialV,
		deltaV,
	}
}

func (this *Balance) Deepcopy() interface{} {
	return &Balance{
		this.finalized,
		new(big.Int).Set(this.value),
		new(big.Int).Set(this.delta),
	}
}

func (this *Balance) HeaderSize() uint32 {
	return 4 * codec.UINT32_LEN
}

func (this *Balance) Value() interface{} {
	return this.value
}

func (this *Balance) ToAccess() interface{} {
	return this
}

func (this *Balance) TypeID() uint8 {
	return ccurlcommon.CommutativeBalance
}

func (this *Balance) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	if this.delta.Cmp(big.NewInt(0)) == 0 {
		return this, 1, 0
	}

	this.finalized = true
	temp := &Balance{
		finalized: true,
		value:     new(big.Int).Add(this.value, this.delta),
		delta:     big.NewInt(0),
	}

	if temp.value.Cmp(big.NewInt(0)) == -1 {
		panic("Balance cannot be negative")
	}
	return temp, 1, 1
}

func (this *Balance) Delta(source interface{}) interface{} {
	return this
}

// Set delta
func (this *Balance) Set(tx uint32, path string, v interface{}, source interface{}) (uint32, uint32, error) {
	if this.value.Cmp(big.NewInt(0)) == -1 {
		return 0, 1, errors.New("Balance cannot be negative")
	}

	this.delta = this.delta.Add(this.delta, v.(*Balance).delta)
	return 0, 1, nil
}

func (this *Balance) Peek(source interface{}) interface{} {
	return &Balance{
		this.finalized,
		new(big.Int).Add(this.value, this.delta),
		big.NewInt(0),
	}
}

func (this *Balance) ApplyDelta(tx uint32, v interface{}) ccurlcommon.TypeInterface {
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Balance)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(tx, "", v.(*Balance), nil)
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil
	}

	this.value = this.value.Add(this.value, this.delta)
	this.delta = big.NewInt(0) // reset the delta
	return this
}

func (this *Balance) Composite() bool { return !this.finalized }

func (this *Balance) Purge() {
	this.finalized = false
	this.delta = &big.Int{}
}

func (this *Balance) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Balance) Size() uint32 {
	v := codec.Bigint(*this.value)
	d := codec.Bigint(*this.delta)
	return this.HeaderSize() +
		codec.Bool(this.finalized).Size() +
		(&v).Size() +
		(&d).Size()
}

func (this *Balance) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Balance) EncodeToBuffer(buffer []byte) {
	v := codec.Bigint(*(this.value))
	d := codec.Bigint(*(this.delta))

	codec.Uint32(3).EncodeToBuffer(buffer)

	offset := uint32(0)
	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*1:])
	codec.Bool(this.finalized).EncodeToBuffer(buffer[4*codec.UINT32_LEN+offset:])
	offset += codec.Bool(this.finalized).Size()

	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*2:])
	v.EncodeToBuffer(buffer[4*codec.UINT32_LEN+offset:])
	offset += v.Size()

	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*3:])
	d.EncodeToBuffer(buffer[4*codec.UINT32_LEN+offset:])
}

func (*Balance) Decode(data []byte) interface{} {
	fields := codec.Byteset{}.Decode(data).(codec.Byteset)
	return &Balance{
		finalized: bool(codec.Bool(true).Decode(fields[0]).(codec.Bool)),
		value:     (*big.Int)((&codec.Bigint{}).Decode(fields[1]).(*codec.Bigint)),
		delta:     (*big.Int)((&codec.Bigint{}).Decode(fields[2]).(*codec.Bigint)),
	}
}

func (this *Balance) EncodeCompact() []byte {
	v := codec.Bigint(*(this.value))
	totalSize := 2*codec.UINT32_LEN + (&v).Size()
	buffer := make([]byte, totalSize)

	codec.Uint32(1).EncodeToBuffer(buffer)

	offset := uint32(0)
	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*1:])
	v.EncodeToBuffer(buffer[2*codec.UINT32_LEN+offset:])

	return buffer //this.value.Bytes()
}

func (this *Balance) DecodeCompact(bytes []byte) interface{} {
	//return NewBalance((&big.Int{}).SetBytes(bytes), &big.Int{})
	fields := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	return NewBalance((*big.Int)((&codec.Bigint{}).Decode(fields[0]).(*codec.Bigint)), &big.Int{})
}

func (this *Balance) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}

func (this *Balance) GetDelta() interface{} {
	return this.delta
}
