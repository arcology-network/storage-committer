package importer

import (
	"sort"
	"sync"

	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	committercommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	intf "github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type DeltaSequence struct {
	key         string
	transitions []*univalue.Univalue
	lock        sync.RWMutex
	rawBytes    interface{}
}

func NewDeltaSequence(key string, store interfaces.Datastore) *DeltaSequence {
	seq := &DeltaSequence{
		key:         key,
		transitions: make([]*univalue.Univalue, 0, 16),
		rawBytes:    nil,
	}

	if store != nil {
		seq.rawBytes = common.FilterFirst(store.Retrive(key, nil))
	}
	return seq
}

func (this *DeltaSequence) Init(key string, store interfaces.Datastore) *DeltaSequence {
	this.key = key
	this.transitions = this.transitions[:0]
	this.rawBytes = common.FilterFirst(store.Retrive(key, nil))
	return this
}

func (this *DeltaSequence) Add(v *univalue.Univalue) *DeltaSequence {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.transitions = append(this.transitions, v)
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
	array.RemoveIf(&this.transitions, func(_ int, v *univalue.Univalue) bool {
		return v.GetPath() == nil
	})

	if len(this.transitions) == 0 {
		return nil
	}

	finalized := this.transitions[0]

	if (this.rawBytes != nil) && (finalized.Value() != nil) { // Value update not an assignment or deletion
		if encoded, ok := this.rawBytes.([]byte); ok {
			v := finalized.Value().(intf.Type).StorageDecode(encoded).(intf.Type).Value()
			finalized.Value().(intf.Type).SetValue(v)
		}
	}

	if err := finalized.ApplyDelta(this.transitions[1:]); err != nil {
		panic(err)
	}
	return finalized
}

func (this *DeltaSequence) Reclaim() {
	for i := range this.transitions {
		this.transitions[i] = nil
	}
}
