/*
 *   Copyright (c) 2023 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package commutative

import (
	"fmt"
	"testing"

	orderedset "github.com/arcology-network/common-lib/container/set"
	"github.com/arcology-network/common-lib/exp/slice"
)

func TestMeta(t *testing.T) {
	/* Noncommutative Path Test*/

	meta := NewPath()
	inPath := meta.(*Path)

	inPath.SetSubs([]string{"e-01", "e-001", "e-002", "e-002"})
	inPath.SetAdded([]string{"+01", "+001", "+002", "+002"})
	inPath.SetRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	metaout, _, _ := inPath.Get()

	if !slice.Equal(inPath.Value().(*orderedset.OrderedSet).Keys(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!!")
	}

	if !slice.Equal(inPath.Delta().(*PathDelta).Added(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!")
	}

	if !slice.Equal(inPath.Delta().(*PathDelta).Removed(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
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

	if !slice.Equal(out.Value().(*orderedset.OrderedSet).Keys(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!!")
	}

	if !slice.Equal(out.Delta().(*PathDelta).Added(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Added())
	}

	if !slice.Equal(out.Delta().(*PathDelta).Removed(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Removed())
	}

	buffer = in.Encode()
	out = (&Path{}).Decode(buffer).(*Path)

	if !slice.Equal(out.Value().(*orderedset.OrderedSet).Keys(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!! Error: Should have gone!")
	}

	if !slice.Equal(out.Delta().(*PathDelta).Added(), []string{"+01", "+001", "+002", "+002"}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Added())
	}

	if !slice.Equal(out.Delta().(*PathDelta).Removed(), []string{"-091", "-0092", "-092", "-092", "-097"}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Removed())
	}

	in = in.New(in.Value(), nil, nil, nil, nil).(*Path)
	buffer = in.Encode()
	out = (&Path{}).Decode(buffer).(*Path)

	if !slice.Equal(out.Value().(*orderedset.OrderedSet).Keys(), []string{"e-01", "e-001", "e-002", "e-002"}) {
		t.Error("Error: Don't match!! Error: Should have gone!")
	}

	if !slice.Equal(out.Delta().(*PathDelta).Added(), []string{}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Added())
	}

	if !slice.Equal(out.Delta().(*PathDelta).Removed(), []string{}) {
		t.Error("Error: Don't match!!", out.Delta().(*PathDelta).Removed())
	}
}
