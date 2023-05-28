package common

import (
	"math"
)

const (
	MAX_DEPTH            uint8 = 12
	SYSTEM                     = math.MaxInt32
	ETH10_ACCOUNT_LENGTH       = 40
)

const (
	SUCCESSFUL = iota
	ERR_OUT_OF_LOWER_LIMIT
	ERR_OUT_OF_UPPER_LIMIT
	ERR_ACCESS_CONFLICT
	ERR_EXEC_FAILED
)
