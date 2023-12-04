package storage

import (
	"math/big"
	"testing"

	ethcommon "github.com/arcology-network/evm/common"
	ethtypes "github.com/arcology-network/evm/core/types"
	"github.com/arcology-network/evm/rlp"
)

func TestAccountCode(t *testing.T) {
	state := &ethtypes.StateAccount{
		Nonce:    111,
		Balance:  big.NewInt(99),
		Root:     ethcommon.Hash{}, // merkle root of the storage trie
		CodeHash: []byte{9, 8, 0, 7},
	}

	if encoded, err := rlp.EncodeToBytes(state); err != nil {
		t.Error("Error: Should be empty!!")
	} else {
		var acct ethtypes.StateAccount
		rlp.DecodeBytes(encoded, &acct)
		if state.Balance.Uint64() != acct.Balance.Uint64() {
			t.Error("Error: Blance mismatched!!")
		}
	}

	acct := &Account{
		addr:         "3456",
		StateAccount: *state,
		code:         []byte{1, 2, 3, 4},
	}
	buffer := acct.Encode()

	decodeAcct := (&Account{}).Decode(buffer)
	if state.Balance.Uint64() != decodeAcct.Balance.Uint64() {
		t.Error("Error: Blance mismatched!!")
	}
}
