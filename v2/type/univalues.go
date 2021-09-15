package ccurltype

import (
	"fmt"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Univalues []ccurlcommon.UnivalueInterface

func (this Univalues) IfContains(condition ccurlcommon.UnivalueInterface) bool {
	for _, v := range this {
		if (v).(*Univalue).EqualTransition(condition.(*Univalue)) {
			return true
		}
	}
	return false
}

func (this Univalues) Encode() []byte {
	byteset := make([][]byte, len(this))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			byteset[i] = this[i].Encode()
		}
	}
	common.ParallelWorker(len(this), 6, worker)
	return codec.Byteset(byteset).Encode()
}

func (Univalues) Decode(bytes []byte) interface{} {
	bytesset := codec.Byteset{}.Decode(bytes)
	univalues := make([]ccurlcommon.UnivalueInterface, len(bytesset))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			v := (&Univalue{}).Decode(bytesset[i])
			univalues[i] = v.(ccurlcommon.UnivalueInterface)
		}
	}
	common.ParallelWorker(len(bytesset), 6, worker)
	return Univalues(univalues)
}

func (this Univalues) EncodeV2() [][]byte {
	byteset := make([][]byte, len(this))
	for i := range this {
		byteset[i] = this[i].Encode()
	}
	return byteset
}

func (Univalues) DecodeV2(bytesset [][]byte) Univalues {
	univalues := make([]ccurlcommon.UnivalueInterface, len(bytesset))
	for i := range bytesset {
		v := (&Univalue{}).Decode(bytesset[i])
		univalues[i] = v.(ccurlcommon.UnivalueInterface)
	}
	return Univalues(univalues)
}

func (this Univalues) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}

func (this *Univalues) GobDecode(data []byte) error {
	v := this.Decode(data)
	*this = v.(Univalues)
	return nil
}

func (this Univalues) Print() {
	for _, v := range this {
		v.Print()
	}
	fmt.Println(" --------------------  ")
}
