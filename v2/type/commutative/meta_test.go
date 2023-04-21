package commutative

import (
	"fmt"
	"testing"

	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/datacompression"
)

func TestMeta(t *testing.T) {
	/* Noncommutative Path Test*/
	alice := datacompression.RandomAccount()
	meta, _ := NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	inPath := meta.(*Meta)

	inPath.committedKeys = ([]string{"0", "1", "2", "3"})
	inPath.added = ([]string{"5", "6"})
	inPath.removed = []string{"2", "3"}

	inPath.SetCommittedKeys([]string{"e-01", "e-001", "e-002", "e-002"})
	inPath.SetAdded([]string{"+01", "+001", "+002", "+002"})
	inPath.SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	meta, _, _ = inPath.Get(nil)
	fmt.Println(meta)
}

func TestCodecPathMeta(t *testing.T) {
	// /* Commutative Int64 Test */
	in, _ := NewMeta("blcc://eth1.0/account/0x12345456/")

	in.(*Meta).SetCommittedKeys([]string{"e-01", "e-001", "e-002", "e-002"})
	in.(*Meta).SetAdded([]string{"+01", "+001", "+002", "+002"})
	in.(*Meta).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	buffer := in.(*Meta).Encode()
	out := (&Meta{}).Decode(buffer).(*Meta)

	if !common.EqualArray(in.(*Meta).Added(), out.Added()) ||
		!common.EqualArray(in.(*Meta).Removed(), out.Removed()) {
		t.Error("Error: Don't match!!")
	}
}
