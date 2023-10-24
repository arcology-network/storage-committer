package ccdb

type CodecIntf interface {
	Encoder(interface{}) func(interface{}) []byte
	Decoder(interface{}) func([]byte) (interface{}, error)
}
