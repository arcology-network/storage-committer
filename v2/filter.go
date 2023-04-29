package concurrenturl

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type AccessFilter struct{}

func (AccessFilter) NonCumulative(v ccurlcommon.UnivalueInterface, _ ...interface{}) bool { // Cumulative = Commutative + Numeric
	return v.Reads() > 0 || v.Writes() > 0 || v.TypeID() == ccurlcommon.Commutative{}.Path() // Path is Commutative but not Numeric
}

type TransitionFilter struct{}

func (TransitionFilter) ReadOnly(v ccurlcommon.UnivalueInterface) bool { return v.IsReadOnly() }
func (TransitionFilter) DelNonExist(v ccurlcommon.UnivalueInterface) bool {
	return v.Value() == nil && !v.Preexist()
}

type ValueFilter struct{}

// func (TransitionFilter) ReadOnly(v ccurlcommon.UnivalueInterface) bool { return v.IsReadOnly() }
// func (TransitionFilter) DelNonExist(v ccurlcommon.UnivalueInterface) bool {
// 	return v.Value() == nil && !v.Preexist()
// }

// // func To() {
// 	common.To()
// }
