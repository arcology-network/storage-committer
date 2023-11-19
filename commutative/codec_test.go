package commutative

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/holiman/uint256"
)

func TestCommutativeCodec(t *testing.T) {
	/* Noncommutative Path Test*/
	v := NewBoundedU256(uint256.NewInt(1), uint256.NewInt(400))
	v.SetValue(*uint256.NewInt(37))

	buffer := v.StorageEncode()
	output := (&U256{}).StorageDecode(buffer)

	if !reflect.DeepEqual(v, output) {
		fmt.Println("Error: Missmatched")
	}

	v = NewBoundedUint64(uint64(1), uint64(400))
	v.SetValue(uint64(37))

	buffer = v.StorageEncode()
	output = (&Uint64{}).StorageDecode(buffer)

	if !reflect.DeepEqual(v, output) {
		fmt.Println("Error: Missmatched")
	}
}
