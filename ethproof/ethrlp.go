package mkproof

import (
	"log"

	rlp "github.com/arcology-network/evm/rlp"
)

type EthRlp struct{}

func (EthRlp) Encode(value ...interface{}) []byte {
	encoded, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Fatal("Error encoding data:", err)
	}
	return encoded
}

func (this EthRlp) Decode(buffer []byte) interface{} {
	var decoded []interface{}
	if err := rlp.DecodeBytes(buffer, &decoded); err != nil {
		log.Fatal("Error decoding data:", err)
	}
	return decoded
}
