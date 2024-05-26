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

package common

import (
	"math"
)

const (
	MAX_DEPTH uint8 = 12
	SYSTEM          = math.MaxInt32

	ETH10                              = "blcc://eth1.0/"
	ETH10_ACCOUNT_PREFIX               = ETH10 + "account/"
	ETH10_ACCOUNT_PREFIX_LENGTH        = len(ETH10_ACCOUNT_PREFIX)
	ETH10_ACCOUNT_LENGTH               = 42 // 40 hex digits + 0x
	ETH10_ACCOUNT_FULL_LENGTH          = ETH10_ACCOUNT_PREFIX_LENGTH + ETH10_ACCOUNT_LENGTH
	ETH10_STORAGE_PREFIX               = ETH10_ACCOUNT_PREFIX + "storage/"
	ETH10_STORAGE_PREFIX_LENGTH        = len(ETH10_STORAGE_PREFIX) + ETH10_ACCOUNT_LENGTH
	ETH10_STORAGE_NATIVE_PREFIX_LENGTH = ETH10_STORAGE_PREFIX_LENGTH + len("/native/")

	ETH10_FUNC_PROPERTY_PREFIX = "/func/"
)

const (
	UNKNOWN       uint8 = iota
	ETH_PATH_TYPE       // 1
	ACL_PATH_TYPE       // 2
)

var WARN_OUT_OF_LOWER_LIMIT string = "Warning: Out of the lower limit!"
var WARN_OUT_OF_UPPER_LIMIT string = "Warning: Out of the upper limit!"
var WARN_ACCESS_CONFLICT = "Warning: State access conflict detected!"
var WARN_EXEC_FAILED = "Warning: Execution execution failed!"
