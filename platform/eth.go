/*
 *   Copyright (c) 2023 Arcology Network

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

package ethplatform

import (
	"strings"

	common "github.com/arcology-network/common-lib/common"
	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	noncommutative "github.com/arcology-network/storage-committer/type/noncommutative"
)

type Platform struct {
	syspaths map[string]uint8
}

func NewPlatform() *Platform {
	return &Platform{
		map[string]uint8{
			"/":        commutative.PATH,
			"/code":    noncommutative.BYTES,
			"/nonce":   commutative.UINT64,
			"/balance": commutative.UINT256,

			// Arcology specific paths
			"/sponsoredGas": commutative.UINT256, // Gas reserved
			// "/sponsor/":           commutative.PATH,    // Sponsor account for execution
			"/func/":              commutative.PATH,
			"/storage/":           commutative.PATH,
			"/storage/container/": commutative.PATH, // Container storage
			"/storage/native/":    commutative.PATH, // Native storage
		},
	}
}

func ETH10AccountShard(numOfShard int, key string) int {
	if len(key) < 24 {
		panic("Invalid eth1.0 account shard key: " + key)
	}
	return (hex2int(key[22])*16 + hex2int(key[23])) % numOfShard
}

// func (this *Platform) RootLength() int { return len(this.stgcommon.ETH10Account()) + 40 }

func hex2int(c byte) int {
	if c >= 'a' {
		return int(c-'a') + 10
	} else {
		return int(c - '0')
	}
}

// Get ths builtin paths
func (this *Platform) GetBuiltins(acct string) ([]string, []uint8) {
	paths, typeIds := common.MapKVs(this.syspaths)
	slice.SortBy1st(paths, typeIds, func(lhv, rhv string) bool { return lhv < rhv })

	for i, path := range paths {
		paths[i] = stgcommon.ETH10_ACCOUNT_PREFIX + acct + path
	}
	return paths, typeIds
}

// These paths won't keep the sub elements
func (this *Platform) IsSysPath(path string) bool {
	if len(path) <= stgcommon.ETH10_ACCOUNT_FULL_LENGTH {
		return path == stgcommon.ETH10 || path == stgcommon.ETH10_ACCOUNT_PREFIX
	}

	subPath := path[stgcommon.ETH10_ACCOUNT_FULL_LENGTH:] // Removed the shared part
	_, ok := this.syspaths[subPath]
	return ok
}

func (this *Platform) GetSysPaths() []string {
	return mapi.Keys(this.syspaths)
}

func (this *Platform) Builtins(acct string, idx int) string {
	paths, _ := common.MapKVs(this.syspaths)
	return stgcommon.ETH10_ACCOUNT_PREFIX + acct + paths[idx]
}

func ParseAccountAddr(acct string) (string, string, string) {
	if len(acct) < stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH+stgcommon.ETH10_ACCOUNT_LENGTH {
		return acct, "", ""
	}
	return acct[:stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH],
		acct[stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH : stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH+stgcommon.ETH10_ACCOUNT_LENGTH],
		acct[stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH+stgcommon.ETH10_ACCOUNT_LENGTH:]
}

func GetAccountAddr(acct string) string {
	if len(acct) < stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH+stgcommon.ETH10_ACCOUNT_LENGTH {
		return acct
	}
	return acct[stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH : stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH+stgcommon.ETH10_ACCOUNT_LENGTH]
}

func GetPathUnder(key, prefix string) string {
	if len(key) > stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH+stgcommon.ETH10_ACCOUNT_LENGTH {
		subKey := key[stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH+stgcommon.ETH10_ACCOUNT_LENGTH:]
		if subKey != prefix && strings.HasPrefix(subKey, prefix) {
			return subKey[len(prefix):]
		}
	}
	return ""
}

// IsEthPath checks if the path is an eth path, some paths are not Arcology only.
func IsEthPath(path string) bool {
	return !strings.HasSuffix(path, "container/")
}
