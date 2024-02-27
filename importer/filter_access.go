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

package importer

import (
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/storage-committer/interfaces"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

type IPAccess struct {
	*univalue.Univalue
	Err error
}

func (this IPAccess) From(v *univalue.Univalue) *univalue.Univalue {
	if this.Err != nil || v.Persistent() {
		return nil
	}

	if v.Value() == nil {
		return v
	}

	value := v.Value().(interfaces.Type)
	return v.New(
		&v.Property,
		common.IfThen(value.IsCommutative() && value.IsNumeric(), value, nil), // commutative but not meta, for the accumulator
		[]byte{},
	)
}

type ITAccess struct{ IPAccess }

func (this ITAccess) From(v *univalue.Univalue) *univalue.Univalue {
	value := this.IPAccess.From(v)
	// converted := common.IfThenDo1st(value != nil, func() *univalue.Univalue { return value.(*univalue.Univalue) }, nil)
	if value == nil {
		return nil
	}

	if value.Value() == nil { // Entry deletion
		return value
	}

	typed := value.Value().(interfaces.Type)
	newv := typed.New(
		nil,
		typed.Delta(),
		typed.DeltaSign(),
		typed.Min(),
		typed.Max(),
	).(interfaces.Type)

	if typed.IsCommutative() && typed.IsNumeric() { // For the accumulator, commutative u64 & U256
		newv.SetValue(typed.Value())
	}

	return value.New(
		&value.Property,
		typed,
		[]byte{},
	)
}
