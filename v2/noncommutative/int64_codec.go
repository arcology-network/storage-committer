package noncommutative

import (
	"fmt"

	"github.com/arcology-network/common-lib/codec"
)

func (this *Int64) Size(...interface{}) uint32 {
	return 8 // 8 bytes
}
func (this *Int64) Encode(...interface{}) []byte {
	return codec.Int64(int64(*this)).Encode()
}

func (this *Int64) EncodeToBuffer(buffer []byte, _ ...interface{}) int {
	return codec.Int64(*this).EncodeToBuffer(buffer)
}

func (*Int64) Decode(bytes []byte) interface{} {
	this := Int64(codec.Int64(0).Decode(bytes).(codec.Int64))
	return &this
}

// func (this *Int64) Encode() []byte {
// 	return this.Encode()
// }

// func (this *Int64) DecodeCompact(bytes []byte) interface{} {
// 	return this.Decode(bytes)
// }

func (this *Int64) Purge() {}

func (this *Int64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

func (this *Int64) Print() {
	fmt.Println(*this)
	fmt.Println()
}
