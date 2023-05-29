package indexer

import (
	"strings"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type IPCTransition struct{ interfaces.Univalue }

func (IPCTransition) From(v interfaces.Univalue) interface{} {
	if v == nil ||
		v.IsReadOnly() ||
		(v.Value() == nil && !v.Preexist()) { // Del Non Exist
		return nil
	}

	if v.Value() == nil { // Entry deletion
		return v
	}

	if v.GetErrorCode() != ccurlcommon.SUCCESSFUL &&
		strings.HasSuffix(*v.GetPath(), "/balance") &&
		strings.HasSuffix(*v.GetPath(), "/nonce") {
		return nil
	}

	typed := v.Value().(interfaces.Type)
	typed = typed.New(
		nil,
		typed.Delta(),
		typed.DeltaSign(),
		typed.Min(),
		typed.Max(),
	).(interfaces.Type)

	return v.New(
		v.GetUnimeta(),
		typed,
		[]byte{},
		v.GetErrorCode(),
	)
}
