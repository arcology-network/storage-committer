package common

import "math"

// User access control
const (
	USER_READABLE = iota
	USER_CREATABLE
	USER_UPDATABLE
)

const (
	MaxDepth uint8  = 12
	SYSTEM          = math.MaxInt32
	Root     string = "/"

	CommutativeMeta    uint8 = 100
	CommutativeInt64   uint8 = 101
	CommutativeUint64  uint8 = 102
	CommutativeUint256 uint8 = 103

	NoncommutativeInt64  uint8 = 104
	NoncommutativeString uint8 = 105
	NoncommutativeBigint uint8 = 106
	NoncommutativeBytes  uint8 = 107

	VARIATE_TRANSITIONS   uint8 = 0
	INVARIATE_TRANSITIONS uint8 = 1
)

const (
	WRITE   uint8 = 0
	REWRITE uint8 = 12
)

func IsTypeValid(typeID uint8) bool {
	return typeID >= CommutativeMeta && typeID <= NoncommutativeBytes
}
