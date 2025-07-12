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

package storagecommon

import (
	"encoding/hex"

	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func SponsoredGasPath(source evmcommon.Address) string {
	return ETH10_ACCOUNT_PREFIX + hexutil.Encode(source[:]) + "/sponsoredGas"
}

func PropertyPath(source evmcommon.Address) string {
	return ETH10_ACCOUNT_PREFIX + hexutil.Encode(source[:]) + "/" + FUNC_PATH
}

func FuncPath(source evmcommon.Address, sourceFun [4]byte) string {
	return PropertyPath(source) + hex.EncodeToString(sourceFun[:]) + "/"
}

func ExecutionParallelism(source evmcommon.Address, sourceFun [4]byte) string {
	return FuncPath(source, sourceFun) + EXECUTION_PARALLELISM
}

func ExceptPaths(source evmcommon.Address, sourceFun [4]byte) string {
	return FuncPath(source, sourceFun) + EXECUTION_EXCEPTED
}

func DeferrablePath(source evmcommon.Address, sourceFun [4]byte) string {
	return FuncPath(source, sourceFun) + DEFERRED
}

func RequiredPrepaidGasAmountPath(source evmcommon.Address, sourceFun [4]byte) string {
	return FuncPath(source, sourceFun) + REQUIRED_PREPAID_GAS_AMOUNT
}

// The path for the prepayers that are used to pay for the prepaid gas.
func PrepayersPath() string {
	return GAS_PREPAYERS
}
