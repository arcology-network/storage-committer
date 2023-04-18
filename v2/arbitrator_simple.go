package concurrenturl

import (
	"bytes"
	"sort"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"

	orderedmap "github.com/elliotchance/orderedmap"
	murmur "github.com/spaolacci/murmur3"
)

type ArbitratorSlow struct {
	transitions map[string]*[]ccurlcommon.UnivalueInterface
}

func NewArbitratorSlow() *ArbitratorSlow {
	return &ArbitratorSlow{
		transitions: make(map[string]*[]ccurlcommon.UnivalueInterface),
	}
}

func (this *ArbitratorSlow) Detect(newTrans []ccurlcommon.UnivalueInterface) (map[string][]ccurlcommon.UnivalueInterface, []uint32) {
	for _, trans := range newTrans {
		if arr, ok := this.transitions[*trans.GetPath()]; !ok {
			this.transitions[*trans.GetPath()] = &[]ccurlcommon.UnivalueInterface{trans}
		} else {
			(*arr) = append((*arr), trans)
		}
	}

	// ccmap := concurrentmap.NewConcurrentMap()
	// ccmap.BatchSet(ccurltype.Univalues(newTrans).Keys(), common.From(newTrans))

	for _, value := range this.transitions {
		v := *value
		if len(v) == 1 {
			continue
		}

		sort.SliceStable(v, func(i, j int) bool {
			if len(*v[i].GetPath()) != len(*v[j].GetPath()) {
				return len(*v[i].GetPath()) < len(*v[j].GetPath())
			}

			if !bytes.Equal([]byte(*v[i].GetPath()), []byte(*v[j].GetPath())) {
				return bytes.Compare([]byte(*v[i].GetPath()), []byte(*v[j].GetPath())) < 0
			}

			if v[i].GetTx() != v[j].GetTx() {
				return v[i].GetTx() < v[j].GetTx()
			}

			if v[i].Writes() != v[j].Writes() {
				return v[i].Writes() < v[j].Writes()
			}

			if v[i].Reads() != v[j].Reads() {
				return v[i].Reads() < v[j].Reads()
			}
			return false
		})
	}

	txToRemove := orderedmap.NewOrderedMap()
	conflictDict := make(map[string][]ccurlcommon.UnivalueInterface)
	for _, value := range this.transitions {
		v := *value
		for i := 1; i < len(v); i++ {
			if v[0].GetTx() == v[i].GetTx() {
				continue
			}

			if v[0].DeltaWritable() && v[i].DeltaWritable() {
				continue
			}

			if v[0].Writes() > 0 || v[i].Writes() > 0 {
				conflictDict[*v[0].GetPath()] = append(conflictDict[*v[0].GetPath()], v[i])
				txToRemove.Set(v[i].GetTx(), true)
			}
		}
	}
	keys := txToRemove.Keys()
	txs := make([]uint32, len(keys))
	for i, v := range keys {
		txs[i] = v.(uint32)
	}

	this.transitions = make(map[string]*[]ccurlcommon.UnivalueInterface)
	return conflictDict, txs
}

func HashPaths(records []ccurlcommon.UnivalueInterface) {
	numThreads := 1
	if len(records) > 128 {
		numThreads = 4
	}

	hasher := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			h0, h1 := murmur.Sum128(codec.String(*records[i].GetPath()).Encode())
			records[i].SetPath(codec.Bytes(codec.Uint64(h0).Encode()).ToString() + codec.Bytes(codec.Uint64(h1).Encode()).ToString())
		}
	}
	common.ParallelWorker(len(records), numThreads, hasher)
}
