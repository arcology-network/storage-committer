/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package ethstorage

import (
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
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
		addr:         ethcommon.BytesToAddress([]byte("3456")),
		StateAccount: *state,
		code:         []byte{1, 2, 3, 4},
	}
	buffer := acct.Encode()

	decodeAcct := (&Account{}).Decode(buffer)
	if state.Balance.Uint64() != decodeAcct.Balance.Uint64() {
		t.Error("Error: Blance mismatched!!")
	}
}
