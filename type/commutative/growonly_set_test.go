/*
 *   Copyright (c) 2025 Arcology Network

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
	"reflect"
	"testing"
)

func TestGrowonlySet(t *testing.T) {
	set := NewGrowOnlySet(
		func(v []byte) uint64 { return uint64(len(v)) },
		func(v []byte, bf []byte) int { return copy(bf, v) },
		func(v []byte) []byte { return v },
		func(a, b []byte) bool { return reflect.DeepEqual(a, b) },
	)

	if _, _, _, _, err := set.Set([]byte{1, 2, 3, 4, 5}, nil); err != nil {
		t.Errorf("Failed to set value in GrowOnlySet: %v", err)
	}

	if _, _, _, _, err := set.Set([]byte{7, 8, 9, 10, 11}, nil); err != nil {
		t.Errorf("Failed to set value in GrowOnlySet: %v", err)
	}

	v, _, _ := set.Get()
	if value := v.([][]byte); !reflect.DeepEqual(value, [][]byte{{1, 2, 3, 4, 5}, {7, 8, 9, 10, 11}}) {
		t.Errorf("GrowOnlySet value mismatch: got %v, want %v", value, [][]byte{{1, 2, 3, 4, 5}, {7, 8, 9, 10, 11}})
	}

	buffer := set.Encode()

	set2 := NewGrowOnlySet(
		func(v []byte) uint64 { return uint64(len(v)) },
		func(v []byte, bf []byte) int { return copy(bf, v) },
		func(v []byte) []byte { return v },
		func(a, b []byte) bool { return reflect.DeepEqual(a, b) },
	)

	out := set2.Decode(buffer).(*GrowOnlySet[[]byte])
	v2, _, _ := out.Get()

	if !reflect.DeepEqual(v2, [][]byte{{1, 2, 3, 4, 5}, {7, 8, 9, 10, 11}}) {
		t.Errorf("GrowOnlySet value mismatch: got %v, want %v", v2, [][]byte{{1, 2, 3, 4, 5}, {7, 8, 9, 10, 11}})
	}
}
