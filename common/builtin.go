package common

import (
	"errors"
)

type Platform struct {
	syspaths map[string][]bool
}

func NewPlatform() *Platform {
	return &Platform{
		syspaths: make(map[string][]bool),
	}
}

const (
	BASE_URL         = "blcc://eth1.0/"
	ACCOUNT_BASE_URL = BASE_URL + "accounts/"
)

func (this *Platform) Eth10() string { return ACCOUNT_BASE_URL }

// Get builtin path
func (this *Platform) Builtin(platform string, acct string) ([]string, error) {
	switch platform {
	case ACCOUNT_BASE_URL:
		break
	default:
		return nil, errors.New("Error: Unknown platform !")
	}

	this.syspaths[platform+acct+"/"] = []bool{true, false, false}
	this.syspaths[platform+acct+"/code"] = []bool{true, true, false}
	this.syspaths[platform+acct+"/nonce"] = []bool{true, true, true}
	this.syspaths[platform+acct+"/balance"] = []bool{true, true, true}
	this.syspaths[platform+acct+"/defer/"] = []bool{true, true, true}
	this.syspaths[platform+acct+"/storage/"] = []bool{true, false, false}
	this.syspaths[platform+acct+"/storage/containers/"] = []bool{true, true, false}
	this.syspaths[platform+acct+"/storage/native/"] = []bool{true, true, false}
	this.syspaths[platform+acct+"/storage/containers/!/"] = []bool{true, true, false}

	return []string{
		platform + acct + "/",
		platform + acct + "/code",
		platform + acct + "/nonce",
		platform + acct + "/balance",
		platform + acct + "/defer/",
		platform + acct + "/storage/",
		platform + acct + "/storage/containers/",
		platform + acct + "/storage/native/",
		platform + acct + "/storage/containers/!/",
	}, nil
}

// If users are allowed to access the path
func (this *Platform) IsPermissible(path string, operation uint8) bool {
	if syspaths := this.syspaths[path]; syspaths != nil {
		return syspaths[operation]
	}
	return false
}

// THe path on the control list
func (this *Platform) OnControlList(path string) bool {
	return this.syspaths[path] != nil
}
