package indexer

import (
	"strings"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
)

type Unsuccessful struct{}

func (Unsuccessful) Is(univ ccurlcommon.UnivalueInterface) bool {
	if int(univ.GetErrorCode()) == ccurlcommon.SUCCESSFUL {
		return false
	}
	return strings.HasSuffix(*univ.GetPath(), "/balance") ||
		strings.HasSuffix(*univ.GetPath(), "/nonce")
}
