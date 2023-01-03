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

var (
	uint256min = uint256.NewInt(0)
	uint256max = uint256.NewInt(0).SetAllOne()
)

type Balance struct {
	finalized bool
	value     *uint256.Int
	min       *uint256.Int
	max       *uint256.Int
	delta     *big.Int
}

func NewBalance(value *uint256.Int, delta *big.Int) interface{} {
	return &Balance{
		value: value,
		delta: delta,
		min:   uint256min.Clone(),
		max:   uint256max.Clone(),
	}
}

func NewBalanceWithLimit(min, max *uint256.Int) interface{} {
	return &Balance{
		min: min,
		max: max,
	}
}

func (this *Balance) Deepcopy() interface{} {
	return &Balance{
		this.finalized,
		this.value.Clone(),
		this.min.Clone(),
		this.max.Clone(),
		new(big.Int).Set(this.delta),
	}
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

func (*Balance) check(value *uint256.Int, deltaBigInt *big.Int, min, max *uint256.Int) (bool, *uint256.Int, error) {
	b := new(big.Int).Set(deltaBigInt)
	delta, failed := uint256.FromBig(b.Abs(b))
	if failed {
		return false, nil, errors.New("Error: Delta Overflow!!!")
	}

	isNegative := deltaBigInt.Sign() == -1
	if isNegative {
		if diff, overflow := new(uint256.Int).SubOverflow(value, delta); overflow {
			return isNegative, nil, errors.New("Error: Underflow!!!")
		} else {
			if min != nil && diff.Cmp(min) == -1 {
				return isNegative, nil, errors.New("Error: Sum overflow!!!")
			}
		}
	} else {
		if sum, overflow := new(uint256.Int).AddOverflow(value, delta); overflow {
			return isNegative, nil, errors.New("Error: Sum overflow!!!")
		} else {
			if max != nil && sum.Cmp(max) == 1 {
				return isNegative, nil, errors.New("Error: Sum overflow!!!")
			}
		}
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
		delta:     big.NewInt(0),
	}

	isNegative, delta, err := this.check(temp.value, this.delta, this.min, this.max)
	if err != nil {
		return nil, 1, 1
	}

	if isNegative {
		temp.value.Sub(temp.value, delta)
	} else {
		temp.value.Add(temp.value, delta)
	}
	return temp, 1, 1
}

func (this *Balance) Delta(source interface{}) interface{} {
	return this
}

// Set delta
func (this *Balance) Set(tx uint32, path string, v interface{}, source interface{}) (uint32, uint32, error) {
	b := v.(*Balance)
	if _, _, err := this.check(this.value, new(big.Int).Add(this.delta, b.delta), this.min, this.max); err != nil {
		return 0, 1, err
	}

	this.delta.Add(this.delta, b.delta)
	return 0, 1, nil
}

func (this *Balance) Reset(tx uint32, path string, v interface{}, source interface{}) (uint32, uint32, error) {
	b := v.(*Balance)
	if b.value != nil {
		this.value = b.value
	}
	if b.delta != nil {
		this.delta = b.delta
	}
	if b.min != nil {
		this.min = b.min
	}
	if b.max != nil {
		this.max = b.max
	}

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
			if _, _, err := this.Set(tx, "", v.(*Balance), nil); err != nil {
				panic(err)
			}
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	newValue, _, _ := this.Get(tx, "", nil)
	*this = (*newValue.(*Balance))
	return this
}

func (this *Balance) Composite() bool { return !this.finalized }

func (this *Balance) Purge() {
	this.finalized = false
	this.delta = big.NewInt(0)
}

func (this *Balance) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Balance) HeaderSize() uint32 {
	return (1 + 5) * codec.UINT32_LEN // Total number of fields + offsets of these fields
}

func (this *Balance) Size() uint32 {
	delta := codec.Bigint(*this.delta)
	return codec.Bool(this.finalized).Size() +
		32 + // Values
		32 + // Min
		32 + // Max
		delta.Size()
}

func (this *Balance) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Balance) EncodeToBuffer(buffer []byte) int {
	offset := codec.Bool(this.finalized).EncodeToBuffer(buffer)
	offset += codec.Uint64s(this.value[:]).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64s(this.min[:]).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64s(this.max[:]).EncodeToBuffer(buffer[offset:])

	delta := codec.Bigint(*this.delta)
	return offset + (&delta).EncodeToBuffer(buffer[offset:])
}

func (*Balance) Decode(buffer []byte) interface{} {
	balance := &Balance{
		finalized: bool(codec.Bool(true).Decode(buffer).(codec.Bool)),
		value:     uint256.NewInt(0),
		min:       uint256.NewInt(0),
		max:       uint256.NewInt(0),
	}

	copy(balance.value[:], codec.Uint64s{}.Decode(buffer[1+32*0:1+32*1]).(codec.Uint64s))
	copy(balance.min[:], codec.Uint64s{}.Decode(buffer[1+32*1:1+32*2]).(codec.Uint64s))
	copy(balance.max[:], codec.Uint64s{}.Decode(buffer[1+32*2:1+32*3]).(codec.Uint64s))
	balance.delta = (*big.Int)((&codec.Bigint{}).Decode(buffer[1+32*3:]).(*codec.Bigint))
	return balance
}

func (this *Balance) EncodeCompact() []byte {
	totalLen := 32
	if this.min != nil {
		totalLen += 32
	}

	if this.max != nil {
		totalLen += 32
	}

	buffer := make([]byte, 2+totalLen) // labels + actual length

	offset := codec.Uint64s(this.value[:]).EncodeToBuffer(buffer[2:])
	if this.min != nil {
		buffer[0] = 1
		offset += codec.Uint64s(this.min[:]).EncodeToBuffer(buffer[offset+2:])
	}

	if this.max != nil {
		buffer[1] = 1
		codec.Uint64s(this.max[:]).EncodeToBuffer(buffer[offset+2:])
	}
	return buffer
}

func (this *Balance) DecodeCompact(buffer []byte) interface{} {
	v := uint256.NewInt(0)
	offset := 2
	copy((*v)[:], codec.Uint64s{}.Decode(buffer[offset:offset+32]).(codec.Uint64s))

	var min, max *uint256.Int
	offset += 32
	if buffer[0] == 1 {
		min = uint256.NewInt(0)
		copy(min[:], codec.Uint64s{}.Decode(buffer[offset:offset+32]).(codec.Uint64s))
		offset += 32
	}

	if buffer[1] == 1 {
		max = uint256.NewInt(0)
		copy(max[:], codec.Uint64s{}.Decode(buffer[offset:offset+32]).(codec.Uint64s))
	}

	return &Balance{
		value: v,
		delta: big.NewInt(0),
		min:   min,
		max:   max,
	}
}

func (this *Balance) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}

func (this *Balance) GetDelta() interface{} {
	return this.delta
}
