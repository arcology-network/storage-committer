package state

import (
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	"github.com/arcology-network/concurrenturl/v2/commutative"
)

type AccessFilter struct{}

func (AccessFilter) NonCumulative(v ccurlcommon.UnivalueInterface, _ ...interface{}) bool { // Cumulative = Commutative + Numeric
	return v.Reads() > 0 || v.Writes() > 0 || v.TypeID() == commutative.PATH // Path is Commutative but not Numeric
}

type TransitionFilter struct{}

func (TransitionFilter) ReadOnly(v ccurlcommon.UnivalueInterface) bool { return v.IsReadOnly() }
func (TransitionFilter) DelNonExist(v ccurlcommon.UnivalueInterface) bool {
	return v.Value() == nil && !v.Preexist()
}

func (AccessFilter) TransitionSelector(v ccurlcommon.UnivalueInterface, _ ...interface{}) []interface{} {
	return common.IfThen(
		!v.Preexist(),
		[]interface{}{false, true, true, true},
		[]interface{}{false, true, false, false},
	)
}

func (AccessFilter) AccessMetaSelector(v ccurlcommon.UnivalueInterface, _ ...interface{}) []interface{} {
	return common.IfThen(
		v.DeltaWrites() > 0 && v.Reads() == 0 && v.Writes() == 0 && v.TypeID() != commutative.PATH,
		[]interface{}{true, true, true, true},
		[]interface{}{false, false, false, false},
	)
}

func ReadOnly(v ccurlcommon.UnivalueInterface) bool { return v.IsReadOnly() }
func DelNonExist(v ccurlcommon.UnivalueInterface) bool {
	return v.Value() == nil && !v.Preexist()
}
