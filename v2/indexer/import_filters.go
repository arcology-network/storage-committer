package indexer

import (
	"strings"
)

type NonceFilter struct{}

func (this *NonceFilter) Is(offset int, key string) bool {
	return len(key) > offset && strings.LastIndex(key[offset:], "/storage/nonce") >= 0
}

type BalanceFilter struct{}

func (this *BalanceFilter) Is(offset int, key string) bool {
	return len(key) > offset && strings.LastIndex(key[offset:], "/balance") >= 0
}
