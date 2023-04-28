package common

import "math"

// User access control
// const (
// 	USER_READABLE = iota
// 	USER_CREATABLE
// 	USER_UPDATABLE
// )

const (
	MaxDepth uint8  = 12
	SYSTEM          = math.MaxInt32
	Root     string = "/"

	// Commutative{}.Path()    uint8 = 100
	// Commutative{}.Int64()   uint8 = 101
	// CommutativeUint64  uint8 = 102
	// Commutative{}.Uint256() uint8 = 103

	// NonCommutative{}.Int64()  uint8 = 104
	// NonCommutative{}.String() uint8 = 105
	// NonCommutative{}.Bigint() uint8 = 106
	// NonCommutative{}.Bytes()  uint8 = 107

	VARIATE_TRANSITIONS   uint8 = 0
	INVARIATE_TRANSITIONS uint8 = 1
)

// const (
// 	WRITE   uint8 = 0
// 	REWRITE uint8 = 12
// )

// func IsTypeValid(typeID uint8) bool {
// 	return typeID >= Commutative{}.Path() && typeID <= NonCommutative{}.Bytes()
// }

type Commutative struct{}

func (this Commutative) Path() uint8    { return 100 }
func (this Commutative) Int64() uint8   { return 101 }
func (this Commutative) Uint64() uint8  { return 102 }
func (this Commutative) Uint256() uint8 { return 103 }

type NonCommutative struct{}

func (this NonCommutative) Int64() uint8  { return 104 }
func (this NonCommutative) String() uint8 { return 105 }
func (this NonCommutative) Bigint() uint8 { return 106 }
func (this NonCommutative) Bytes() uint8  { return 107 }
