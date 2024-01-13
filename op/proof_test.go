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
package opadapter

import (
	"testing"

	committercommon "github.com/arcology-network/concurrenturl/common"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	ccurlstorage "github.com/arcology-network/concurrenturl/storage"
	storage "github.com/arcology-network/concurrenturl/storage"
	tests "github.com/arcology-network/concurrenturl/tests"
	cache "github.com/arcology-network/eu/cache"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethmpt "github.com/ethereum/go-ethereum/trie"
)

func TestGetProofAPI(t *testing.T) {
	store := ccurlstorage.NewParallelEthMemDataStore()
	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	bob := tests.BobAccount()
	writeCache.CreateNewAccount(committercommon.SYSTEM, bob)
	writeCache.FlushToDataSource(store)

	/* Bob updates */
	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124"))
	writeCache.FlushToDataSource(store)
	roothash := store.Root()

	proofProvider, err := storage.NewProofProvider(store.EthDB(), roothash)
	if err != nil {
		t.Error(err)
	}

	if data, err := proofProvider.DataStore.IsAccountProvable(bob); len(data) == 0 || err != nil {
		t.Error(err)
	}
	bobAddr := ethcommon.BytesToAddress(hexutil.MustDecode(bob))

	bobCache, _ := proofProvider.DataStore.GetAccount(bobAddr, &ethmpt.AccessListCache{})
	if _, _, err := bobCache.IsStorageProvable("0x0000000000000000000000000000000000000000000000000000000000000000"); err != nil {
		t.Error(err)
	}

	accountResult, err := proofProvider.GetProof(bobAddr, []string{string("0x0000000000000000000000000000000000000000000000000000000000000000")})
	err = accountResult.Validate(roothash)
	if err != nil {
		t.Error(err)
	}

	res := Convertible(*accountResult).New()
	if err := res.Verify(roothash); err != nil {
		t.Error(err)
	}
}
