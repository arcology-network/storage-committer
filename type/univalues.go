package urltype

import (
	"fmt"

	"github.com/HPISTechnologies/common-lib/codec"
	urlcommon "github.com/HPISTechnologies/concurrenturl/common"
)

type Univalues []urlcommon.UnivalueInterface

func (this Univalues) IfContains(condition urlcommon.UnivalueInterface) bool {
	for _, v := range this {
		if (v).(*Univalue).EqualAccess(condition.(*Univalue)) {
			return true
		}
	}
	return false
}

func (this Univalues) Print() {
	for _, v := range this {
		v.Print()
	}
	fmt.Println(" --------------------  ")
}

func (this Univalues) Encode() []byte {
	byteset := make([][]byte, len(this))
	for i, v := range this {
		byteset[i] = v.Encode()
	}
	return codec.Byteset(byteset).Encode()
}

func (Univalues) Decode(bytes []byte, urlcodec urlcommon.Decoder) interface{} {
	bytesset := codec.Byteset{}.Decode(bytes)
	univalues := make([]urlcommon.UnivalueInterface, len(bytesset))
	for i, v := range bytesset {
		v := (&Univalue{}).Decode(v, urlcodec)
		univalues[i] = v.(urlcommon.UnivalueInterface)
	}
	return Univalues(univalues)
}
