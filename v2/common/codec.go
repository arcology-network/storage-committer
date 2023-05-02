package common

// type Generic struct {
// 	value interface{}
// 	delta interface{}
// 	min   interface{}
// 	max   interface{}
// }

// func (this *Generic) Value() interface{} { return this.value }
// func (this *Generic) Delta() interface{} { return this.delta }
// func (this *Generic) Min() interface{}   { return this.min }
// func (this *Generic) Max() interface{}   { return this.max }

// func (this *Generic) HeaderSize() uint32 {
// 	return 5 * codec.UINT32_LEN //static size only , no header needed,
// }

// func (this *Generic) Size() uint32 {
// 	return this.HeaderSize() +
// 		common.IfThen(this.Value() != nil, uint32(8), 0) +
// 		common.IfThen(this.Delta() != nil, uint32(8), 0) +
// 		common.IfThen(this.Min() != nil, uint32(8), 0) +
// 		common.IfThen(this.Max() != nil, uint32(8), 0)
// }

// func (this *Generic) Encode() []byte {
// 	buffer := make([]byte, this.Size())
// 	offset := codec.Encoder{}.FillHeader(
// 		buffer,
// 		[]uint32{
// 			common.IfThen(this.value != nil, uint32(8), 0),
// 			common.IfThen(this.delta != nil, uint32(8), 0),
// 			common.IfThen(this.min != nil, uint32(8), 0),
// 			common.IfThen(this.max != nil, uint32(8), 0),
// 		},
// 	)
// 	this.EncodeToBuffer(buffer[offset:])
// 	return buffer
// }

// func (this *Generic) EncodeToBuffer(buffer []byte, ) int {
// 	offset := common.IfThenDo1st(this.value != nil, func() int { return this.value.()EncodeToBuffer(buffer) }, 0)
// 	offset += common.IfThenDo1st(this.delta != nil, func() int { return this.delta.EncodeToBuffer(buffer[offset:]) }, 0)
// 	offset += common.IfThenDo1st(this.min != nil, func() int { return this.min.EncodeToBuffer(buffer[offset:]) }, 0)
// 	offset += common.IfThenDo1st(this.max != nil, func() int { return this.max.EncodeToBuffer(buffer[offset:]) }, 0)
// 	return offset
// }
