package merklepatriciatrie

type Node interface {
	Hash() []byte // common.Hash
	Raw() []byte
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
	if node != nil {
		return node.Raw()
	}
	return EmptyNodeRaw
}
