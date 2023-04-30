package common

type NonCommutative struct{}

func (this NonCommutative) Int64() uint8  { return 104 }
func (this NonCommutative) String() uint8 { return 105 }
func (this NonCommutative) Bigint() uint8 { return 106 }
func (this NonCommutative) Bytes() uint8  { return 107 }
