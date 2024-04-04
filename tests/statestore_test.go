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
	"testing"

	"github.com/arcology-network/common-lib/exp/slice"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	"github.com/arcology-network/storage-committer/interfaces"

	statestore "github.com/arcology-network/storage-committer"
	importer "github.com/arcology-network/storage-committer/committer/importer"
	stgcomm "github.com/arcology-network/storage-committer/common"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

func commitToStateStore(store interfaces.Datastore, t *testing.T) {
	alice := AliceAccount()
	stateStore := statestore.NewStateStore(store)

	if _, err := adaptorcommon.CreateNewAccount(stgcomm.SYSTEM, alice, stateStore.ShardedWriteCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := univalue.Univalues(slice.Clone(stateStore.Export(importer.Sorter))).To(importer.IPTransition{})

	stateStore.Import(acctTrans).Precommit([]uint32{stgcomm.SYSTEM})
	stateStore.Commit(stgcomm.SYSTEM)
	stateStore.Clear()

	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), noncommutative.NewBytes([]byte{1, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), noncommutative.NewBytes([]byte{2, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}
	acctTrans = univalue.Univalues(slice.Clone(stateStore.Export(importer.Sorter))).To(importer.IPTransition{})

	// stateStore.Import(acctTrans)
	stateStore = statestore.NewStateStore(store)
	stateStore.Import(acctTrans).Precommit([]uint32{1})
	stateStore.Commit(2).Clear()

	outV, _, _ := stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{1, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{2, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 45, 67}) {
		t.Error("Error: The path should exist", outV)
	}

	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 199, 199})); err != nil {
		t.Error(err)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 199, 199}) {
		t.Error("Error: The path should exist", outV)
	}

	// Delete the entry
	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), nil); err != nil {
		t.Error(err)
	}

	// Delete the entry
	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(77), noncommutative.NewBytes([]byte{77, 77})); err != nil {
		t.Error(err)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV != nil {
		t.Error("Error: The path should not exist", outV)
	}

	acctTrans = univalue.Univalues(slice.Clone(stateStore.Export(importer.Sorter))).To(importer.IPTransition{})
	// stateStore = statestore.NewStateStore(store)
	stateStore.Import(acctTrans).Precommit([]uint32{1})
	stateStore.Commit(2).Clear()

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV != nil {
		t.Error("Error: The path should not exist", outV)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(77), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{77, 77}) {
		t.Error("Error: The path should not exist", outV)
	}
}

func TestCommitToStatStore(t *testing.T) {
	commitToStateStore(stgproxy.NewStoreProxy().EnableCache(), t) // Use cache
	// commitToStateStore(stgproxy.NewStoreProxy().DisableCache(), t)
}

func TestAsyncCommitToStateStore(t *testing.T) {
	alice := AliceAccount()
	store := stgproxy.NewStoreProxy().EnableCache()
	stateStore := statestore.NewStateStore(store)

	if _, err := adaptorcommon.CreateNewAccount(stgcomm.SYSTEM, alice, stateStore.ShardedWriteCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := univalue.Univalues(slice.Clone(stateStore.Export(importer.Sorter))).To(importer.IPTransition{})

	stateStore.Import(acctTrans).Precommit([]uint32{stgcomm.SYSTEM})
	stateStore.Commit(stgcomm.SYSTEM)
	stateStore.Clear()

	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), noncommutative.NewBytes([]byte{1, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), noncommutative.NewBytes([]byte{2, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}
	acctTrans = univalue.Univalues(slice.Clone(stateStore.Export(importer.Sorter))).To(importer.IPTransition{})

	// stateStore.Import(acctTrans)
	stateStore = statestore.NewStateStore(store)
	stateStore.Import(acctTrans).Precommit([]uint32{1})
	stateStore.Commit(2).Clear()

	outV, _, _ := stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{1, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{2, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 45, 67}) {
		t.Error("Error: The path should exist", outV)
	}

	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 199, 199})); err != nil {
		t.Error(err)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 199, 199}) {
		t.Error("Error: The path should exist", outV)
	}

	// Delete the entry
	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), nil); err != nil {
		t.Error(err)
	}

	// Delete the entry
	if _, err := stateStore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(77), noncommutative.NewBytes([]byte{77, 77})); err != nil {
		t.Error(err)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV != nil {
		t.Error("Error: The path should not exist", outV)
	}

	acctTrans = univalue.Univalues(slice.Clone(stateStore.Export(importer.Sorter))).To(importer.IPTransition{})
	// stateStore = statestore.NewStateStore(store)
	stateStore.Import(acctTrans).Precommit([]uint32{1})
	stateStore.Commit(2).Clear()

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV != nil {
		t.Error("Error: The path should not exist", outV)
	}

	outV, _, _ = stateStore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(77), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{77, 77}) {
		t.Error("Error: The path should not exist", outV)
	}
}
