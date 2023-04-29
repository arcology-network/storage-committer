package commutative

import (
	"fmt"
	"testing"

	common "github.com/arcology-network/common-lib/common"
)

func TestMeta(t *testing.T) {
	/* Noncommutative Path Test*/

	meta := NewPath()
	inPath := meta.(*Path)

	inPath.SetSubDirs([]string{"e-01", "e-001", "e-002", "e-002"})
	inPath.SetAdded([]string{"+01", "+001", "+002", "+002"})
	inPath.SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	meta, _, _ = inPath.Get()

	if !common.EqualArray(inPath.PeekValue(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!!")
	}

	if !common.EqualArray(inPath.PeekAdded(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!")
	}

	if !common.EqualArray(inPath.PeekRemoved(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
		t.Error("Error: Don't match!!")
	}

	fmt.Println(meta)
}

func TestCodecPathMeta(t *testing.T) {
	// /* Commutative Int64 Test */
	in := NewPath()

	in.(*Path).SetSubDirs([]string{"e-01", "e-001", "e-002", "e-002"})
	in.(*Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	in.(*Path).SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	buffer := in.(*Path).Encode()
	out := (&Path{}).Decode(buffer).(*Path)

	if common.EqualArray(out.PeekValue(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Should have gone!")
	}

	if !common.EqualArray(out.PeekAdded(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!")
	}

	if !common.EqualArray(out.PeekRemoved(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
		t.Error("Error: Don't match!!")
	}

}
