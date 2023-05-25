package indexer

import (
	"fmt"
	"time"

	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	univalue "github.com/arcology-network/concurrenturl/univalue"

	murmur "github.com/spaolacci/murmur3"
)

type Arbitrator struct {
	dict map[string]*[]ccurlcommon.UnivalueInterface
}

func NewArbitrator() *Arbitrator {
	return &Arbitrator{
		dict: make(map[string]*[]ccurlcommon.UnivalueInterface),
	}
}

func (this *Arbitrator) Detect(newTrans []ccurlcommon.UnivalueInterface) []*Conflict {
	t0 := time.Now()
	univalue.Univalues(newTrans).Sort()
	fmt.Println("Sort: ", time.Since(t0))

	ranges := common.FindRange(newTrans, func(lhv, rhv ccurlcommon.UnivalueInterface) bool { return *lhv.GetPath() == *rhv.GetPath() })

	conflicts := []*Conflict{}
	for i := 0; i < len(ranges)-1; i++ {
		if ranges[i]+1 == ranges[i+1] {
			continue // Only one entry
		}

		var offset int
		if newTrans[ranges[i]].Writes() == 0 {
			if newTrans[ranges[i]].IsConcurrentWritable() {
				offset = common.LocateFirstIf(newTrans[ranges[i]+1:ranges[i+1]], func(v ccurlcommon.UnivalueInterface) bool { return !v.IsConcurrentWritable() })
			} else {
				offset = common.LocateFirstIf(newTrans[ranges[i]+1:ranges[i+1]], func(v ccurlcommon.UnivalueInterface) bool { return v.Writes() > 0 || v.DeltaWrites() > 0 }) + 1
			}
		} else {
			offset = ranges[i] + 1
		}

		if offset >= 0 {
			ids := []uint32{}
			common.Foreach(newTrans[offset:ranges[i+1]], func(v *ccurlcommon.UnivalueInterface) { ids = append(ids, (*v).GetTx()) })
			conflicts = append(conflicts,
				&Conflict{
					key:   *newTrans[ranges[i]].GetPath(),
					txIDs: ids,
					err:   nil,
				},
			)
		}
	}
	return conflicts
}

func HashPaths(records []ccurlcommon.UnivalueInterface) {
	numThreads := 1
	if len(records) > 128 {
		numThreads = 4
	}

	hasher := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			h0, h1 := murmur.Sum128(codec.String(*records[i].GetPath()).Encode())
			path := codec.Bytes(codec.Uint64(h0).Encode()).ToString() + codec.Bytes(codec.Uint64(h1).Encode()).ToString()
			records[i].SetPath(&path)
		}
	}
	common.ParallelWorker(len(records), numThreads, hasher)
}
