package univalue

import (
	"github.com/arcology-network/common-lib/codec"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/noncommutative"
)

// Wrappers for the type encoder / decoder
func ToBytes(value interface{}) []byte {
	if value == nil {
		return []byte{} // Deletion
	}
	encoded := value.(ccurlcommon.TypeInterface).Value().(codec.Encodeable).Encode()
	encoded = append(encoded, value.(ccurlcommon.TypeInterface).TypeID())
	return encoded
}

func FromBytes(bytes []byte) interface{} {
	if len(bytes) == 0 {
		return nil
	}
	return Decoder{}.DecodeCompact(bytes[0:len(bytes)-1], bytes[len(bytes)-1])
}

type Decoder struct{}

func (Decoder) DecodeCompact(bytes []byte, vType uint8) interface{} {
	if len(bytes) == 0 {
		return nil
	}

	switch vType {
	case ccurlcommon.NonCommutative{}.String(): // delta big int
		stringer := noncommutative.String("")
		return stringer.DecodeCompact(bytes)

	case ccurlcommon.NonCommutative{}.Int64():
		i64 := noncommutative.Int64(0)
		return i64.DecodeCompact(bytes)

	case ccurlcommon.NonCommutative{}.Bigint(): // big int pointer
		return (&noncommutative.Bigint{}).DecodeCompact(bytes)

	case ccurlcommon.NonCommutative{}.Bytes(): // big int pointer
		return (&noncommutative.Bytes{}).DecodeCompact(bytes)

	case ccurlcommon.Commutative{}.Path(): // Path
		return (&commutative.Path{}).DecodeCompact(bytes)

	case ccurlcommon.Commutative{}.Uint256(): // delta big int
		return (&commutative.U256{}).DecodeCompact(bytes)

	case ccurlcommon.Commutative{}.Int64(): // delta int 64
		return (&commutative.Int64{}).DecodeCompact(bytes)
	}

	return nil
}

func (Decoder) Decode(bytes []byte, vType uint8) interface{} {
	if len(bytes) == 0 {
		return nil
	}

	switch vType {
	case ccurlcommon.NonCommutative{}.String(): // delta big int
		stringer := noncommutative.String("")
		return stringer.Decode(bytes)

	case ccurlcommon.NonCommutative{}.Int64():
		i64 := noncommutative.Int64(0)
		return i64.Decode(bytes)

	case ccurlcommon.NonCommutative{}.Bigint(): // big int pointer
		return (&noncommutative.Bigint{}).Decode(bytes)

	case ccurlcommon.NonCommutative{}.Bytes(): // big int pointer
		return (&noncommutative.Bytes{}).Decode(bytes)

	case ccurlcommon.Commutative{}.Path(): // Path
		return (&commutative.Path{}).Decode(bytes)

	case ccurlcommon.Commutative{}.Uint256(): // delta big int
		return (&commutative.U256{}).Decode(bytes)

	case ccurlcommon.Commutative{}.Uint64(): // delta int 64
		return (&commutative.Uint64{}).Decode(bytes)

	case ccurlcommon.Commutative{}.Int64(): // delta int 64
		return (&commutative.Int64{}).Decode(bytes)
	}

	return nil
}
