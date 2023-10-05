package merklepatriciatrie

import (
	"github.com/arcology-network/evm/crypto"
	"github.com/arcology-network/evm/rlp"
)

var extensionCounter int

type ExtensionNode struct {
	Path  []Nibble
	Next  Node
	cache *[]byte
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

func (e ExtensionNode) Raw() []byte {
	hashes := make([]interface{}, 2)
	hashes[0] = ToBytes(ToPrefixed(e.Path, false))

	if cache := Serialize(e.Next); len(cache) >= 32 {
		hashes[1] = crypto.Keccak256(cache) //e.Next.Hash()
	} else {
		hashes[1] = cache //e.Next.Raw()
	}

	rlp, err := rlp.EncodeToBytes(hashes)
	if err != nil {
		panic(err)
	}
	return rlp
	// return hashes
}

// func (l ExtensionNode) GetCached() *[]byte {
// 	return l.Next.GetCached()
// }

// func (l ExtensionNode) SetCached(this *Node, cache *[]byte) {
// 	v := (*this).(*ExtensionNode)
// 	(v).cache = cache

// 	// l.Next.SetCached(this, cache)
// }

func (e ExtensionNode) Serialize() []byte {
	return Serialize(e)
}
