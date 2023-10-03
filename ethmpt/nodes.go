package merklepatriciatrie

type Node interface {
	Hash() []byte // common.Hash
	Raw() []interface{}
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
	// rlp, err := rlp.EncodeToBytes(raw)
	// if err != nil {
	// 	panic(err)
	// }

	if _, ok := raw.(*Node); ok {

	}
	v := [32]byte{}
	return v[:]
	// return rlp

}
