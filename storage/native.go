package storage

import (
	commutative "github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
)

type Codec struct {
	ID uint8
}

func (Codec) Encode(_ string, value interface{}) []byte {
	if value == nil {
		return []byte{} // Deletion
	}

	encoded := value.(interfaces.Type).Encode()
	encoded = append(encoded, value.(interfaces.Type).TypeID())
	return encoded
}

func (this Codec) Decode(buffer []byte, _ any) interface{} {
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

// func (Codec) Size(v interface{}) uint64 {
// 	switch v.(type) {
// 	case int64: // delta big int
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
