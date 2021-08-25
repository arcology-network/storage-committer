package commutative

import (
	"errors"
	"fmt"
	"math/big"

	codec "github.com/HPISTechnologies/common-lib/codec"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
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

func (this *Balance) ApplyDelta(tx uint32, others []ccurlcommon.UnivalueInterface) ccurlcommon.TypeInterface {
	for _, other := range others {
		if other != nil && other.Value() != nil {
			this.Set(tx, "", other.Value().(*Balance), nil)
		}
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

func (this *Balance) Encode() []byte {
	v := codec.Bigint(*(this.value))
	d := codec.Bigint(*(this.delta))
	return codec.Byteset{
		codec.Bool(this.finalized).Encode(),
		(&v).Encode(),
		(&d).Encode(),
	}.Encode()
}

func (*Balance) Decode(data []byte) interface{} {
	fields := codec.Byteset{}.Decode(data)
	return &Balance{
		finalized: bool(codec.Bool(true).Decode(fields[0])),
		value:     (&codec.Bigint{}).Decode(fields[1]),
		delta:     (&codec.Bigint{}).Decode(fields[2]),
	}
}

func (this *Balance) EncodeCompact() []byte {
	return this.value.Bytes()
}

func (this *Balance) DecodeCompact(bytes []byte) interface{} {
	return NewBalance((&big.Int{}).SetBytes(bytes), &big.Int{})
}

func (this *Balance) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}

func (this *Balance) GetDelta() *big.Int {
	return this.delta
}
