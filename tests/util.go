/*
 *   Copyright (c) 2023 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// Generate a random account, testing only
package ccurltest

import (
	"log"
	"math/rand"
	"time"

	rlp "github.com/ethereum/go-ethereum/rlp"
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

func CarolAccount() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	// rand.Seed(2)
	b := make([]rune, 40)
	for i := range b {
		b[i] = letters[3]
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
