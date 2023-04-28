package commutative

import (
	codec "github.com/arcology-network/common-lib/codec"
	orderedset "github.com/arcology-network/common-lib/container/set"
)

func (this *MetaDelta) Encode(processors ...interface{}) []byte {
	buffer := make([]byte, this.Size()) //  no need to send the committed keys
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *MetaDelta) Size() uint32 {
	return codec.Stringset([][]string{this.addDict.Keys(), this.delDict.Keys()}).Size()
}

func (this *MetaDelta) EncodeToBuffer(buffer []byte, processors ...interface{}) int {
	return codec.Stringset([][]string{this.addDict.Keys(), this.delDict.Keys()}).EncodeToBuffer(buffer)
}

func (this *MetaDelta) Decode(buffer []byte) interface{} {
	fields := codec.Stringset{}.Decode(buffer).(codec.Stringset)
	this = &MetaDelta{
		addDict: orderedset.NewOrderedSet(fields[0]),
		delDict: orderedset.NewOrderedSet(fields[1]),
	}
	return this
}
