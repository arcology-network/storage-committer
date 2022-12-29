package commutative

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	uint256 "github.com/holiman/uint256"
)

type Balance struct {
	finalized bool
	value     uint256.Int
	min       *uint256.Int
	max       *uint256.Int
	delta     *big.Int
}

func NewBalance(initialV *big.Int, deltaV *big.Int, min, max *uint256.Int) interface{} {
	v, overflow := uint256.FromBig(initialV)
	if overflow {
		panic("Overflow!")
	}
	return &Balance{
		false,
		*v,
		min,
		max,
		deltaV,
	}
}

func (this *Balance) Deepcopy() interface{} {
	min := *this.min
	max := *this.max
	return &Balance{
		this.finalized,
		this.value,
		&min,
		&max,
		new(big.Int).Set(this.delta),
	}
}

func (this *Balance) HeaderSize() uint32 {
	return 6 * codec.UINT32_LEN
}

func (this *Balance) Value() interface{} {
	return this.value
}

func (this *Balance) ToAccess() interface{} {
	return this
}

func (this *Balance) TypeID() uint8 {
	return ccurlcommon.CommutativeUint256
}

func (*Balance) check(value uint256.Int, deltaBigInt *big.Int, min, max *uint256.Int) (bool, *uint256.Int, error) {
	b := new(big.Int).Set(deltaBigInt)
	delta, failed := uint256.FromBig(b.Abs(b))
	if failed {
		panic("Error: Failed convert to uint256!!!")
	}

	isNegative := deltaBigInt.Sign() == -1

	if isNegative {
		if value.Cmp(delta) == -1 || (min != nil && new(uint256.Int).Sub(&value, delta).Cmp(min) == -1) { // Check against the min value
			return isNegative, delta, errors.New("Error: Underflow!!!")
		}
		return isNegative, delta, nil
	}

	if max != nil && new(uint256.Int).Add(&value, delta).Cmp(max) == 1 { // Check against the MAX value
		return isNegative, delta, errors.New("Error: Overflow!!!")
	}
	return isNegative, delta, nil
}

func (this *Balance) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	if this.delta.Cmp(big.NewInt(0)) == 0 {
		return this, 1, 0
	}

	this.finalized = true
	temp := &Balance{
		finalized: this.finalized,
		value:     this.value,
		min:       this.min,
		max:       this.max,
	}

	isNegative, delta, err := this.check(temp.value, this.delta, this.min, this.max)
	if err != nil {
		return nil, 1, 1
	}

	if isNegative {
		temp.value.Sub(&temp.value, delta)
	} else {
		temp.value.Add(&temp.value, delta)
	}
	return temp, 1, 1
}

func (this *Balance) Delta(source interface{}) interface{} {
	return this
}

// Set delta
func (this *Balance) Set(tx uint32, path string, v interface{}, source interface{}) (uint32, uint32, error) {
	sum := new(big.Int).Add(this.delta, v.(*big.Int))
	_, _, err := this.check(this.value, sum, this.min, this.max)
	if err != nil {
		return 0, 1, errors.New("Wrong Value!!!")
	}

	this.delta.Add(this.delta, v.(*big.Int))
	return 0, 1, nil
}

func (this *Balance) Reset(tx uint32, path string, v interface{}, source interface{}) (uint32, uint32, error) {
	this.value = *(v.(*uint256.Int))
	this.delta = big.NewInt(0)
	this.finalized = true
	this.min = nil
	this.max = nil

	return 0, 1, nil
}

func (this *Balance) Peek(source interface{}) interface{} {
	v, _, _ := this.Deepcopy().(*Balance).Get(math.MaxUint32, "", source)
	return v
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

	isNegative, delta, err := this.check(this.value, this.delta, this.min, this.max)
	if err != nil {
		panic(err)
	}

	if isNegative {
		this.value.Sub(&this.value, delta)
	} else {
		this.value.Add(&this.value, delta)
	}

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
	// v := codec.Bigint(*this.value)
	d := codec.Bigint(*this.delta)
	return this.HeaderSize() +
		codec.Bool(this.finalized).Size() +
		//(&v).Size() +
		(&d).Size()
}

func (this *Balance) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Balance) EncodeToBuffer(buffer []byte) {

	// v := codec.Bigint(*(uint256.ToBig(*this.value)))
	d := codec.Bigint(*(this.delta))

	codec.Uint32(3).EncodeToBuffer(buffer)

	offset := uint32(0)
	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*1:])
	codec.Bool(this.finalized).EncodeToBuffer(buffer[4*codec.UINT32_LEN+offset:])
	offset += codec.Bool(this.finalized).Size()

	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*2:])
	// v.EncodeToBuffer(buffer[4*codec.UINT32_LEN+offset:])
	// offset += v.Size()

	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*3:])
	d.EncodeToBuffer(buffer[4*codec.UINT32_LEN+offset:])
}

func (*Balance) Decode(data []byte) interface{} {
	fields := codec.Byteset{}.Decode(data).(codec.Byteset)
	return &Balance{
		finalized: bool(codec.Bool(true).Decode(fields[0]).(codec.Bool)),
		// value:     (*big.Int)((&codec.Bigint{}).Decode(fields[1]).(*codec.Bigint)),
		delta: (*big.Int)((&codec.Bigint{}).Decode(fields[2]).(*codec.Bigint)),
	}
}

func (this *Balance) EncodeCompact() []byte {
	// v := codec.Bigint(*(this.value))
	vSize := len(this.value)
	totalSize := 2*codec.UINT32_LEN + vSize //(&v).Size()
	buffer := make([]byte, totalSize)

	codec.Uint32(1).EncodeToBuffer(buffer)

	offset := uint32(0)
	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*1:])
	//v.EncodeToBuffer(buffer[2*codec.UINT32_LEN+offset:])

	return buffer //this.value.Bytes()
}

func (this *Balance) DecodeCompact(bytes []byte) interface{} {
	//return NewBalance((&big.Int{}).SetBytes(bytes), &big.Int{})
	// fields := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	// return NewBalance((*big.Int)((&codec.Bigint{}).Decode(fields[0]).(*codec.Bigint)), &big.Int{})
	return nil
}

func (this *Balance) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}

func (this *Balance) GetDelta() interface{} {
	return this.delta
}
