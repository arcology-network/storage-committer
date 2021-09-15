package ccurltype

import (
	"fmt"
	"time"

	"github.com/arcology-network/common-lib/common"
	merkle "github.com/arcology-network/common-lib/merkle"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	orderedmap "github.com/elliotchance/orderedmap"
)

type AccountMerkle struct {
	byAcct   *orderedmap.OrderedMap
	branches uint32
	merkles  *map[string]*merkle.Merkle
	platform *ccurlcommon.Platform
}

func NewAccountMerkle(platform *ccurlcommon.Platform) *AccountMerkle {
	merkles := make(map[string]*merkle.Merkle)
	return &AccountMerkle{
		branches: 16,
		byAcct:   orderedmap.NewOrderedMap(),
		merkles:  &merkles,
		platform: platform,
	}
}
func (this *AccountMerkle) GetMerkles() map[string]*merkle.Merkle {
	return *this.merkles
}

// Insert to the merkle tree
func (this *AccountMerkle) Import(transitions []ccurlcommon.UnivalueInterface) {
	for _, v := range transitions {
		path := v.GetPath()
		pos := ccurlcommon.SubpathOf(this.platform.Eth10Account(), path)
		if pos >= 0 {
			acct := path[:pos]
			byAcct, ok := this.byAcct.Get(acct)
			if !ok {
				this.byAcct.Set(acct, orderedmap.NewOrderedMap()) // Add if not in the account dictionary yet
				byAcct, _ = this.byAcct.Get(acct)
			}
			byAcct.(*orderedmap.OrderedMap).Set(path, true) // Add to the account

			if (*this.merkles)[acct] == nil {
				(*this.merkles)[acct] = merkle.NewMerkle(int(this.branches), merkle.Sha256) // one merkle for each account
			}
		}
	}
}

// Build a Merkle for every updated account
func (this *AccountMerkle) Build(branches int, stateDict *map[string]*orderedmap.Element) []string {
	t0 := time.Now()
	uniqueAccts := this.byAcct.Keys()
	flags := make([]bool, len(uniqueAccts))
	hasher := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			keyDict, _ := this.byAcct.Get(uniqueAccts[i]) // Get all paths under the same account

			// All the changes under the same account
			data := make([][]byte, 0, keyDict.(*orderedmap.OrderedMap).Len())
			for iter := keyDict.(*orderedmap.OrderedMap).Front(); iter != nil; iter = iter.Next() {
				if v := (*stateDict)[iter.Key.(string)]; v != nil { // 120
					univ := v.Value.(ccurlcommon.UnivalueInterface)
					data = append(data, univ.GetCachedEncoded())
				}
			}
			flags[i] = len(data) > 0 // Found at least one transition under the account
			(*this.merkles)[uniqueAccts[i].(string)].Init(data)
		}
	}
	common.ParallelWorker(len(uniqueAccts), 6, hasher)
	fmt.Println("Build the Tree :", fmt.Sprint(100000*9), time.Since(t0))

	t0 = time.Now()
	accounts := make([]string, 0, len(uniqueAccts))
	for i := 0; i < len(uniqueAccts); i++ {
		if flags[i] {
			accounts = append(accounts, uniqueAccts[i].(string))
		}
	}
	fmt.Println("accounts ========= :", time.Since(t0))
	return accounts
}
