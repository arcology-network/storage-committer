package univalue

import (
	"encoding/gob"
)

func init() {
	gob.Register(&Univalue{})
	// gob.Register(Univalues{})
}
