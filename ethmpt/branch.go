package merklepatriciatrie

import (
	"github.com/arcology-network/evm/crypto"

	ethrlp "github.com/arcology-network/concurrenturl/ethrlp"
)

type BranchNode struct {
	Branches [16]Node
	Value    []byte
	cache    *[]byte
}

func NewBranchNode() *BranchNode {
	// branchCounter++
	// fmt.Println(branchCounter)

	return &BranchNode{
		Branches: [16]Node{},
	}
}

func (b BranchNode) Hash() []byte {
	return crypto.Keccak256(b.Serialize())
}

func (b *BranchNode) SetBranch(nibble Nibble, node Node) {
	b.Branches[int(nibble)] = node
}

func (b *BranchNode) RemoveBranch(nibble Nibble) {
	b.Branches[int(nibble)] = nil
}

func (b *BranchNode) SetValue(value []byte) {
	b.Value = value
}

func (b *BranchNode) RemoveValue() {
	b.Value = nil
}

func (b BranchNode) Raw() []byte {
	hashes := make([][]byte, 17)
	for i := 0; i < 16; i++ {
		if b.Branches[i] == nil {
			hashes[i] = EmptyNodeRaw
		} else {
			node := b.Branches[i]
			if hashes[i] = node.Raw(); len(hashes[i]) >= 32 {
				hashes[i] = crypto.Keccak256(hashes[i])
			}
			// } else {
			// 	// if node can be serialized to less than 32 bits, then
			// 	// use Serialized directly.
			// 	// it has to be ">=", rather than ">",
			// 	// so that when deserialized, the content can be distinguished
			// 	// by length
			// 	hashes[i] = serialized //node.Raw()
			// }
		}
	}

	hashes[16] = b.Value

	rlp := ethrlp.Bytes{}.Encode(hashes)
	return rlp
}

func (b BranchNode) Serialize() []byte {
	return Serialize(b)
}

func (b BranchNode) HasValue() bool {
	return b.Value != nil
}

func (b BranchNode) SetCached(this *Node, cache *[]byte) {
	v := (*this).(BranchNode)
	(v).cache = cache
}

func (b BranchNode) GetCached() *[]byte {
	return b.cache
}
