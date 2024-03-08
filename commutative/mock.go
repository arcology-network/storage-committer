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

package commutative

// A mockWriteCache is a mock implementation of the WriteCache for testing purposes.
type mockWriteCache struct {
	dict map[string]interface{}
}

func NewMockWriteCache() *mockWriteCache {
	return &mockWriteCache{
		dict: map[string]interface{}{},
	}
}

func (this *mockWriteCache) Write(tx uint32, key string, value interface{}) (int64, error) {
	this.dict[key] = value
	return 0, nil
}

func (this *mockWriteCache) InCache(key string) (interface{}, bool) {
	v, ok := this.dict[key]
	return v, ok
}
