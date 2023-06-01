package indexer

import (
	"fmt"
	"strings"
	"time"

	common "github.com/arcology-network/common-lib/common"
	mempool "github.com/arcology-network/common-lib/mempool"
	merkle "github.com/arcology-network/common-lib/merkle"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

const (
	concurrency = 8
)

type AccountMerkle struct {
	branches   uint32
	merkles    map[string]*merkle.Merkle
	platform   interfaces.Platform
	nodePool   *mempool.Mempool
	merklePool *mempool.Mempool
}

func NewAccountMerkle(platform interfaces.Platform) *AccountMerkle {
	am := &AccountMerkle{
		branches: 16,
		merkles:  make(map[string]*merkle.Merkle),
		platform: platform,
		nodePool: mempool.NewMempool("node", func() interface{} {
			return merkle.NewNode()
		}),
		merklePool: mempool.NewMempool("merkle", func() interface{} {
			return merkle.NewMerkle(16, merkle.Sha256)
		}),
	}
	return am
}

func (this *AccountMerkle) Clear() {
	this.merkles = make(map[string]*merkle.Merkle)
	this.nodePool.ReclaimRecursive()
	this.merklePool.ReclaimRecursive()
}

func (this *AccountMerkle) GetMerkles() *map[string]*merkle.Merkle {
	return &this.merkles
}

// Insert to the merkle tree
func (this *AccountMerkle) Import(transitions []interfaces.Univalue) {
	offset := len(this.platform.Eth10Account())
	for _, v := range transitions {
		path := *v.GetPath()
		pos := strings.Index(path[offset:], "/")
		if pos >= 0 {
			acct := path[offset : pos+offset]
			if this.merkles[acct] == nil {
				mk := this.merklePool.Get().(*merkle.Merkle)
				mk.Reset()
				this.merkles[acct] = mk // one merkle for each account
			}
		}
	}
}

// Build a Merkle for every updated account
func (this *AccountMerkle) Build(keys []string, values [][]byte) []*string {
	common.SortBy1st(keys, values, func(lhv, rhv string) bool { return lhv < rhv })

	if len(keys) == 0 {
		return nil
	}

	t0 := time.Now()
	offset := len(this.platform.Eth10Account())
	ranges, accountKeys := this.markAccountRange(keys)
	hasher := func(start, end, index int, args ...interface{}) {
		mempool := this.nodePool.GetTlsMempool(index)
		for i := start; i < end; i++ {
			path := keys[ranges[i]]
			if len(path) == 0 {
				continue
			}

			pos := strings.Index(path[offset:], "/")
			acct := path[offset : pos+offset]

			dataSet := make([][]byte, 0, ranges[i+1]-ranges[i])
			for j := ranges[i]; j < ranges[i+1]; j++ {
				if keys[j][len(keys[j])-1] == '/' {
					continue // Skip path meta
				}
				dataSet = append(dataSet, values[j])
			}

			// Create a merkle
			if this.merkles[acct] == nil {
				// mk := this.merklePool.Get().(*merkle.Merkle).Reset()
				this.merkles[acct] = this.merklePool.Get().(*merkle.Merkle).Reset() // one merkle for each account
			}
			this.merkles[acct].Init(dataSet, mempool)
		}
	}
	common.ParallelWorker(len(ranges)-1, concurrency, hasher)
	fmt.Println("Build the Tree in:", time.Since(t0))
	return accountKeys
}

// Assume the paths are already sorted
func (this *AccountMerkle) markAccountRange(paths []string) ([]int, []*string) {
	positions := make([]int, 0, len(paths))
	positions = append(positions, 0)
	current := paths[0]
	for i := 1; i < len(paths); i++ {
		p0 := current[:len(this.platform.Eth10Account())+ccurlcommon.ETH10_ACCOUNT_LENGTH]
		p1 := paths[i][:len(this.platform.Eth10Account())+ccurlcommon.ETH10_ACCOUNT_LENGTH]
		if p0 != p1 {
			current = paths[i]
			positions = append(positions, i)
		}
	}
	positions = append(positions, len(paths))

	accountKeys := make([]*string, len(positions)-1)
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			accountKeys[i] = &paths[positions[i]]
		}
	}
	common.ParallelWorker(len(accountKeys), 6, worker)
	return positions, accountKeys
}