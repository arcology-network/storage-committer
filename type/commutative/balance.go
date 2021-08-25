package commutative

import (
	"fmt"
	"math/big"

	codec "github.com/HPISTechnologies/common-lib/codec"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/common"
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

func (this *Balance) Delta() *big.Int {
	return this.delta
}

func (this *Balance) TypeID() uint8 {
	return ccurlcommon.CommutativeBalance
}

func (this *Balance) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	if tx == ccurlcommon.SYSTEM {
		return this, 0, 0
	}
	this.finalized = true
	if this.delta.Cmp(big.NewInt(0)) == 0 {
		return this, 1, 0
	}

	this.value = this.value.Add(this.value, this.delta)
	this.delta = big.NewInt(0) // reset the delta
	return this, 1, 1
}

func (this *Balance) BigInt() *big.Int {
	return new(big.Int).Add(this.value, this.delta)
}

func (this *Balance) Transitional(source interface{}) interface{} {
	return this.delta
}

// Set delta
func (this *Balance) Set(tx uint32, path string, v interface{}, source interface{}) (uint32, uint32, error) {
	this.delta = this.delta.Add(this.delta, v.(*Balance).delta)
	return 0, 1, nil
}

func (this *Balance) Merge(tx uint32, other interface{}) {
	if this != other {
		this.Set(tx, "", other, nil)
	}
}

func (this *Balance) Finalize()       { this.Get(0, "", nil) }
func (this *Balance) Composite() bool { return !this.finalized }

func (this *Balance) Purge() {
	this.finalized = false
	this.delta = &big.Int{}
}

func (this *Balance) GobEncode() ([]byte, error) {

	return this.Encode(), nil
}
func (this *Balance) GobDecode(data []byte) error {
	balance := this.Decode(data).(*Balance)
	this.delta = balance.delta
	this.finalized = balance.finalized
	this.value = balance.value
	return nil
}

func (this *Balance) Encode() []byte {
	return codec.Byteset{
		codec.Bool(this.finalized).Encode(),
		//codec.Bigint(*this.value)
		(*codec.Bigint)(this.value).Encode(),
		(*codec.Bigint)(this.delta).Encode(),
	}.Encode()
}

func (*Balance) Decode(data []byte) interface{} {
	fields := codec.Byteset{}.Decode(data)
	balance := NewBalance(
		(&codec.Bigint{}).Decode(fields[1]),
		(&codec.Bigint{}).Decode(fields[2]),
	)

	balance.(*Balance).finalized = bool(codec.Bool(true).Decode(fields[0]))
	return balance
}

func (this *Balance) EncodeStripped() []byte {
	return (*codec.Bigint)(this.value).Encode()
}

func (this *Balance) DecodeStripped(bytes []byte) interface{} {
	return NewBalance((&codec.Bigint{}).Decode(bytes), &big.Int{})
}

func (this *Balance) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}
