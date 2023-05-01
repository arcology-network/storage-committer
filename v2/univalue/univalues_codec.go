package univalue

import (
	"fmt"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

func (this Univalues) Size(selectors ...interface{}) int {
	size := (len(this) + 1) * codec.UINT32_LEN
	for _, v := range this {
		size += int(v.Size(selectors...))
	}
	return size
}

func (this Univalues) Sizes(selectors ...interface{}) []int {
	sizes := make([]int, len(this))
	for i, v := range this {
		sizes[i] = common.IfThenDo1st(v != nil, func() int { return int(v.Size(selectors...)) }, 0)
	}
	return sizes
}

func (this Univalues) Encode(selector ...interface{}) []byte {
	lengths := make([]uint32, len(this))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			if this[i] != nil {
				lengths[i] = this[i].(ccurlcommon.UnivalueInterface).Size(selector...)
			}
		}
	}
	common.ParallelWorker(len(this), 6, worker)

	offsets := make([]uint32, len(this)+1)
	for i := 0; i < len(lengths); i++ {
		offsets[i+1] = offsets[i] + lengths[i]
	}

	headerLen := uint32((len(this) + 1) * codec.UINT32_LEN)
	buffer := make([]byte, headerLen+offsets[len(offsets)-1])

	codec.Uint32(len(this)).EncodeToBuffer(buffer)
	worker = func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			codec.Uint32(offsets[i]).EncodeToBuffer(buffer[(i+1)*codec.UINT32_LEN:])
			this[i].(ccurlcommon.UnivalueInterface).EncodeToBuffer(buffer[headerLen+offsets[i]:])
		}
	}
	common.ParallelWorker(len(this), 6, worker)
	return buffer
}

// func (this Univalues) EncodeSimple() []byte {
// 	byteset := make([][]byte, len(this))
// 	worker := func(start, end, index int, args ...interface{}) {
// 		for i := start; i < end; i++ {
// 			byteset[i] = this[i].Encode()
// 		}
// 	}
// 	common.ParallelWorker(len(this), 6, worker)
// 	return codec.Byteset(byteset).Encode()
// }

// func (this Univalues) EncodeSimple(args interface{}) []byte {
// 	byteset := make([][]byte, len(this))
// 	for i := range this {
// 		byteset[i] = this[i].Encode()

// 	}
// 	return codec.Byteset(byteset).Encode()
// }

func (Univalues) Decode(bytes []byte) interface{} {
	if len(bytes) == 0 {
		return nil
	}

	buffers := [][]byte(codec.Byteset{}.Decode(bytes).(codec.Byteset))
	univalues := make([]ccurlcommon.UnivalueInterface, len(buffers))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			v := (&Univalue{}).Decode(buffers[i])
			univalues[i] = v.(ccurlcommon.UnivalueInterface)
		}
	}
	common.ParallelWorker(len(buffers), 6, worker)
	return Univalues(univalues)
}

// func (Univalues) DecodeV2(bytesset [][]byte, get func() interface{}, put func(interface{})) Univalues {
// 	univalues := make([]ccurlcommon.UnivalueInterface, len(bytesset))
// 	for i := range bytesset {
// 		v := get().(*Univalue)
// 		v.reclaimFunc = put
// 		v.Decode(bytesset[i])
// 		univalues[i] = v
// 	}
// 	return Univalues(univalues)
// }

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
