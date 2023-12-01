package commutative

import (
	"fmt"
	"testing"

	common "github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
)

func TestMeta(t *testing.T) {
	/* Noncommutative Path Test*/

	meta := NewPath()
	inPath := meta.(*Path)

	inPath.SetSubs([]string{"e-01", "e-001", "e-002", "e-002"})
	inPath.SetAdded([]string{"+01", "+001", "+002", "+002"})
	inPath.SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	metaout, _, _ := inPath.Get()

	if !common.EqualArray(inPath.Value().(*orderedset.OrderedSet).Keys(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!!")
	}

	if !common.EqualArray(inPath.Delta().(*PathDelta).Added(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!")
	}

	if !common.EqualArray(inPath.Delta().(*PathDelta).Removed(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
		t.Error("Error: Don't match!!")
	}

	fmt.Println(metaout)
}

func TestCodecPathMeta(t *testing.T) {
	in := NewPath().(*Path)

	in.SetSubs([]string{"e-01", "e-001", "e-002", "e-002"})
	in.SetAdded([]string{"+01", "+001", "+002", "+002"})
	in.SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	buffer := in.Encode()
	out := (&Path{}).Decode(buffer).(*Path)

	if !common.EqualArray(out.Value().(*orderedset.OrderedSet).Keys(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!!")
	}

	if !common.EqualArray(out.Delta().(*PathDelta).Added(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Added())
	}

	if !common.EqualArray(out.Delta().(*PathDelta).Removed(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Removed())
	}

	buffer = in.Encode()
	out = (&Path{}).Decode(buffer).(*Path)

	if !common.EqualArray(out.Value().(*orderedset.OrderedSet).Keys(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!! Error: Should have gone!")
	}

	if !common.EqualArray(out.Delta().(*PathDelta).Added(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Added())
	}

	if !common.EqualArray(out.Delta().(*PathDelta).Removed(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Removed())
	}

	in = in.New(in.Value(), nil, nil, nil, nil).(*Path)
	buffer = in.Encode()
	out = (&Path{}).Decode(buffer).(*Path)

	if !common.EqualArray(out.Value().(*orderedset.OrderedSet).Keys(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!! Error: Should have gone!")
	}

	if !common.EqualArray(out.Delta().(*PathDelta).Added(), []string{}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Added())
	}

	if !common.EqualArray(out.Delta().(*PathDelta).Removed(), []string{}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Removed())
	}
}
