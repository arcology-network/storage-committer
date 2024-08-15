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

package ccstorage

// import (
// 	"fmt"
// 	"strings"
// 	"time"

// 	common "github.com/arcology-network/common-lib/common"
// 	mempool "github.com/arcology-network/common-lib/mempool"
// 	merkle "github.com/arcology-network/common-lib/merkle"
// 	stgcommcommon "github.com/arcology-network/common-lib/types/storage/common"
// 	"github.com/arcology-network/storage-committer/interfaces"
// 	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
// )

// const (
// 	concurrency = 8
// )

// type AccountMerkle struct {
// 	branches   uint32
// 	merkles    map[string]*merkle.Merkle
// 	platform   interfaces.Platform
// 	nodePool   *mempool.Mempool
// 	merklePool *mempool.Mempool
// 	encoder    func(...interface{}) []byte
// }

// func NewAccountMerkle(platform interfaces.Platform, encoder func(...interface{}) []byte, hashFunc func([]byte) []byte) *AccountMerkle {
// 	am := &AccountMerkle{
// 		branches: 16,
// 		merkles:  make(map[string]*merkle.Merkle),
// 		platform: platform,
// 		nodePool: mempool.NewMempool("node", func() interface{} {
// 			return merkle.NewNode()
// 		}),
// 		merklePool: mempool.NewMempool("merkle", func() interface{} {
// 			return merkle.NewMerkle(16, merkle.Concatenator{}, merkle.Keccak256{})
// 		}),

// 		encoder: encoder,
// 	}
// 	return am
// }

// func (this *AccountMerkle) Clear() {
// 	this.merkles = make(map[string]*merkle.Merkle)
// 	this.nodePool.ReclaimRecursive()
// 	this.merklePool.ReclaimRecursive()
// }

// func (this *AccountMerkle) GetMerkles() *map[string]*merkle.Merkle {
// 	return &this.merkles
// }

// // Insert to the merkle tree
// func (this *AccountMerkle) Import(transitions []*univalue.Univalue) {
// 	offset := stgcommcommon.ETH10_ACCOUNT_PREFIX_LENGTH
// 	for _, v := range transitions {
// 		path := *v.GetPath()
// 		pos := strings.Index(path[offset:], "/")
// 		if pos >= 0 {
// 			acct := path[offset : pos+offset]
// 			if this.merkles[acct] == nil {
// 				mk := this.merklePool.Get().(*merkle.Merkle)
// 				mk.Reset()
// 				this.merkles[acct] = mk // one merkle for each account
// 			}
// 		}
// 	}
// }

// // Build a Merkle for every updated account
// func (this *AccountMerkle) Build(keys []string, encodedVals [][]byte) []*string {
// 	common.SortBy1st(keys, encodedVals, func(lhv, rhv string) bool { return lhv < rhv })

// 	if len(keys) == 0 {
// 		return nil
// 	}

// 	t0 := time.Now()
// 	offset := stgcommcommon.ETH10_ACCOUNT_PREFIX_LENGTH
// 	ranges, ParseAccountAddrs := this.markAccountRange(keys)
// 	builder := func(start, end, index int, args ...interface{}) {
// 		mempool := this.nodePool.GetPool(index)
// 		for i := start; i < end; i++ {
// 			path := keys[ranges[i]]
// 			if len(path) == 0 {
// 				continue
// 			}

// 			pos := strings.Index(path[offset:], "/")
// 			acct := path[offset : pos+offset]

// 			serializedKVs := make([][]byte, 0, ranges[i+1]-ranges[i])
// 			for j := ranges[i]; j < ranges[i+1]; j++ {
// 				if keys[j][len(keys[j])-1] != '/' { // Skip the path meta
// 					// serializedKVs = append(serializedKVs, encodedVals[j])
// 					serializedKVs = append(
// 						serializedKVs,
// 						this.encoder(keys[j], encodedVals[j]))
// 				}
// 			}

// 			// Create a merkle
// 			if this.merkles[acct] == nil {
// 				// mk := this.merklePool.Get().(*merkle.Merkle).Reset()
// 				this.merkles[acct] = this.merklePool.Get().(*merkle.Merkle).Reset() // one merkle for each account
// 			}
// 			this.merkles[acct].Init(serializedKVs, mempool)
// 		}
// 	}
// 	common.ParallelWorker(len(ranges)-1, concurrency, builder)
// 	fmt.Println("Build the Tree in:", time.Since(t0))
// 	return ParseAccountAddrs
// }

// // Assume the paths are already sorted
// func (this *AccountMerkle) markAccountRange(paths []string) ([]int, []*string) {
// 	positions := make([]int, 0, len(paths))
// 	positions = append(positions, 0)
// 	current := paths[0]
// 	for i := 1; i < len(paths); i++ {
// 		p0 := current[:stgcommcommon.ETH10_ACCOUNT_PREFIX_LENGTH+stgcommcommon.ETH10_ACCOUNT_LENGTH]
// 		p1 := paths[i][:stgcommcommon.ETH10_ACCOUNT_PREFIX_LENGTH+stgcommcommon.ETH10_ACCOUNT_LENGTH]
// 		if p0 != p1 {
// 			current = paths[i]
// 			positions = append(positions, i)
// 		}
// 	}
// 	positions = append(positions, len(paths))

// 	ParseAccountAddrs := make([]*string, len(positions)-1)
// 	worker := func(start, end, index int, args ...interface{}) {
// 		for i := start; i < end; i++ {
// 			ParseAccountAddrs[i] = &paths[positions[i]]
// 		}
// 	}
// 	common.ParallelWorker(len(ParseAccountAddrs), 6, worker)
// 	return positions, ParseAccountAddrs
// }
