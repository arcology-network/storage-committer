package noncommutative

import (
	"encoding/gob"
)

func init() {
	gob.Register(&Bigint{})
	gob.Register(&Bytes{})
	i64 := Int64(0)
	gob.Register(&i64)
	str := String("")
	gob.Register(&str)
}
