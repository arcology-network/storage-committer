package mkproof

// "github.com/arcology-network/common-lib/merkle"

// type Convertbile merkle.Node
import (
	"fmt"

	"github.com/arcology-network/common-lib/common"
	ethdb "github.com/arcology-network/evm/ethdb"
)

type EthMerkleProof map[string][]byte

func (this EthMerkleProof) Put(key []byte, value []byte) error {
	keyS := fmt.Sprintf("%x", key)
	this[keyS] = value
	fmt.Printf("put key: %x, value: %x\n", key, value)
	return nil
}

func (this EthMerkleProof) Has(key []byte) (bool, error) {
	keyS := fmt.Sprintf("%x", key)
	_, ok := this[keyS]
	return ok, nil
}

func (this EthMerkleProof) Get(key []byte) ([]byte, error) {
	keyS := fmt.Sprintf("%x", key)
	val, ok := this[keyS]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return val, nil
}

func (this EthMerkleProof) ConvertProofs(subroot [][]byte, proofs [][][]byte) ethdb.KeyValueReader {
	common.Reverse(&proofs)
	for i := 0; i < len(proofs)-1; i++ {
		v := EthRlp{}.Encode(proofs[i])
		if err := this.Put(keybytesToHex(subroot[i]), v); err != nil {
			panic(err)
		}
	}
	return this
}

func keybytesToHex(str []byte) []byte {
	l := len(str)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = 16
	return nibbles
}
