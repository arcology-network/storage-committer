package common

const (
	ETH10_ACCOUNT_LENGTH = 40
)

type Platform struct {
	syspaths map[string]uint8
}

func NewPlatform() *Platform {
	return &Platform{
		map[string]uint8{
			"/":                    Commutative{}.Path(),
			"/code":                NonCommutative{}.Bytes(),
			"/nonce":               Commutative{}.Uint64(),
			"/balance":             Commutative{}.Uint256(),
			"/storage/":            Commutative{}.Path(),
			"/storage/containers/": Commutative{}.Path(),
			"/storage/native/":     Commutative{}.Path(),
		},
	}
}

func (this *Platform) Eth10() string        { return "blcc://eth1.0/" }
func (this *Platform) Eth10Account() string { return this.Eth10() + "account/" }
func (this *Platform) Eth10AccountLenght() int {
	return len(this.Eth10()+"account/") + ETH10_ACCOUNT_LENGTH
}

func Eth10AccountShard(numOfShard int, key string) int {
	if len(key) < 24 {
		panic("Invalid eth1.0 account shard key: " + key)
	}
	return (hex2int(key[22])*16 + hex2int(key[23])) % numOfShard
}

func (this *Platform) RootLength() int { return len(this.Eth10Account()) + ETH10_ACCOUNT_LENGTH }

func hex2int(c byte) int {
	if c >= 'a' {
		return int(c-'a') + 10
	} else {
		return int(c - '0')
	}
}

// Get ths builtin paths
func (this *Platform) GetBuiltins(acct string) ([]string, []uint8) {
	paths := make([]string, len(this.syspaths))
	typeIds := make([]uint8, len(this.syspaths))
	i := 0
	for k, v := range this.syspaths {
		paths[i] = this.Eth10Account() + acct + k
		typeIds[i] = v
		i++
	}
	return paths, typeIds
}

// These paths won't keep the sub elements
func (this *Platform) IsSysPath(path string) bool {
	if len(path) <= this.Eth10AccountLenght() {
		return path == this.Eth10() || path == this.Eth10Account()
	}

	subPath := path[this.Eth10AccountLenght():]
	_, ok := this.syspaths[subPath]
	return ok
}
