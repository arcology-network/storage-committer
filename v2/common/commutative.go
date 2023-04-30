package common

type Commutative struct{}

func (this Commutative) Path() uint8    { return 100 }
func (this Commutative) Int64() uint8   { return 101 }
func (this Commutative) Uint64() uint8  { return 102 }
func (this Commutative) Uint256() uint8 { return 103 }
