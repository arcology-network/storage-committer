package concurrenturl

import (
	"math"

	common "github.com/arcology-network/common-lib/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
)

const (
	MAX_DEPTH            uint8 = 12
	SYSTEM                     = math.MaxInt32
	ETH10_ACCOUNT_LENGTH       = 40
)

type Platform struct {
	syspaths map[string]uint8
}

func NewPlatform() *Platform {
	return &Platform{
		map[string]uint8{
			"/":                      commutative.PATH,
			"/code":                  noncommutative.BYTES,
			"/nonce":                 commutative.UINT64,
			"/balance":               commutative.UINT256,
			"/storage/":              commutative.PATH,
			"/storage/container/":    commutative.PATH,
			"/storage/native/":       commutative.PATH,
			"/storage/native/local/": commutative.PATH,
		},
	}
}

func (this *Platform) Eth10() string        { return "blcc://eth1.0/" }
func (this *Platform) Eth10Account() string { return this.Eth10() + "account/" }

func (this *Platform) Eth10AccountLength() int {
	return len(this.Eth10()+"account/") + ETH10_ACCOUNT_LENGTH
}

func (this *Platform) GetAccountAddr(path string) string {
	length := this.Eth10AccountLength()
	return common.IfThenDo1st(len(path) >= length, func() string { return path[:length] }, "")
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
	paths, typeIds := common.MapKVs(this.syspaths)
	common.SortBy1st(paths, typeIds, func(lhv, rhv string) bool { return lhv < rhv })

	for i, path := range paths {
		paths[i] = this.Eth10Account() + acct + path
	}
	return paths, typeIds
}

// These paths won't keep the sub elements
func (this *Platform) IsSysPath(path string) bool {
	if len(path) <= this.Eth10AccountLength() {
		return path == this.Eth10() || path == this.Eth10Account()
	}

	subPath := path[this.Eth10AccountLength():] // Removed the shared part
	_, ok := this.syspaths[subPath]
	return ok
}

func (this *Platform) GetSysPaths() []string {
	return common.MapKeys(this.syspaths)
}

func (this *Platform) Builtins(acct string, idx int) string {
	paths, _ := common.MapKVs(this.syspaths)
	return this.Eth10Account() + acct + paths[idx]
}
