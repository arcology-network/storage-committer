package commutative

import (
	"encoding/gob"
)

func init() {
	gob.Register(&Balance{})
	gob.Register(&Meta{})
	gob.Register(&Int64{})
}
