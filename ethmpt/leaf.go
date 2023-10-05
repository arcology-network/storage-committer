package merklepatriciatrie

import (
	"fmt"

	ethrlp "github.com/arcology-network/concurrenturl/ethrlp"
	"github.com/arcology-network/evm/crypto"
)

type LeafNode struct {
	Path  []Nibble
	Value []byte
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

func (l LeafNode) Hash() []byte {
	return crypto.Keccak256(l.Raw())
}

func (l LeafNode) Raw() []byte {
	path := ToBytes(ToPrefixed(l.Path, true))
	rlp := ethrlp.Bytes{}.Encode([][]byte{path, l.Value})
	return rlp
}

// Encode + Raw
func (l LeafNode) Serialize() []byte {
	return Serialize(l)
}
