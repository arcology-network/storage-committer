package merklepatriciatrie

import (
	"github.com/arcology-network/evm/crypto"
)

var extensionCounter int

type ExtensionNode struct {
	Path []Nibble
	Next Node
}

func NewExtensionNode(nibbles []Nibble, next Node) *ExtensionNode {
	// extensionCounter++
	// fmt.Println("extensionCounter:", extensionCounter)

	return &ExtensionNode{
		Path: nibbles,
		Next: next,
	}
}

func (e ExtensionNode) Hash() []byte {
	return crypto.Keccak256(e.Serialize())
}

func (e ExtensionNode) Raw() []interface{} {
	hashes := make([]interface{}, 2)
	hashes[0] = ToBytes(ToPrefixed(e.Path, false))
	if len(Serialize(e.Next)) >= 32 {
		hashes[1] = e.Next.Hash()
	} else {
		hashes[1] = e.Next.Raw()
	}
	return hashes
}

func (e ExtensionNode) Serialize() []byte {
	return Serialize(e)
}
