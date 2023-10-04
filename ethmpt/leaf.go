package merklepatriciatrie

import (
	"fmt"

	"github.com/arcology-network/evm/crypto"
)

type LeafNode struct {
	Path  []Nibble
	Value []byte
	// cache *[]byte
}

func NewLeafNodeFromNibbleBytes(nibbles []byte, value []byte) (*LeafNode, error) {
	ns, err := FromNibbleBytes(nibbles)
	if err != nil {
		return nil, fmt.Errorf("could not leaf node from nibbles: %w", err)
	}

	return NewLeafNodeFromNibbles(ns, value), nil
}

func NewLeafNodeFromNibbles(nibbles []Nibble, value []byte) *LeafNode {
	return &LeafNode{
		Path:  nibbles,
		Value: value,
	}
}

func NewLeafNodeFromKeyValue(key, value string) *LeafNode {
	return NewLeafNodeFromBytes([]byte(key), []byte(value))
}

func NewLeafNodeFromBytes(key, value []byte) *LeafNode {
	return NewLeafNodeFromNibbles(FromBytes(key), value)
}

// func (l LeafNode) GetCached() *[]byte {
// 	return l.cache
// }

// func (l LeafNode) SetCached(this *Node, cache *[]byte) {
// 	v := (*this).(*LeafNode)
// 	(*v).cache = cache
// }

func (l LeafNode) Hash() []byte {
	return crypto.Keccak256(l.Serialize())
}

func (l LeafNode) Raw() []interface{} {
	path := ToBytes(ToPrefixed(l.Path, true))
	raw := []interface{}{path, l.Value}
	return raw
}

// Encode + Raw
func (l LeafNode) Serialize() []byte {
	return Serialize(l)
}
