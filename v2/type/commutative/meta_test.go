package commutative

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/datacompression"
)

func TestMeta(t *testing.T) {
	/* Noncommutative Path Test*/
	alice := datacompression.RandomAccount()
	meta, _ := NewMeta("blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil)
	inPath := meta.(*Meta)

	inPath.keys = ([]string{"0", "1", "2", "3"})
	inPath.added = ([]string{"5", "6"})
	inPath.removed = []string{"2", "3"}

	meta, _, _ = inPath.Get("blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil)
	fmt.Println(meta)

	//inPath.Get()

	// buffer := inPath.(*Meta).Encode()
	// out := (&Meta{}).Decode(buffer).(*Meta)

	// reflect.DeepEqual(inPath, out)

	// fmt.Println("Path Encoded size :", len(inPath.(*Meta).Encode()))
	// fmt.Println("U256 Encoded Compact size :", len(inPath.(*Meta).Encode()))
}
func TestCodecPathMeta(t *testing.T) {
	/* Commutative Int64 Test */
	in, _ := NewMeta("blcc://eth1.0/account/0x12345456/", nil)
	buffer := in.(*Meta).Encode()
	out := (&Meta{}).Decode(buffer).(*Meta)

	if !reflect.DeepEqual(in.(*Meta).KeyView(), out.KeyView()) ||
		!reflect.DeepEqual(in.(*Meta).Added(), out.Added()) ||
		!reflect.DeepEqual(in.(*Meta).Removed(), out.Removed()) ||
		!reflect.DeepEqual(in.(*Meta).Composite(), out.Composite()) {
		t.Error()
	}
}
