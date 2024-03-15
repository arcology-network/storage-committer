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
	"bytes"
	"fmt"
	"testing"

	cache "github.com/arcology-network/eu/cache"
	// cache "github.com/arcology-network/common-lib/cache"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommitter "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	importer "github.com/arcology-network/storage-committer/importer"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	platform "github.com/arcology-network/storage-committer/platform"
	storage "github.com/arcology-network/storage-committer/storage"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

func TestNewCommitter(t *testing.T) {
	// store := chooseDataStore()
	store := storage.NewHybirdStore()

	alice := AliceAccount()
	committer := stgcommitter.NewStorageCommitterV2(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := writeCache.Export(importer.Sorter)
	acctTrans := univalue.Univalues(slice.Clone(raw)).To(importer.IPTransition{})

	committer.Import(acctTrans)
	committer.Precommit([]uint32{1})
	committer.Commit().Clear()
	committer.Clear()
	writeCache.Reset(writeCache)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), noncommutative.NewBytes([]byte{1, 2, 3})); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), noncommutative.NewBytes([]byte{2, 2, 3})); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}

	raw = writeCache.Export(importer.Sorter)
	acctTrans = univalue.Univalues(slice.Clone(raw)).To(importer.IPTransition{})

	committer.Import(acctTrans)
	committer.Precommit([]uint32{1})
	committer.Commit().Clear()
	writeCache.Reset(writeCache)

	outV, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{1, 2, 3}) {
		t.Error("Error: The path should exists", outV)
	}

	outV, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{2, 2, 3}) {
		t.Error("Error: The path should exists", outV)
	}

	arr := [][]int{
		{1, 2, 3},
		{4, 5, 6},
	}

	remover := func(arr [][]int) {
		arr[0] = arr[0][:0]
	}
	fmt.Println(arr)
	remover(arr)
	fmt.Println(arr)
}
