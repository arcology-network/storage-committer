package common

import (
	"errors"

	"github.com/holiman/uint256"
)

const (
	ETH10_ACCOUNT_LENGTH = 40
)

type Syspath struct {
	Permissions []bool
	ID          uint8
	Default     interface{}
}

type Platform struct {
	syspaths map[string]Syspath
}

func NewPlatform() *Platform {
	return &Platform{
		syspaths: make(map[string]Syspath),
	}
}

func (this *Platform) Eth10() string        { return "blcc://eth1.0/" }
func (this *Platform) Eth10Account() string { return this.Eth10() + "account/" }
func Eth10AccountShard(numOfShard int, key string) int {
	if len(key) < 24 {
		panic("Invalid eth1.0 account shard key: " + key)
	}
	return (hex2int(key[22])*16 + hex2int(key[23])) % numOfShard
}

func (this *Platform) RootLength() int { return len(this.Eth10Account()) + 20 }

func hex2int(c byte) int {
	if c >= 'a' {
		return int(c-'a') + 10
	} else {
		return int(c - '0')
	}
}

// Get ths builtin paths
func (this *Platform) Builtin(platform string, acct string) ([]string, map[string]Syspath, error) {
	switch platform {
	case this.Eth10():
		break
	default:
		return nil, nil, errors.New("Error: Unknown platform !")
	}

	this.syspaths[this.Eth10Account()+acct+"/"] = Syspath{[]bool{true, false, false}, CommutativeMeta, this.Eth10Account() + acct + "/"}
	this.syspaths[this.Eth10Account()+acct+"/code"] = Syspath{[]bool{true, true, true}, NoncommutativeBytes, []byte{}}
	this.syspaths[this.Eth10Account()+acct+"/nonce"] = Syspath{[]bool{true, true, true}, CommutativeInt64, int64(0)}
	this.syspaths[this.Eth10Account()+acct+"/balance"] = Syspath{[]bool{true, true, true}, CommutativeUint256, uint256.NewInt(0)}
	this.syspaths[this.Eth10Account()+acct+"/defer/"] = Syspath{[]bool{true, true, true}, CommutativeMeta, this.Eth10Account() + acct + "/defer/"}
	this.syspaths[this.Eth10Account()+acct+"/storage/"] = Syspath{[]bool{true, false, false}, CommutativeMeta, this.Eth10Account() + acct + "/storage/"}
	this.syspaths[this.Eth10Account()+acct+"/storage/containers/"] = Syspath{[]bool{true, true, false}, CommutativeMeta, this.Eth10Account() + acct + "/storage/containers/"}
	this.syspaths[this.Eth10Account()+acct+"/storage/native/"] = Syspath{[]bool{true, true, false}, CommutativeMeta, this.Eth10Account() + acct + "/storage/native/"}
	this.syspaths[this.Eth10Account()+acct+"/storage/containers/!/"] = Syspath{[]bool{true, true, false}, CommutativeMeta, this.Eth10Account() + acct + "/storage/containers/!/"}

	paths := []string{
		this.Eth10Account() + acct + "/",
		this.Eth10Account() + acct + "/code",
		this.Eth10Account() + acct + "/nonce",
		this.Eth10Account() + acct + "/balance",
		this.Eth10Account() + acct + "/defer/",
		this.Eth10Account() + acct + "/storage/",
		this.Eth10Account() + acct + "/storage/containers/",
		this.Eth10Account() + acct + "/storage/native/",
		this.Eth10Account() + acct + "/storage/containers/!/",
	}

	return paths, this.syspaths, nil
}

// The path on the control list
func (this *Platform) IsSysPath(path string) bool {
	_, ok := this.syspaths[path]
	return ok || path == this.Eth10() || path == this.Eth10Account()
}
