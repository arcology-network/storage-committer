// Generate a random account, testing only
package ccurltest

import (
	"log"
	"math/rand"
	"time"

	rlp "github.com/arcology-network/evm/rlp"
)

func RandomAccount() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 40)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func AliceAccount() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	// rand.Seed(1)
	b := make([]rune, 40)
	for i := range b {
		b[i] = letters[1]
	}
	return string(b)
}

func BobAccount() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	// rand.Seed(2)
	b := make([]rune, 40)
	for i := range b {
		b[i] = letters[2]
	}
	return string(b)
}

func rlpEncoder(args ...interface{}) []byte {
	encoded, err := rlp.EncodeToBytes(args)
	if err != nil {
		log.Fatal("Error encoding data:", err)
	}
	return encoded
}
