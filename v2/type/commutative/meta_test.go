package commutative

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/HPISTechnologies/common-lib/datacompression"
)

func TestMeta(t *testing.T) {
	/* Noncommutative Path Test*/
	alice := datacompression.RandomAccount()
	inPath, _ := NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	inPath.(*Meta).SetAdded([]string{"blcc://eth1.0/account/0", "blcc://eth1.0/account/1"})
	inPath.(*Meta).SetRemoved([]string{"blcc://eth1.0/account/2", "blcc://eth1.0/account/3"})

	buffer := inPath.(*Meta).Encode()
	out := (&Meta{}).Decode(buffer).(*Meta)

	reflect.DeepEqual(inPath, out)

	fmt.Println("Path Encoded size :", len(inPath.(*Meta).Encode()))
	fmt.Println("Balance Encoded Compact size :", len(inPath.(*Meta).Encode()))
}

func TestCodecPathMeta(t *testing.T) {
	/* Commutative Int64 Test */
	in, _ := NewMeta("blcc://eth1.0/account/0x12345456/")
	buffer := in.(*Meta).Encode()
	out := (&Meta{}).Decode(buffer).(*Meta)

	if !reflect.DeepEqual(in.(*Meta).PeekKeys(), out.PeekKeys()) ||
		!reflect.DeepEqual(in.(*Meta).PeekAdded(), out.PeekAdded()) ||
		!reflect.DeepEqual(in.(*Meta).PeekRemoved(), out.PeekRemoved()) ||
		!reflect.DeepEqual(in.(*Meta).Composite(), out.Composite()) {
		t.Error()
	}
}
