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

package committertest

import (
	"testing"

	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommitter "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/commutative"
	importer "github.com/arcology-network/storage-committer/importer"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

func TestPathMultiBatch(b *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	// bob := BobAccount()

	writeCache := NewWriteCacheWithAcounts(store, AliceAccount(), BobAccount())
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(acctTrans)
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)
	writeCache.Reset()

	keys := RandomKeys(0, 5)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(acctTrans)
	committer.Precommit([]uint32{0})
	committer.Commit(0)
	writeCache.Reset()

	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}
	}

	v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path))
	if v == nil || (v.(*deltaset.DeltaSet[string]).Length()) != uint64(len(keys)) {
		b.Error(err)
	}

	keys2 := RandomKeys(6, 8)
	for i := 0; i < len(keys2); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+keys2[i], noncommutative.NewInt64(int64(11))); err != nil {
			b.Error(err)
		}
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(acctTrans)
	committer.Precommit([]uint32{0})
	committer.Commit(0)
	writeCache.Reset()

	for i, k := range keys {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+k, new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}
	}

	for _, k := range keys2 {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/"+k, new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(11) {
			b.Error(err)
		}
	}

	v, _, err = writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path))
	if v == nil || (v.(*deltaset.DeltaSet[string]).Length()) != 7 {
		b.Error(err, v.(*deltaset.DeltaSet[string]).Length())
	}

}
