package common

import (
	"math"
)

const (
	MAX_DEPTH            uint8 = 12
	SYSTEM                     = math.MaxInt32
	ETH10_ACCOUNT_LENGTH       = 40
)

var ERR_OUT_OF_LOWER_LIMIT string = "Error: Out of the lower limit!"
var ERR_OUT_OF_UPPER_LIMIT string = "Error: Out of the upper limit!"
var ERR_ACCESS_CONFLICT = "Error: State access conflict detected!"
var ERR_EXEC_FAILED = "Error: Execution execution failed!"
