/*
 *   Copyright (c) 2024 Arcology Network

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

package storage

import (
	"testing"

	"github.com/arcology-network/storage-committer/importer"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	"github.com/arcology-network/storage-committer/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func TestGetPathType(t *testing.T) {
	_0 := &importer.DeltaSequence{
		Account:   "0",
		Finalized: univalue.NewUnivalue(0, "0x1", 1, 2, 3, noncommutative.NewString("0x"), nil),
		Transitions: []*univalue.Univalue{
			univalue.NewUnivalue(2, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
			univalue.NewUnivalue(2, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
		},
	}

	_1 := &importer.DeltaSequence{
		Account:   "1",
		Finalized: univalue.NewUnivalue(1, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
		Transitions: []*univalue.Univalue{
			univalue.NewUnivalue(2, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
			univalue.NewUnivalue(2, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
		},
	}

	_2 := &importer.DeltaSequence{
		Account:   "2",
		Finalized: univalue.NewUnivalue(2, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
		Transitions: []*univalue.Univalue{
			univalue.NewUnivalue(2, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
			univalue.NewUnivalue(2, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
		},
	}

	_3 := &importer.DeltaSequence{
		Account:   "3",
		Finalized: univalue.NewUnivalue(3, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
		Transitions: []*univalue.Univalue{
			univalue.NewUnivalue(2, "0x", 1, 2, 3, noncommutative.NewString("0x"), nil),
		},
	}

	acctUpdates := []*AccountUpdate{
		{
			Key:  "key0",
			Addr: ethcommon.Address{},
			Seqs: []*importer.DeltaSequence{
				_0,
				_1,
			},
		},
		{
			Key:  "key1",
			Addr: ethcommon.Address{},
			Seqs: []*importer.DeltaSequence{
				_2,
				_3,
			},
		},
	}

	k, vs := AccountUpdates(acctUpdates).KVs()
	if len(k) != 4 || len(vs) != 4 {
		t.Errorf("TestGetPathType failed")
	}
}
