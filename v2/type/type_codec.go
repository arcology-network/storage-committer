package ccurltype

import (
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
	commutative "github.com/arcology/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology/concurrenturl/v2/type/noncommutative"
)

// Wrappers for the type encoder / decoder
func ToBytes(value interface{}) []byte {
	if value == nil {
		return []byte{} // Deletion
	}
	encoded := value.(ccurlcommon.TypeInterface).EncodeCompact()
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
	case ccurlcommon.NoncommutativeString: // delta big int
		stringer := noncommutative.String("")
		return stringer.DecodeCompact(bytes)

	case ccurlcommon.NoncommutativeInt64:
		i64 := noncommutative.Int64(0)
		return i64.DecodeCompact(bytes)

	case ccurlcommon.NoncommutativeBigint: // big int pointer
		return (&noncommutative.Bigint{}).DecodeCompact(bytes)

	case ccurlcommon.NoncommutativeBytes: // big int pointer
		return (&noncommutative.Bytes{}).DecodeCompact(bytes)

	case ccurlcommon.CommutativeMeta: // Path
		return (&commutative.Meta{}).DecodeCompact(bytes)

	case ccurlcommon.CommutativeBalance: // delta big int
		return (&commutative.Balance{}).DecodeCompact(bytes)

	case ccurlcommon.CommutativeInt64: // delta int 64
		return (&commutative.Int64{}).DecodeCompact(bytes)
	}

	return nil
}

func (Decoder) Decode(bytes []byte, vType uint8) interface{} {
	if len(bytes) == 0 {
		return nil
	}

	switch vType {
	case ccurlcommon.NoncommutativeString: // delta big int
		stringer := noncommutative.String("")
		return stringer.Decode(bytes)

	case ccurlcommon.NoncommutativeInt64:
		i64 := noncommutative.Int64(0)
		return i64.Decode(bytes)

	case ccurlcommon.NoncommutativeBigint: // big int pointer
		return (&noncommutative.Bigint{}).Decode(bytes)

	case ccurlcommon.NoncommutativeBytes: // big int pointer
		return (&noncommutative.Bytes{}).Decode(bytes)

	case ccurlcommon.CommutativeMeta: // Path
		return (&commutative.Meta{}).Decode(bytes)

	case ccurlcommon.CommutativeBalance: // delta big int
		return (&commutative.Balance{}).Decode(bytes)

	case ccurlcommon.CommutativeInt64: // delta int 64
		return (&commutative.Int64{}).Decode(bytes)
	}

	return nil
}
