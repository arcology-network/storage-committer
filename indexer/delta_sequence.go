package indexer

import (
	"sort"
	"sync"

	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type DeltaSequence struct {
	key         string
	transitions []interfaces.Univalue
	// initial     interfaces.Univalue
	lock     sync.RWMutex
	rawBytes interface{}
}

func NewDeltaSequence(key string, indexer *Importer) *DeltaSequence {
	return &DeltaSequence{
		key:         key,
		transitions: make([]interfaces.Univalue, 0, 16),
		rawBytes:    common.FilterFirst(indexer.store.Retrive(key, nil)),
		// initial: (&univalue.Univalue{}).Init(ccurlcommon.SYSTEM, key, 0, 0, 0, encoded, indexer.Store()),
	}
}

func (this *DeltaSequence) Reset(key string) *DeltaSequence {
	this.key = key
	this.transitions = this.transitions[:0]
	// this.initial = nil
	return this
}

func (this *DeltaSequence) Add(v interfaces.Univalue) *DeltaSequence {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.transitions = append(this.transitions, v.(*univalue.Univalue))
	return this
}

func (this *DeltaSequence) Sort() {
	if len(this.transitions) <= 1 {
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

func (this *DeltaSequence) Finalize() *univalue.Univalue {
	if len(this.transitions) == 0 {
		return nil
	}

	if (this.rawBytes) == nil { // New value
		return this.transitions[0].(*univalue.Univalue) // Cannot be more than one element in the transtion array
	}

	if this.rawBytes != nil && this.transitions[0].(*univalue.Univalue).Value() == nil { // Deletion
		return this.transitions[0].(*univalue.Univalue) // Cannot be more than one element in the transtion array
	}

	T := this.transitions[0].Value().(interfaces.Type) // Type indicator
	// if this.rawBytes != nil && univ.Value() != nil {
	initial := (&univalue.Univalue{}).SetValue(T.StorageDecode(this.rawBytes.([]byte)).([]byte)).(*univalue.Univalue)
	this.transitions = this.transitions[1:]
	// initial // Update
	// initial.Unimeta.Merge(univ.GetUnimeta().(*univalue.Unimeta))
	// }

	// if this.rawBytes != nil && univ == nil {
	// 	initial.Unimeta.Merge(univ.GetUnimeta().(*univalue.Unimeta))
	// }

	if err := initial.ApplyDelta(this.transitions); err != nil {
		panic(err)
	}
	return initial
}

func (this *DeltaSequence) Reclaim() {
	for i := range this.transitions {
		this.transitions[i] = nil
	}
}
