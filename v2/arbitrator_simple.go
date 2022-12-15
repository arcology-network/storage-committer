package concurrenturl

import (
	"bytes"
	"sort"

<<<<<<< HEAD
	"github.com/arcology/common-lib/codec"
	"github.com/arcology/common-lib/common"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
=======
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
>>>>>>> a0a836c163772efdc5cf3282d8772664ccf8e355
	orderedmap "github.com/elliotchance/orderedmap"
	murmur "github.com/spaolacci/murmur3"
)

type ArbitratorSlow struct {
	transitions map[string][]ccurlcommon.UnivalueInterface
}

func NewArbitratorSlow() *ArbitratorSlow {
	return &ArbitratorSlow{
		transitions: make(map[string][]ccurlcommon.UnivalueInterface),
	}
}

func (this *ArbitratorSlow) Detect(newTrans []ccurlcommon.UnivalueInterface, whitelist []uint32) (map[string][]ccurlcommon.UnivalueInterface, []uint32) {
	whitelistDict := make(map[uint32]bool)
	for _, v := range whitelist {
		whitelistDict[v] = true
	}

	for _, trans := range newTrans {
		if _, ok := whitelistDict[trans.GetTx()]; !ok {
			this.transitions[*trans.GetPath()] = []ccurlcommon.UnivalueInterface{}
		}
		this.transitions[*trans.GetPath()] = append(this.transitions[*trans.GetPath()], trans)
	}

	for _, v := range this.transitions {
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
	for _, v := range this.transitions {
		for i := 1; i < len(v); i++ {
			if v[0].GetTx() == v[i].GetTx() {
				continue
			}

			if v[0].Composite() && v[i].Composite() {
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

	this.transitions = make(map[string][]ccurlcommon.UnivalueInterface)
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
