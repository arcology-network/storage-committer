package ccdb

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/noncommutative"
)

// Wrappers for the type encoder / decoder
func ToBytes(value interface{}) []byte {
	if value == nil {
		return []byte{} // Deletion
	}
	encoded := value.(ccurlcommon.TypeInterface).Encode()
	encoded = append(encoded, value.(ccurlcommon.TypeInterface).TypeID())
	return encoded
}

func FromBytes(bytes []byte) interface{} {
	if len(bytes) == 0 {
		return nil
	}
	return Decoder{}.Decode(bytes[0:len(bytes)-1], bytes[len(bytes)-1])
}

type Decoder struct{}

func (Decoder) Decode(bytes []byte, vType uint8) interface{} {
	if len(bytes) == 0 {
		return nil
	}

	switch vType {
	case noncommutative.STRING: // delta big int
		stringer := noncommutative.String("")
		return stringer.Decode(bytes)

	case noncommutative.BIGINT: // big int pointer
		return (&noncommutative.Bigint{}).Decode(bytes)

	case noncommutative.BYTES: // big int pointer
		return (&noncommutative.Bytes{}).Decode(bytes)

	case commutative.PATH: // Path
		return (&commutative.Path{}).Decode(bytes)

	case commutative.INT64: // delta int 64
		return (&commutative.Int64{}).Decode(bytes)

	case commutative.UINT64: // delta int 64
		return (&commutative.Uint64{}).Decode(bytes)

	case commutative.UINT256: // delta big int
		return (&commutative.U256{}).Decode(bytes)

	case noncommutative.INT64:
		i64 := noncommutative.Int64(0)
		return i64.Decode(bytes)
	}

	return nil
}

// func (Decoder) Decode(bytes []byte, vType uint8) interface{} {
// 	if len(bytes) == 0 {
// 		return nil
// 	}

// 	switch vType {
// 	case noncommutative.STRING: // delta big int
// 		stringer := noncommutative.String("")
// 		return stringer.Decode(bytes)

// 	case noncommutative.INT64:
// 		i64 := noncommutative.Int64(0)
// 		return i64.Decode(bytes)

// 	case noncommutative.BIGINT: // big int pointer
// 		return (&noncommutative.Bigint{}).Decode(bytes)

// 	case noncommutative.BYTES: // big int pointer
// 		return (&noncommutative.Bytes{}).Decode(bytes)

// 	case commutative.PATH: // Path
// 		return (&commutative.Path{}).Decode(bytes)

// 	case commutative.UINT256: // delta big int
// 		return (&commutative.U256{}).Decode(bytes)

// 	case commutative.UINT64: // delta int 64
// 		return (&commutative.Uint64{}).Decode(bytes)

// 	case commutative.INT64: // delta int 64
// 		return (&commutative.Int64{}).Decode(bytes)
// 	}

// 	return nil
// }
