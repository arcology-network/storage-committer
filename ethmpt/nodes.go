package merklepatriciatrie

import "github.com/arcology-network/evm/rlp"

type Node interface {
	Hash() []byte // common.Hash
	Raw() []interface{}
	// GetCached() *[]byte
	// SetCached(*Node, *[]byte)
}

func Hash(node Node) []byte {
	if IsEmptyNode(node) {
		return EmptyNodeHash
	}
	return node.Hash()
}

func Serialize(node Node) []byte {
	var raw interface{}

	if IsEmptyNode(node) {
		raw = EmptyNodeRaw
	} else {
		raw = node.Raw()
	}

	// return EmptyNodeRaw

	rlp, err := rlp.EncodeToBytes(raw)
	if err != nil {
		panic(err)
	}

	return rlp
}
