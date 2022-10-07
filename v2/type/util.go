package ccurltype

import (
	"math/rand"
	"strings"
	"time"

	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
)

func CreateUnivalueForTest(transitType uint8, vType uint8, tx uint32, path string, reads, writes uint32, value interface{}, preexists, composite bool) *Univalue {
	return &Univalue{
		transitType: transitType,
		vType:       vType,
		tx:          tx,
		path:        &path,
		reads:       reads,
		writes:      writes,
		value:       value,
		preexists:   preexists,
		composite:   composite,
		reserved:    nil,
	}
}

// Generate a random account, testing only
func RandomAccount() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 40)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func SetInvariate(trans []ccurlcommon.UnivalueInterface, name string) {
	for i := 0; i < len(trans); i++ {
		if strings.Contains(*(trans[i].GetPath()), name) {
			trans[i].SetTransitionType(ccurlcommon.INVARIATE_TRANSITIONS)
		}
	}
}
