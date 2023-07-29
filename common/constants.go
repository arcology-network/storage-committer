package common

import (
	"math"
)

const (
	MAX_DEPTH uint8 = 12
	SYSTEM          = math.MaxInt32

	ETH10                       = "blcc://eth1.0/"
	ETH10_ACCOUNT_PREFIX        = ETH10 + "account/"
	ETH10_ACCOUNT_PREFIX_LENGTH = len(ETH10_ACCOUNT_PREFIX)
	ETH10_ACCOUNT_LENGTH        = 40
	ETH10_ACCOUNT_FULL_LENGTH   = ETH10_ACCOUNT_PREFIX_LENGTH + ETH10_ACCOUNT_LENGTH
)

var WARN_OUT_OF_LOWER_LIMIT string = "Warning: Out of the lower limit!"
var WARN_OUT_OF_UPPER_LIMIT string = "Warning: Out of the upper limit!"
var WARN_ACCESS_CONFLICT = "Warning: State access conflict detected!"
var WARN_EXEC_FAILED = "Warning: Execution execution failed!"
