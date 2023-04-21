package univalue

func CreateUnivalueForTest(transitType uint8, vType uint8, tx uint32, path string, reads, writes uint32, value interface{}, preexists, composite bool) *Univalue {
	return &Univalue{
		vType:     vType,
		tx:        tx,
		path:      &path,
		reads:     reads,
		writes:    writes,
		value:     value,
		preexists: preexists,
		// composite: composite,
		reserved: nil,
	}
}

// Only work when
// func SetInvariate(trans []ccurlcommon.UnivalueInterface, name string) {
// 	for i := 0; i < len(trans); i++ {
// 		if strings.Contains(*(trans[i].GetPath()), name) {
// 			trans[i].SetTransitionType(ccurlcommon.INVARIATE_TRANSITIONS)
// 		}
// 	}
// }
