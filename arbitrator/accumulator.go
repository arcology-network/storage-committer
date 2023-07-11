package indexer

import (
	"errors"
	"sort"

	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type Accumulator struct{}

func (this *Accumulator) CheckMinMax(transitions []interfaces.Univalue) []*Conflict {
	if len(transitions) <= 1 ||
		(transitions)[0].Value() == nil ||
		!(transitions)[0].Value().(interfaces.Type).IsCommutative() ||
		!(transitions)[0].Value().(interfaces.Type).IsNumeric() {
		return nil
	}

	common.RemoveIf(&transitions, func(v interfaces.Univalue) bool {
		return v.IsReadOnly()
	})

	if len(transitions) <= 1 {
		return nil
	}

	sort.SliceStable(transitions, func(i, j int) bool {
		lhv := transitions[i].Value().(interfaces.Type)
		rhv := transitions[i].Value().(interfaces.Type)
		return lhv.DeltaSign() != rhv.DeltaSign() && !lhv.DeltaSign()
	})

	negatives, positives := this.Categorize(transitions)

	underflown := this.isOutOfLimits(*(transitions)[0].GetPath(), negatives)
	if underflown != nil {
		underflown.Err = errors.New(ccurlcommon.ERR_OUT_OF_LOWER_LIMIT)
	}

	overflown := this.isOutOfLimits(*(transitions)[0].GetPath(), positives)
	if overflown != nil {
		overflown.Err = errors.New(ccurlcommon.ERR_OUT_OF_UPPER_LIMIT)
	}

	if overflown == nil && underflown == nil {
		return nil
	}

	conflicts := []*Conflict{}
	if underflown != nil {
		conflicts = append(conflicts, underflown)
	}

	if overflown != nil {
		conflicts = append(conflicts, overflown)
	}
	return conflicts
}

func (*Accumulator) Categorize(transitions []interfaces.Univalue) ([]interfaces.Univalue, []interfaces.Univalue) {
	offset := common.LocateFirstIf(transitions, func(v interfaces.Univalue) bool { return v.Value().(interfaces.Type).DeltaSign() })

	if offset < 0 {
		offset = len(transitions)
	}
	return transitions[:offset], transitions[offset:]
}

func (this *Accumulator) isOutOfLimits(k string, transitions []interfaces.Univalue) *Conflict {
	if len(transitions) <= 1 {
		return nil
	}

	initialv := transitions[0].Value().(interfaces.Type).Clone().(interfaces.Type)
	_, length, err := initialv.ApplyDelta(transitions[1:])
	if err == nil {
		return nil
	}

	txIDs := []uint32{}
	common.Foreach(transitions[length+1:], func(v *interfaces.Univalue) { txIDs = append(txIDs, (*v).GetTx()) })

	return &Conflict{
		key:   k,
		txIDs: txIDs,
	}
}
