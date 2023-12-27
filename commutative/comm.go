/*
 *   Copyright (c) 2023 Arcology Network
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package commutative

import intf "github.com/arcology-network/concurrenturl/interfaces"

func ApplyDelta[T any](initv intf.Type, typedVals []intf.Type) (interface{}, int, error) {
	for i, v := range typedVals {
		// v := vec[i].Value()
		if initv == nil && v != nil { // New value
			initv = v
		}

		if initv == nil && v == nil { // Delete a non-existent
			initv = nil
		}

		if initv != nil && v != nil { // Update an existent
			if _, _, _, _, err := initv.Set(v.(T), nil); err != nil {
				return nil, i, err
			}
		}

		if initv != nil && v == nil { // Delete an existent
			initv = nil
		}
	}

	newValue, _, _ := initv.Get()
	return newValue, len(typedVals), nil
}
