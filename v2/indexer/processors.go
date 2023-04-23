package indexer

import (
	"bytes"
	"sort"

	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

func Sorter(univals []ccurlcommon.UnivalueInterface) interface{} {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}

// Skip meta for system paths
func Bypasser(univals []ccurlcommon.UnivalueInterface, args interface{}) []ccurlcommon.UnivalueInterface {
	platform := args.(*ccurlcommon.Platform)
	for i := 0; i < len(univals); i++ {
		if platform.IsSysPath(*univals[i].GetPath()) && // Skip meta changes for the system path
			univals[i].TypeID() == ccurlcommon.CommutativeMeta &&
			univals[i].Writes() == 0 { // Meta CREATION and DELETION should pass!!!
			univals[i] = nil
		}
	}
	common.RemoveIf(&univals, func(v ccurlcommon.UnivalueInterface) bool { return v == nil })
	return univals
}

// For the arbitrator and accumulator
func FilterAccesses(buffer []ccurlcommon.UnivalueInterface, platform interface{}) []ccurlcommon.UnivalueInterface {
	univals := common.DeepCopy(buffer)
	for i := 0; i < len(univals); i++ {
		if univals[i].DeltaWrites() > 0 && univals[i].TypeID() != ccurlcommon.CommutativeMeta {
			// univals[i].SetValue(univals[i].Value().(ccurlcommon.TypeInterface).Delta()) // Numeric Delta writes only, no meta
		} else {
			// univals[i].SetValue(nil)
		}
	}
	return univals
}

// For state storage
func FilterTransitions(buffer []ccurlcommon.UnivalueInterface, platform interface{}) []ccurlcommon.UnivalueInterface {
	univals := Bypasser(common.DeepCopy(buffer), platform)

	for i := 0; i < len(univals); i++ {
		if univals[i].Writes() > 0 || univals[i].DeltaWrites() > 0 {
			if univals[i].Value() != nil { // A new one or update or an existing entry
				univals[i].SetValue(univals[i].Value().(ccurlcommon.TypeInterface).Delta())
			} else {
				if !univals[i].Preexist() { // Delete entries, ignore non existing ones
					univals[i] = nil
				}
			}
		} else {
			univals[i] = nil
		}
	}
	common.RemoveIf(&univals, func(v ccurlcommon.UnivalueInterface) bool { return v == nil })
	return univals
}
