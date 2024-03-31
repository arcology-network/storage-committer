/*
 *   Copyright (c) 2024 Arcology Network

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

package statestore

// import (
// 	"reflect"

// 	"github.com/arcology-network/storage-committer/noncommutative"
// 	platform "github.com/arcology-network/storage-committer/platform"

// 	"github.com/arcology-network/storage-committer/commutative"
// 	univalue "github.com/arcology-network/storage-committer/univalue"
// )

// // CreateNewAccount creates a new account in the write cache.
// // It returns the transitions and an error, if any.
// func CreateNewAccount(tx uint32, acct string, this interface {
// 	IfExists(string) bool
// 	Write(uint32, string, interface{}) (interface{}, error)
// }) ([]*univalue.Univalue, error) {
// 	paths, typeids := platform.NewPlatform().GetBuiltins(acct)

// 	transitions := []*univalue.Univalue{}
// 	for i, path := range paths {
// 		var v interface{}
// 		switch typeids[i] {
// 		case commutative.PATH: // Path
// 			v = commutative.NewPath()

// 		case uint8(reflect.Kind(noncommutative.STRING)): // delta big int
// 			v = noncommutative.NewString("")

// 		case uint8(reflect.Kind(commutative.UINT256)): // delta big int
// 			v = commutative.NewUnboundedU256()

// 		case uint8(reflect.Kind(commutative.UINT64)):
// 			v = commutative.NewUnboundedUint64()

// 		case uint8(reflect.Kind(noncommutative.INT64)):
// 			v = new(noncommutative.Int64)

// 		case uint8(reflect.Kind(noncommutative.BYTES)):
// 			v = noncommutative.NewBytes([]byte{})
// 		}

// 		// fmt.Println(path)
// 		if !this.IfExists(path) {
// 			transitions = append(transitions, univalue.NewUnivalue(tx, path, 0, 1, 0, v, nil))

// 			if _, err := this.Write(tx, path, v); err != nil { // root path
// 				return nil, err
// 			}

// 			if !this.IfExists(path) {
// 				_, err := this.Write(tx, path, v)
// 				return transitions, err // root path
// 			}
// 		}
// 	}
// 	return transitions, nil
// }
