package importer

import (
	"sort"

	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	intf "github.com/arcology-network/storage-committer/interfaces"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

type DeltaSequenceV2 []*univalue.Univalue

func (this DeltaSequenceV2) sort() DeltaSequenceV2 {
	if len(this) <= 1 {
		return this
	}

	sort.SliceStable(this, func(i, j int) bool {
		if this[i].GetTx() == stgcommcommon.SYSTEM {
			return true
		}

		if this[j].GetTx() == stgcommcommon.SYSTEM {
			return false
		}
		return this[i].GetTx() < this[j].GetTx()
	})
	return this
}

func (this DeltaSequenceV2) Finalize(store intf.ReadOnlyDataStore) *univalue.Univalue {
	trans := []*univalue.Univalue(this)
	slice.RemoveIf(&trans, func(_ int, v *univalue.Univalue) bool {
		return v.GetPath() == nil
	})

	if len(this) == 0 {
		return nil
	}

	this.sort()
	if err := this[0].ApplyDelta(this[1:]); err != nil {
		panic(err)
	}

	// Remove the transition to indicate that the delta sequence has been finalized
	this = this[:1]
	return this[0]
}

func (this DeltaSequenceV2) Finalized() *univalue.Univalue { return this[0] }

type DeltaSequencesV2 []DeltaSequenceV2

func (this DeltaSequencesV2) Finalized() []intf.Type {
	return slice.Transform(this, func(_ int, v DeltaSequenceV2) intf.Type {
		return v[0].Value().(intf.Type)
	})
}

func (this DeltaSequencesV2) Keys() []*string {
	return slice.Transform(this, func(_ int, v DeltaSequenceV2) *string {
		return v[0].GetPath()
	})
}
