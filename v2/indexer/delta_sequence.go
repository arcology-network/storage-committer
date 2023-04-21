package indexer

import (
	"sort"
	"sync"

	"github.com/arcology-network/common-lib/mempool"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	univalue "github.com/arcology-network/concurrenturl/v2/univalue"
)

type DeltaSequence struct {
	key    string
	values []ccurlcommon.UnivalueInterface
	base   ccurlcommon.UnivalueInterface
	lock   sync.RWMutex
}

func NewDeltaSequence() *DeltaSequence {
	return &DeltaSequence{
		values: make([]ccurlcommon.UnivalueInterface, 0, 16),
	}
}

func (this *DeltaSequence) Reset(key string, indexer *Indexer, mempool *mempool.Mempool) {
	this.key = key
	this.values = this.values[:0]
	this.base = nil
}

func (this *DeltaSequence) Init(key string, indexer *Indexer, mempool *mempool.Mempool) {
	if initialState := indexer.RetriveShallow(key); initialState != nil {
		nVal := mempool.Get().(*univalue.Univalue)
		nVal.Init(ccurlcommon.SYSTEM, key, 0, 0, initialState.(ccurlcommon.TypeInterface).Deepcopy(), indexer)
		this.values = append(this.values, nVal) //Transitions are ordered by Tx, -1 will guarantee the initial state is always the first one
	}
}

func (this *DeltaSequence) Value() interface{} {
	return this.base
}

func (this *DeltaSequence) Insert(v ccurlcommon.UnivalueInterface) {
	this.lock.Lock()
	this.values = append(this.values, v.(*univalue.Univalue))
	this.lock.Unlock()
}

func (this *DeltaSequence) Sort() {
	if len(this.values) <= 1 {
		return
	}

	if len(this.values) == 2 && this.values[0].GetTx() == ccurlcommon.SYSTEM {
		return
	}

	sort.SliceStable(this.values, func(i, j int) bool {
		if this.values[i].GetTx() == ccurlcommon.SYSTEM {
			return true
		}

		if this.values[j].GetTx() == ccurlcommon.SYSTEM {
			return false
		}

		return this.values[i].GetTx() < this.values[j].GetTx()
	})
}

func (this *DeltaSequence) Finalize() {
	if len(this.values) == 0 {
		return
	}

	i := 0
	for ; i < len(this.values); i++ {
		if this.values[i].GetPath() != nil {
			this.base = this.values[i]
			break
		}
	}

	if this.base == nil {
		return
	}

	if err := this.base.ApplyDelta(ccurlcommon.SYSTEM, this.values[i+1:]); err != nil {
		panic(err)
	}

	if this.base.Value() != nil {
		this.base.Value().(ccurlcommon.TypeInterface).Purge() // Remove non-essential attributes
	}
}

func (this *DeltaSequence) Reclaim() {
	for i := range this.values {
		this.values[i] = nil
	}
}
