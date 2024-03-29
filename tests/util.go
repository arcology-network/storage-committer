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

// Generate a random account, testing only
package committertest

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	slice "github.com/arcology-network/common-lib/exp/slice"
	cache "github.com/arcology-network/eu/cache"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/interfaces"
	opadapter "github.com/arcology-network/storage-committer/op"
	platform "github.com/arcology-network/storage-committer/platform"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	rlp "github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
	"golang.org/x/exp/constraints"
)

func RandomAccount() string {
	var letters = []byte("abcdef0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 20)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	addr := hexutil.Encode(b)
	return addr
}

func AliceAccount() string {
	b := make([]byte, 20)
	slice.Fill(b, 10)
	return hexutil.Encode(b)
}

func BobAccount() string {
	b := make([]byte, 20)
	slice.Fill(b, 11)
	return hexutil.Encode(b)
}

func CarolAccount() string {
	b := make([]byte, 20)
	slice.Fill(b, 12)
	return hexutil.Encode(b)
}

func RandomAccounts(n int) []string {
	accounts := make([]string, n)
	for i := range n {
		b := sha3.Sum256([]byte(fmt.Sprintf("%v", rand.Intn(1000000))))
		accounts[i] = hexutil.Encode(b[:20])
	}
	return accounts
}

func rlpEncoder(args ...interface{}) []byte {
	encoded, err := rlp.EncodeToBytes(args)
	if err != nil {
		log.Fatal("Error encoding data:", err)
	}
	return encoded
}

func RandomKey[T constraints.Integer](seed T) string {
	buf := sha3.Sum256([]byte(fmt.Sprint(seed)))
	return hexutil.Encode(buf[:20])
}

func RandomKeys[T constraints.Integer](s0, s1 T) []string {
	keys := make([]string, s1-s0)
	for i := range keys {
		keys[i] = RandomKey(s0 + T(i))
	}
	return keys
}

func NewWriteCacheWithAcounts(store interfaces.Datastore, accounts ...string) *cache.WriteCache {
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	for i := range accounts {
		if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, accounts[i]); err != nil { // NewAccount account structure {
			fmt.Println(err)
		}
	}
	return writeCache
}

func verifierEthMerkle(roothash [32]byte, acct string, key string, store interfaces.Datastore, t *testing.T) {
	// roothash := store.(*stgproxy.StoreProxy).EthStore().Root()                               // Get the proof provider by a root hash.
	ethdb := store.(*stgproxy.StoreProxy).EthStore().EthDB()                         // Get the proof provider by a root hash.
	provider, err := ethstg.NewMerkleProofCache(2, ethdb).GetProofProvider(roothash) // Initiate the proof cache, maximum 2 blocks
	if err != nil {
		t.Fatal(err)
	}

	// Verify Bob's stgproxy for a big int value.
	bobAddr := ethcommon.BytesToAddress([]byte(hexutil.MustDecode(acct)))
	accountResult, err := provider.GetProof(bobAddr, []string{key})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	opProof := opadapter.Convertible(*accountResult).New() // Convert Bob's proof to OP format and verify.
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}
}
