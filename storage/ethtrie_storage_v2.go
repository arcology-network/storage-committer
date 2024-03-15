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
	"errors"
	"runtime"

	common "github.com/arcology-network/common-lib/common"

	"github.com/arcology-network/common-lib/exp/associative"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/storage-committer/univalue"
	"github.com/ethereum/go-ethereum/crypto"
)

// Precommit update the account state tries and the world state trie.
// It returns the hash of the world state trie.
func (this *EthDataStore) PrecommitV2(arg []*associative.Pair[*Account, []*univalue.Univalue]) [32]byte {
	this.dirtyAccounts = arg
	if len(this.dirtyAccounts) == 0 {
		return this.worldStateTrie.Hash() // No updates
	}

	// Need to check if this is necessary or could be moved to the import phase
	slice.Foreach(this.dirtyAccounts, func(_ int, pair **associative.Pair[*Account, []*univalue.Univalue]) {
		this.accountCache[(**pair).First.Address()] = (**pair).First // Add the account to the cache
	})

	slice.ParallelForeach(this.dirtyAccounts, runtime.NumCPU(), func(i int, account **associative.Pair[*Account, []*univalue.Univalue]) {
		keys, vals := univalue.Univalues((**account).Second).KVs() // Get all transitions under the same account
		(**account).First.UpdateAccountTrie(keys, vals)            // Update the account trie with the transitions
	})
	return this.WriteWorldTrie(this.dirtyAccounts)
}

// The WriteWorldTrie writes the updated accounts to the world trie.
func (this *EthDataStore) WriteWorldTrie(dirtyAccounts []*associative.Pair[*Account, []*univalue.Univalue]) [32]byte {
	encodedAddrs, encodedAcct := [][]byte{}, [][]byte{} // Encode the account key and values
	common.ParallelExecute(
		func() { // Account keys
			encodedAddrs = slice.Append(dirtyAccounts, func(_ int, update *associative.Pair[*Account, []*univalue.Univalue]) []byte {
				return crypto.Keccak256(update.First.addr[:]) // Hash the account address
			})
		},
		func() { // Encode the account content.
			encodedAcct = slice.Append(dirtyAccounts, func(_ int, update *associative.Pair[*Account, []*univalue.Univalue]) []byte {
				return update.First.Encode() // Encode the account
			})
		},
	)

	// Write the world tree and return the first error if any.
	errs := this.worldStateTrie.ParallelUpdate(encodedAddrs, encodedAcct)
	this.dbErr = errors.Join(this.dbErr, errors.Join(errs...))
	return this.worldStateTrie.Hash()

}
