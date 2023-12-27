package importer

import (
	"errors"
	"sort"

	common "github.com/arcology-network/common-lib/common"
	committercommon "github.com/arcology-network/concurrenturl/common"
	intf "github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/concurrenturl/univalue"
)

type Accumulator struct{}

// check if the value is out of limits. It sorts the transitions by the delta sign and type.
// The neative delta is in the front of the array to make sure it has sufficient balance for the positive deltas.
// The underflow is always checked first before the overflow.
func (this *Accumulator) CheckMinMax(transitions []*univalue.Univalue) []*Conflict {
	if len(transitions) <= 1 ||
		(transitions)[0].Value() == nil ||
		!(transitions)[0].Value().(intf.Type).IsCommutative() ||
		!(transitions)[0].Value().(intf.Type).IsNumeric() {
		return nil
	}

	common.RemoveIf(&transitions, func(v *univalue.Univalue) bool {
		return v.IsReadOnly()
	})

	if len(transitions) <= 1 {
		return nil
	}

	sort.SliceStable(transitions, func(i, j int) bool {
		lhv := transitions[i].Value().(intf.Type)
		rhv := transitions[i].Value().(intf.Type)
		return lhv.DeltaSign() != rhv.DeltaSign() && !lhv.DeltaSign()
	})

	negatives, positives := this.Categorize(transitions)

	// check underflow first
	underflowed := this.isOutOfLimits(*(transitions)[0].GetPath(), negatives)
	if underflowed != nil {
		underflowed.Err = errors.New(committercommon.WARN_OUT_OF_LOWER_LIMIT)
	}

	// check overflow
	overflowed := this.isOutOfLimits(*(transitions)[0].GetPath(), positives)
	if overflowed != nil {
		overflowed.Err = errors.New(committercommon.WARN_OUT_OF_UPPER_LIMIT)
	}

	if overflowed == nil && underflowed == nil {
		return nil
	}

	conflicts := []*Conflict{}
	if underflowed != nil {
		conflicts = append(conflicts, underflowed)
	}

	if overflowed != nil {
		conflicts = append(conflicts, overflowed)
	}
	return conflicts
}

// categorize transitions into two groups, one is negative, the other is positive.
func (*Accumulator) Categorize(transitions []*univalue.Univalue) ([]*univalue.Univalue, []*univalue.Univalue) {
	offset := common.LocateFirstIf(transitions, func(v *univalue.Univalue) bool { return v.Value().(intf.Type).DeltaSign() })

	if offset < 0 {
		offset = len(transitions)
	}
	return transitions[:offset], transitions[offset:]
}

// check if the value is out of limits
func (this *Accumulator) isOutOfLimits(k string, transitions []*univalue.Univalue) *Conflict {
	if len(transitions) <= 1 {
		return nil
	}

	initialv := transitions[0].Value().(intf.Type).Clone().(intf.Type)

	typedVals := common.Append(transitions, func(v *univalue.Univalue) intf.Type {
		return v.Value().(intf.Type)
	})

	_, length, err := initialv.ApplyDelta(typedVals[1:])
	if err == nil {
		return nil
	}

	txIDs := []uint32{}
	common.Foreach(transitions[length+1:], func(v **univalue.Univalue, _ int) { txIDs = append(txIDs, (*v).GetTx()) })

	return &Conflict{
		key:   k,
		txIDs: txIDs,
	}
}
