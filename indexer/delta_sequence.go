package indexer

import (
	"sort"
	"sync"

	"github.com/arcology-network/common-lib/mempool"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type DeltaSequence struct {
	key         string
	transitions []interfaces.Univalue
	base        interfaces.Univalue
	lock        sync.RWMutex
}

func NewDeltaSequence() *DeltaSequence {
	return &DeltaSequence{
		transitions: make([]interfaces.Univalue, 0, 16),
	}
}

func (this *DeltaSequence) Reset(key string, indexer *Importer, mempool *mempool.Mempool) {
	this.key = key
	this.transitions = this.transitions[:0]
	this.base = nil
}

func (this *DeltaSequence) Init(key string, indexer *Importer, mempool *mempool.Mempool) {
	if initialState := indexer.RetriveShallow(key); initialState != nil {
		nVal := mempool.Get().(*univalue.Univalue)
		nVal.Init(ccurlcommon.SYSTEM, key, 0, 0, initialState.(interfaces.Type).Clone(), indexer)
		this.transitions = append(this.transitions, nVal) //Transitions are ordered by Tx, -1 will guarantee the initial state is always the first one
	}
}

func (this *DeltaSequence) Value() interface{} {
	return this.base
}

func (this *DeltaSequence) Insert(v interfaces.Univalue) {
	this.lock.Lock()
	this.transitions = append(this.transitions, v.(*univalue.Univalue))
	this.lock.Unlock()
}

func (this *DeltaSequence) Sort() {
	if len(this.transitions) <= 1 {
		return
	}

	if len(this.transitions) == 2 && this.transitions[0].GetTx() == ccurlcommon.SYSTEM {
		return
	}

	sort.SliceStable(this.transitions, func(i, j int) bool {
		if this.transitions[i].GetTx() == ccurlcommon.SYSTEM {
			return true
		}

		if this.transitions[j].GetTx() == ccurlcommon.SYSTEM {
			return false
		}

		return this.transitions[i].GetTx() < this.transitions[j].GetTx()
	})
}

func (this *DeltaSequence) Finalize() {
	if len(this.transitions) == 0 {
		return
	}

	i := 0
	for ; i < len(this.transitions); i++ { // Find th first non nil value, where transitions will be applied on
		if this.transitions[i].GetPath() != nil {
			this.base = this.transitions[i]
			break
		}
	}

	if this.base == nil {
		return
	}

	if err := this.base.ApplyDelta(this.transitions[i+1:]); err != nil {
		panic(err)
	}
}

func (this *DeltaSequence) Reclaim() {
	for i := range this.transitions {
		this.transitions[i] = nil
	}
}
