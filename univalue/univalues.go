package univalue

import (
	"github.com/arcology-network/concurrenturl/interfaces"
)

func UnivaluesDecode(bytesset [][]byte, get func() interface{}, put func(interface{})) []interfaces.Univalue {
	univalues := make([]interfaces.Univalue, len(bytesset))
	for i := range bytesset {
		v := get().(*Univalue)
		v.reclaimFunc = put
		v.Decode(bytesset[i])
		univalues[i] = v
	}
	return univalues
}
func UnivaluesEncode(this []interfaces.Univalue) [][]byte {
	byteset := make([][]byte, len(this))
	for i := range this {
		byteset[i] = this[i].Encode()
	}
	return byteset
}
