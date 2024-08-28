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
	"testing"

	"github.com/arcology-network/common-lib/exp/orderedset"
	"github.com/arcology-network/common-lib/exp/slice"
)

func TestPath(t *testing.T) {
	/* Noncommutative Path Test*/

	meta := NewPath()
	inPath := meta.(*Path)

	inPath.SetSubPaths([]string{"e-01", "e-001", "e-002", "e-002"})
	inPath.SetAdded([]string{"+01", "+001", "+002", "+002"})
	inPath.InsertRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	// metaout, _, _ := inPath.Get()

	if !slice.EqualSet(inPath.Value().([]string), []string{"e-01", "e-001", "e-002"}) {
		t.Error("Error: Don't match!!", inPath.Value().([]string))
	}

	if v, ok := inPath.GetByIndex(0); !ok || v != "e-01" {
		t.Error("Error: Don't match!!", v)
	}

	if v, ok := inPath.GetByIndex(1); !ok || v != "e-001" {
		t.Error("Error: Don't match!!", v)
	}

	if v, ok := inPath.GetByIndex(2); !ok || v != "e-002" {
		t.Error("Error: Don't match!!", v)
	}

	if v, ok := inPath.GetByIndex(3); !ok || v != "+01" {
		t.Error("Error: Don't match!!", v)
	}

	if v, ok := inPath.GetByIndex(4); !ok || v != "+001" {
		t.Error("Error: Don't match!!", v)
	}

	if v, ok := inPath.GetByIndex(5); !ok || v != "+002" {
		t.Error("Error: Don't match!!", v)
	}

	if v, ok := inPath.GetByIndex(6); ok || v != "" {
		t.Error("Error: Don't match!!", v)
	}
}

// panic: interface conversion: interface {} is *orderedset.OrderedSet[string], not []string [recovered]

func TestCodecPathMeta(t *testing.T) {
	in := NewPath().(*Path)

	in.SetSubPaths([]string{"e-01", "e-001", "e-002", "e-002"})
	in.SetAdded([]string{"+01", "+001", "+002", "+002"})
	in.InsertRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})

	buffer := in.Encode()
	out := (&Path{}).Decode(buffer).(*Path)

	if !slice.EqualSet(out.Value().(*orderedset.OrderedSet[string]).Elements(), []string{"e-01", "e-001", "e-002"}) {
		t.Error("Error: Don't match!!")
	}

	if !slice.EqualSet(out.Updated().Elements(), []string{"+01", "+001", "+002"}) {
		t.Error("Error: Don't match!!", out.Updated())
	}

	if !slice.EqualSet(out.Removed(), []string{"-091", "-0092", "-092", "-097"}) {
		t.Error("Error: Don't match!!", out.Removed())
	}

	buffer = in.Encode()
	out = (&Path{}).Decode(buffer).(*Path)

	if !slice.EqualSet(out.Value().(*orderedset.OrderedSet[string]).Elements(), []string{"e-01", "e-001", "e-002"}) {
		t.Error("Error: Don't match!! Error: Should have gone!")
	}

	if !slice.EqualSet(out.Updated().Elements(), []string{"+01", "+001", "+002"}) {
		t.Error("Error: Don't match!!", out.Updated())
	}

	if !slice.EqualSet(out.Removed(), []string{"-091", "-0092", "-092", "-097"}) {
		t.Error("Error: Don't match!!", out.Removed())
	}

	in = in.New(in.Value().(*orderedset.OrderedSet[string]).Elements(), nil, nil, nil, nil).(*Path)

	out.Committed().Init()

	buffer = in.Encode()
	out = (&Path{}).Decode(buffer).(*Path)

	out.Commit()

	if !slice.EqualSet(out.Value().(*orderedset.OrderedSet[string]).Elements(), []string{"+01", "+001", "+002"}) {
		t.Error("Error: Don't match!! Error: Should have gone!", out.Value().([]string))
	}

	if !slice.EqualSet(out.Updated().Elements(), []string{}) {
		t.Error("Error: Don't match!!", out.Updated())
	}

	if !slice.EqualSet(out.Removed(), []string{}) {
		t.Error("Error: Don't match!!", out.Removed())
	}
}
