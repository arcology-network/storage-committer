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

// AccountIndexer  avoids having duplicate addresses in the account list and dictionary.

package storage

import (
	"github.com/arcology-network/concurrenturl/importer"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type AccountUpdate struct {
	Key  string
	Addr ethcommon.Address
	Seqs []*importer.DeltaSequence
	Acct *Account
}
