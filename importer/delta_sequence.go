package importer

import (
	"sort"
	"sync"

	common "github.com/arcology-network/common-lib/common"
	committercommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type DeltaSequence struct {
	key         string
	transitions []interfaces.Univalue
	lock        sync.RWMutex
	rawBytes    interface{}
}

func NewDeltaSequence(key string, importer *Importer) *DeltaSequence {
	return &DeltaSequence{
		key:         key,
		transitions: make([]interfaces.Univalue, 0, 16),
		rawBytes:    common.FilterFirst(importer.store.Retrive(key, nil)),
		// initial: (&univalue.Univalue{}).Init(committercommon.SYSTEM, key, 0, 0, 0, encoded, importer.Store()),
	}
}

func (this *DeltaSequence) Reset(key string) *DeltaSequence {
	this.key = key
	this.transitions = this.transitions[:0]
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
		if this.transitions[i].GetTx() == committercommon.SYSTEM {
			return true
		}

		if this.transitions[j].GetTx() == committercommon.SYSTEM {
			return false
		}

		return this.transitions[i].GetTx() < this.transitions[j].GetTx()
	})
}

func (this *DeltaSequence) Finalize() *univalue.Univalue {
	common.RemoveIf(&this.transitions, func(v interfaces.Univalue) bool {
		return v.GetPath() == nil
	})

	if len(this.transitions) == 0 {
		return nil
	}
	finalized := this.transitions[0].(*univalue.Univalue)

	if (this.rawBytes != nil) && (finalized.Value() != nil) { // Value update not an assignment or deletion
		if encoded, ok := this.rawBytes.([]byte); ok {
			v := finalized.Value().(interfaces.Type).StorageDecode(encoded).(interfaces.Type).Value()
			finalized.Value().(interfaces.Type).SetValue(v)
		}
	}

	if err := finalized.ApplyDelta(this.transitions[1:]); err != nil {
		panic(err)
	}
	return finalized
}

// func (this *DeltaSequence) Finalize() *univalue.Univalue {
// 	if len(this.transitions) == 0 {
// 		return nil
// 	}
// 	finalized := this.transitions[0].(*univalue.Univalue)

// 	if (this.rawBytes != nil) && (finalized.Value() != nil) { // Value update not an assignment or deletion
// 		var v interface{}
// 		if encoded, ok := this.rawBytes.([]byte); ok {
// 			v = finalized.Value().(interfaces.Type).StorageDecode(encoded).(interfaces.Type).Value()
// 			finalized.Value().(interfaces.Type).SetValue(v)
// 		}
// 	}

// 	if err := finalized.ApplyDelta(this.transitions[1:]); err != nil {
// 		panic(err)
// 	}
// 	return finalized
// }

func (this *DeltaSequence) Reclaim() {
	for i := range this.transitions {
		this.transitions[i] = nil
	}
}
