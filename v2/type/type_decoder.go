package urltype

import (
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	commutative "github.com/HPISTechnologies/concurrenturl/v2/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/v2/type/noncommutative"
)

type Decoder struct{}

func (Decoder) Decode(bytes []byte, vType uint8) interface{} {
	if len(bytes) == 0 {
		return nil
	}

	switch vType {
	case ccurlcommon.NoncommutativeString: // delta big int
		stringer := noncommutative.String("")
		return stringer.Decode(bytes)

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
