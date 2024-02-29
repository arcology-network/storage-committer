package importer

import (
	"sort"
	"sync"

	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

type DeltaSequence struct {
	Key         string
	Transitions []*univalue.Univalue
	lock        sync.RWMutex
	rawBytes    interface{}
	Finalized   *univalue.Univalue
}

func NewDeltaSequence(key string, store interfaces.Datastore) *DeltaSequence {
	seq := &DeltaSequence{
		Key:         platform.GetAccountAddr(key),
		Transitions: make([]*univalue.Univalue, 0, 16),
		rawBytes:    nil,
	}

	if store != nil {
		seq.rawBytes = common.FilterFirst(store.Retrive(key, nil))
	}
	return seq
}

func (this *DeltaSequence) Init(key string, store interfaces.Datastore) *DeltaSequence {
	this.Key = platform.GetAccountAddr(key)
	this.Transitions = this.Transitions[:0]
	this.rawBytes = nil

	if len(key) > 0 {
		this.rawBytes = common.FilterFirst(store.Retrive(key, nil))
	}
	return this
}

// func (this *DeltaSequence) SetFinalized(v *univalue.Univalue) { this.finalized = v } // for debugging only

// func (this *DeltaSequence) Finalized() *univalue.Univalue { return this.finalized }

func (this *DeltaSequence) UnsafeAdd(v *univalue.Univalue) *DeltaSequence {
	this.Transitions = append(this.Transitions, v)
	return this
}

func (this *DeltaSequence) Add(v *univalue.Univalue) *DeltaSequence {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.Transitions = append(this.Transitions, v)
	return this
}

func (this *DeltaSequence) Sort() {
	if len(this.Transitions) <= 1 {
		return
	}

	sort.SliceStable(this.Transitions, func(i, j int) bool {
		if this.Transitions[i].GetTx() == stgcommcommon.SYSTEM {
			return true
		}

		if this.Transitions[j].GetTx() == stgcommcommon.SYSTEM {
			return false
		}

		return this.Transitions[i].GetTx() < this.Transitions[j].GetTx()
	})
}

func (this *DeltaSequence) Finalize() *univalue.Univalue {
	slice.RemoveIf(&this.Transitions, func(_ int, v *univalue.Univalue) bool {
		return v.GetPath() == nil
	})

	if len(this.Transitions) == 0 {
		return nil
	}

	this.Finalized = this.Transitions[0]
	if (this.rawBytes != nil) && (this.Finalized.Value() != nil) { // Value update not an assignment or deletion
		if encoded, ok := this.rawBytes.([]byte); ok {
			v := this.Finalized.Value().(intf.Type).StorageDecode(*this.Finalized.GetPath(), encoded).(intf.Type).Value()
			this.Finalized.Value().(intf.Type).SetValue(v)
		}
	}

	if err := this.Finalized.ApplyDelta(this.Transitions[1:]); err != nil {
		panic(err)
	}
	return this.Finalized
}
