package common

// User access control
const (
	USER_READABLE = iota
	USER_CREATABLE
	USER_UPDATABLE
)

const (
	MaxDepth uint8  = 12
	SYSTEM          = 0xffffffff
	Root     string = "/"

	NoncommutativeMeta   uint8 = 100
	NoncommutativeInt64  uint8 = 101
	NoncommutativeString uint8 = 102
	NoncommutativeBigint uint8 = 103
	NoncommutativeBytes  uint8 = 104

	CommutativeInt64   uint8 = 105
	CommutativeString  uint8 = 106
	CommutativeBalance uint8 = 107
)
