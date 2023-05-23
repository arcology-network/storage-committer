package indexer

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
)

type Accumulator struct {
	transitions *map[string]*[]ccurlcommon.UnivalueInterface
}

func (this *Accumulator) CheckLimits(transitions *map[string]*[]ccurlcommon.UnivalueInterface) *Accumulator {
	for _, sequence := range *transitions {
		if (*sequence)[0].Value() == nil {
			continue
		}

		v := (*sequence)[0].Value().(ccurlcommon.TypeInterface)
		if v.IsCommutative() && v.IsNumeric() {
			v.Clone().(ccurlcommon.TypeInterface).ApplyDelta((*sequence)[1:])
		}
	}
	return this
}

// func (this *Accumulator) CheckMin(transitions []ccurlcommon.UnivalueInterface) *Accumulator {
// 	for _, sequence := range *transitions {
// 		if (*sequence)[0].Value() == nil {
// 			continue
// 		}

// 		v := (*sequence)[0].Value().(ccurlcommon.TypeInterface)
// 		if v.IsCommutative() && v.IsNumeric() {
// 			v.Clone().(ccurlcommon.TypeInterface).ApplyDelta((*sequence)[1:])
// 		}
// 	}
// 	return this
// }

// func (this *Accumulator) CheckMax(transitions []ccurlcommon.UnivalueInterface) *Accumulator {
// 	for _, sequence := range transitions {
// 		if (sequence)[0].Value() == nil {
// 			continue
// 		}

// 		v := (*sequence)[0].Value().(ccurlcommon.TypeInterface)
// 		if v.IsCommutative() && v.IsNumeric() {
// 			v.Clone().(ccurlcommon.TypeInterface).ApplyDelta((*sequence)[1:])
// 		}
// 	}
// 	return this
// }
