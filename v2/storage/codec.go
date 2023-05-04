package ccdb

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/noncommutative"
)

type Codec struct {
	ID uint8
}

func (Codec) Encode(value interface{}) []byte {
	if value == nil {
		return []byte{} // Deletion
	}
	encoded := value.(ccurlcommon.TypeInterface).Encode()
	encoded = append(encoded, value.(ccurlcommon.TypeInterface).TypeID())
	return encoded
}

func (this Codec) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return nil
	}

	if len(buffer) == 0 || this.ID == 0 {
		this.ID = buffer[len(buffer)-1]
		buffer = buffer[0 : len(buffer)-1]
	}

	switch this.ID {
	case noncommutative.STRING: // delta big int
		stringer := noncommutative.String("")
		return stringer.Decode(buffer)

	case noncommutative.BIGINT: // big int pointer
		return (&noncommutative.Bigint{}).Decode(buffer)

	case noncommutative.BYTES: // big int pointer
		return (&noncommutative.Bytes{}).Decode(buffer)

	case commutative.PATH: // Path
		return (&commutative.Path{}).Decode(buffer)

	case commutative.INT64: // delta int 64
		return (&commutative.Int64{}).Decode(buffer)

	case commutative.UINT64: // delta int 64
		return (&commutative.Uint64{}).Decode(buffer)

	case commutative.UINT256: // delta big int
		return (&commutative.U256{}).Decode(buffer)

	case noncommutative.INT64:
		i64 := noncommutative.Int64(0)
		return i64.Decode(buffer)
	}

	return nil
}

// func (Decoder) Decode(buffer []byte, vType uint8) interface{} {
// 	if len(buffer) == 0 {
// 		return nil
// 	}

// 	switch vType {
// 	case noncommutative.STRING: // delta big int
// 		stringer := noncommutative.String("")
// 		return stringer.Decode(buffer)

// 	case noncommutative.INT64:
// 		i64 := noncommutative.Int64(0)
// 		return i64.Decode(buffer)

// 	case noncommutative.BIGINT: // big int pointer
// 		return (&noncommutative.Bigint{}).Decode(buffer)

// 	case noncommutative.BYTES: // big int pointer
// 		return (&noncommutative.Bytes{}).Decode(buffer)

// 	case commutative.PATH: // Path
// 		return (&commutative.Path{}).Decode(buffer)

// 	case commutative.UINT256: // delta big int
// 		return (&commutative.U256{}).Decode(buffer)

// 	case commutative.UINT64: // delta int 64
// 		return (&commutative.Uint64{}).Decode(buffer)

// 	case commutative.INT64: // delta int 64
// 		return (&commutative.Int64{}).Decode(buffer)
// 	}

// 	return nil
// }
