package concurrenturl

import (
	"reflect"

	ccurlcommon "github.com/arcology/concurrenturl/common"
	commutative "github.com/arcology/concurrenturl/type/commutative"
	noncommutative "github.com/arcology/concurrenturl/type/noncommutative"
)

type Decoder struct{}

func (Decoder) Decode(bytes []byte, dtype uint8) interface{} {
	switch dtype {
	case uint8(reflect.Kind(ccurlcommon.NoncommutativeMeta)): // Path
		return (&commutative.Meta{}).Decode(bytes)

	case uint8(reflect.Kind(ccurlcommon.NoncommutativeString)): // delta big int
		stringer := noncommutative.String("")
		return stringer.Decode(bytes)

	case uint8(reflect.Kind(ccurlcommon.CommutativeBalance)): // delta big int
		return (&commutative.Balance{}).Decode(bytes)

	case uint8(reflect.Kind(ccurlcommon.NoncommutativeBigint)): // big int pointer
		// v := noncommutative.Bigint(codec.Bigint{}.Decode(bytes))
		// return &v
		return (&noncommutative.Bigint{}).Decode(bytes)
	}
	return nil
}
