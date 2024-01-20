package noncommutative

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	codec "github.com/arcology-network/common-lib/codec"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func (this *Bytes) HeaderSize() uint32 {
	return 3 * codec.UINT32_LEN
}

func (this *Bytes) Size() uint32 {
	return this.HeaderSize() + this.MemSize()
}

func (this *Bytes) Encode() []byte {
	byteset := [][]byte{
		codec.Bool(this.placeholder).Encode(),
		this.value,
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Bytes) EncodeToBuffer(buffer []byte) int {
	offset := codec.Bool(this.placeholder).EncodeToBuffer(buffer)
	return offset + codec.Bytes(this.value).EncodeToBuffer(buffer[offset:])
}

func (this *Bytes) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}

	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	return &Bytes{
		placeholder: bool(codec.Bool(true).Decode(fields[0]).(codec.Bool)),
		value:       bytes.Clone(fields[1]),
	}
}

func (this *Bytes) Reset() {}

func (this *Bytes) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

func (this *Bytes) Print() {
	fmt.Println(*this)
	fmt.Println()
}

// Get the path type ID based on the path content.
func GetPathType(key string) uint8 {
	if strings.Contains(key, "/native/") {
		return 1
	}
	return 2
}

func (this *Bytes) StorageEncode(key string) []byte {
	// big int can take on arbitrary length, but will remove the leading zeros.
	// This isn't a problem in the original Ethereum implementation because
	// it has a fixed word size of 32 bytes. The leading zeros can be restored by copying the bytes
	// back to a 32 byte array.
	// However, we don't have a fixed word size, so we need to encode the length of the byte array.
	if GetPathType(key) == 1 {
		buffer, err := rlp.EncodeToBytes(new(big.Int).SetBytes(this.value))
		if err != nil {
			panic("Failed to encode bytes")
		}
		return buffer
	}

	buffer, err := rlp.EncodeToBytes(this.value)
	if err != nil {
		panic("Failed to encode bytes")
	}
	return buffer
}

func (this *Bytes) StorageDecode(key string, buffer []byte) interface{} {
	var buf []byte
	if err := rlp.DecodeBytes(buffer, &buf); err != nil {
		panic(err)
	}

	if GetPathType(key) == 1 {
		buf = ethcommon.BytesToHash(buf).Bytes()
	}

	return &Bytes{
		placeholder: true,
		value:       buf,
	}
}
