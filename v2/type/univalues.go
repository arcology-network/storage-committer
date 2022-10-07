package ccurltype

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/HPISTechnologies/common-lib/codec"
	"github.com/HPISTechnologies/common-lib/common"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
)

type Univalues []ccurlcommon.UnivalueInterface

/* --------------------------------------------------------------- */
func (this Univalues) Encode() []byte {
	lengths := make([]uint32, len(this))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			if this[i] != nil {
				lengths[i] = this[i].(*Univalue).Size()
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
			this[i].(*Univalue).EncodeToBuffer(buffer[headerLen+offsets[i]:])
		}
	}
	common.ParallelWorker(len(this), 6, worker)
	return buffer
}

func (this Univalues) EncodeSimple() []byte {
	byteset := make([][]byte, len(this))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			byteset[i] = this[i].Encode()
		}
	}
	common.ParallelWorker(len(this), 6, worker)
	return codec.Byteset(byteset).Encode()
}

func (this Univalues) EncodeV2() [][]byte {
	byteset := make([][]byte, len(this))
	for i := range this {
		byteset[i] = this[i].Encode()
	}
	return byteset
}

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

func (Univalues) DecodeV2(bytesset [][]byte, get func() interface{}, put func(interface{})) Univalues {
	univalues := make([]ccurlcommon.UnivalueInterface, len(bytesset))
	for i := range bytesset {
		v := get().(*Univalue)
		v.reclaimFunc = put
		v.Decode(bytesset[i])
		univalues[i] = v
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

func (this Univalues) GetEncodedSize() [][]int {
	sizes := make([][]int, len(this))
	for i, v := range this {
		sizes[i] = v.(*Univalue).GetEncodedSize()
	}
	return sizes
}

func (this Univalues) Sort() {
	sort.SliceStable(this, func(i, j int) bool {
		lhs := (*(this[i].GetPath()))
		rhs := (*(this[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	//return this
}

func (this Univalues) IfContains(condition ccurlcommon.UnivalueInterface) bool {
	for _, v := range this {
		if (v).(*Univalue).EqualTransition(condition.(*Univalue)) {
			return true
		}
	}
	return false
}

// For debug only
func (this Univalues) Checksum() [32]byte {
	return sha256.Sum256(this.Encode())
}
