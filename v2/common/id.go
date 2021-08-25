package common

import "math"

// User access control
const (
	USER_READABLE = iota
	USER_WRITABLE
	USER_UPDATABLE
)

const (
	MaxDepth uint8  = 12
	SYSTEM          = math.MaxInt32
	Root     string = "/"

	CommutativeMeta      uint8 = 100
	NoncommutativeInt64  uint8 = 101
	NoncommutativeString uint8 = 102
	NoncommutativeBigint uint8 = 103
	NoncommutativeBytes  uint8 = 104

	CommutativeInt64   uint8 = 105
	CommutativeString  uint8 = 106
	CommutativeBalance uint8 = 107
)

// func GetTypeID(value interface{}) (uint8, error) {
// 	if value != nil {
// 		name := reflect.TypeOf(value).String()
// 		switch name {
// 		/* Non Commutative */
// 		case "*commutative.Meta": // Meta
// 			return CommutativeMeta, nil
// 		case "*noncommutative.Bigint":
// 			return NoncommutativeBigint, nil
// 		case "*noncommutative.Int64":
// 			return NoncommutativeInt64, nil
// 		case "*noncommutative.String":
// 			return NoncommutativeString, nil
// 		case "*noncommutative.Bytes":
// 			return NoncommutativeBytes, nil

// 		/* Commutative */
// 		case "*commutative.Balance":
// 			return CommutativeBalance, nil
// 		case "*commutative.Int64":
// 			return CommutativeInt64, nil
// 		}
// 	}
// 	return uint8(reflect.Invalid), errors.New("Error: Unknown type ID !")
// }
