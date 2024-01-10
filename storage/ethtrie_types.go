package storage

import (
	"encoding/hex"
	"fmt"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	// ethapi "github.com/ethereum/go-ethereum/internal/ethapi"
)

// ------------------------Types copied from ETH------------------------
type AccountResult struct {
	Address      ethcommon.Address `json:"address"` // Address of the account
	AccountProof []string          `json:"accountProof"`
	Balance      *hexutil.Big      `json:"balance"`
	CodeHash     ethcommon.Hash    `json:"codeHash"`
	Nonce        hexutil.Uint64    `json:"nonce"`
	StorageHash  ethcommon.Hash    `json:"storageHash"`
	StorageProof []StorageResult   `json:"storageProof"`
}

type StorageResult struct {
	Key   string       `json:"key"`
	Value *hexutil.Big `json:"value"`
	Proof []string     `json:"proof"`
}

type proofList [][]byte

func (n *proofList) Put(key []byte, value []byte) error {
	*n = append(*n, value)
	return nil
}

func (n *proofList) Delete(key []byte) error {
	panic("not supported")
}

// toHexSlice creates a slice of hex-strings based on []byte.
func toHexSlice(b [][]byte) []string {
	r := make([]string, len(b))
	for i := range b {
		r[i] = hexutil.Encode(b[i])
	}
	return r
}

func decodeHash(s string) (ethcommon.Hash, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	if (len(s) & 1) > 0 {
		s = "0" + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("hex string invalid")
	}
	if len(b) > 32 {
		return ethcommon.Hash{}, fmt.Errorf("hex string too long, want at most 32 bytes")
	}
	return ethcommon.BytesToHash(b), nil
}

// ------------------------End of types copied from ETH------------------------
