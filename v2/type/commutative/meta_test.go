package commutative

import (
	"fmt"
	"testing"

	common "github.com/arcology-network/common-lib/common"
)

func TestMeta(t *testing.T) {
	/* Noncommutative Path Test*/

	meta := NewMeta()
	inPath := meta.(*Meta)

	inPath.SetSubDirs([]string{"e-01", "e-001", "e-002", "e-002"})
	inPath.SetAdded([]string{"+01", "+001", "+002", "+002"})
	inPath.SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	meta, _, _ = inPath.Get(nil)

	if !common.EqualArray(inPath.SubDirs(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!!")
	}

	if !common.EqualArray(inPath.Added(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!")
	}

	if !common.EqualArray(inPath.Removed(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
		t.Error("Error: Don't match!!")
	}

	fmt.Println(meta)
}

func TestCodecPathMeta(t *testing.T) {
	// /* Commutative Int64 Test */
	in := NewMeta()

	in.(*Meta).SetSubDirs([]string{"e-01", "e-001", "e-002", "e-002"})
	in.(*Meta).SetAdded([]string{"+01", "+001", "+002", "+002"})
	in.(*Meta).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	buffer := in.(*Meta).Encode()
	out := (&Meta{}).Decode(buffer).(*Meta)

	if common.EqualArray(out.SubDirs(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Should have gone!")
	}

	if !common.EqualArray(out.Added(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!")
	}

	if !common.EqualArray(out.Removed(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
		t.Error("Error: Don't match!!")
	}

}
