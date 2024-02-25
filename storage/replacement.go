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

package storage

import (
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	"github.com/ethereum/go-ethereum/crypto"
)

// Update the account trie
func (this *EthDataStore) WriteEthTries(updates interface{}) [32]byte {
	if updates == nil {
		return this.worldStateTrie.Hash()
	}

	dirties := updates.([]*AccountUpdate)
	if len(dirties) == 0 {
		return this.worldStateTrie.Hash()
	}

	// Precommit the changes to the accounts and update the account storage trie.
	array.ParallelForeach(dirties, 16, func(idx int, acct **AccountUpdate) {
		((*acct).Acct).ApplyAccounts(*acct)
	})

	// Update the to account cache.
	array.Foreach(dirties, func(_ int, acct **AccountUpdate) {
		this.accounts[(*acct).Acct.addr] = (*acct).Acct
	})

	// Encode the account addresses.
	encodedAddrs, encodedVals := [][]byte{}, [][]byte{}
	common.ParallelExecute(
		func() { // Account keys
			encodedAddrs = array.Append(dirties, func(_ int, update *AccountUpdate) []byte {
				return crypto.Keccak256(update.Acct.addr[:])
			})
		},
		func() { // Encode the account content.
			encodedVals = array.Append(dirties, func(_ int, update *AccountUpdate) []byte {
				return update.Acct.Encode()
			})
		},
	)

	// Write the world tree and return the first error if any.
	errs := this.worldStateTrie.ParallelUpdate(encodedAddrs, encodedVals) // Encoded accounts
	if _, err := array.FindFirstIf(errs, func(err error) bool { return err != nil }); err != nil {
		panic("Error in updating the trie: " + (*err).Error())
	}

	// ================================Debug only=======================================
	// for _, k := range encodedAddrs {
	// 	if acctBuffer, err := this.worldStateTrie.Get([]byte(k)); err != nil || len(acctBuffer) == 0 {
	// 		panic("Error in updating the trie failed to retrieve the account: " + string(k))
	// 	}
	// }
	return this.worldStateTrie.Hash()
}
