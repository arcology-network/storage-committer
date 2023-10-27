package noncommutative

import (
	"fmt"
	"math/big"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/evm/rlp"
)

func (this *Int64) Size() uint32 {
	return 8 // 8 bytes
}
func (this *Int64) Encode() []byte {
	return codec.Int64(*this).Encode()
}

func (this *Int64) EncodeToBuffer(buffer []byte) int {
	return codec.Int64(*this).EncodeToBuffer(buffer)
}

func (*Int64) Decode(bytes []byte) interface{} {
	this := Int64(codec.Int64(0).Decode(bytes).(codec.Int64))
	return &this
}

func (this *Int64) Reset() {}

func (this *Int64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

func (this *Int64) Print() {
	fmt.Println(*this)
	fmt.Println()
}

func (this *Int64) StorageEncode() []byte {
	buffer, _ := rlp.EncodeToBytes(new(big.Int).SetInt64(int64(*this.Value().(*Int64))))
	return buffer
}

func (this *Int64) StorageDecode(buffer []byte) interface{} {
	var v big.Int
	rlp.DecodeBytes(buffer, &v)
	*this = Int64(v.Uint64())
	return this
}
