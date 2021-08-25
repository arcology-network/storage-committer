package common

import (
	"errors"
	"math/big"
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
	this.syspaths[this.Eth10Account()+acct+"/balance"] = Syspath{[]bool{true, true, true}, CommutativeBalance, []*big.Int{big.NewInt(0), big.NewInt(0)}}
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

// If users are allowed to access the path
func (this *Platform) IsPermitted(path string, operation uint8) bool {
	if syspath, ok := this.syspaths[path]; ok {
		return syspath.Permissions[operation]
	}
	return false
}

// THe path on the control list
func (this *Platform) OnList(path string) bool {
	_, ok := this.syspaths[path]
	return ok
}
