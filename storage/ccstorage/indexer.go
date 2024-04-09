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

package ccstorage

import (
	"runtime"

	codec "github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/univalue"
)

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
type CCIndexer struct {
	buffer  []*univalue.Univalue
	ccstore *DataStore

	partitionIDs  []uint64
	keyBuffer     []string
	valueBuffer   []interface{}
	encodedBuffer [][]byte //The encoded buffer contains the encoded values
}

func NewCCIndexer(ccstore *DataStore) *CCIndexer {
	return &CCIndexer{
		buffer:  []*univalue.Univalue{},
		ccstore: ccstore,

		partitionIDs:  []uint64{},
		keyBuffer:     []string{},
		valueBuffer:   []interface{}{},
		encodedBuffer: [][]byte{},
	}
}

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
func (this *CCIndexer) Add(transitions []*univalue.Univalue) {
	for _, v := range transitions {
		if v.GetPath() != nil || !platform.IsEthPath(*v.GetPath()) {
			this.buffer = append(this.buffer, v)
		}
	}
}

func (this *CCIndexer) Finalize(_ intf.CommittableStore) {
	slice.RemoveIf(&this.buffer, func(_ int, v *univalue.Univalue) bool {
		return v.GetPath() == nil
	}) // Remove the transitions that are marked

	// Extract the keys and values from the buffer
	this.keyBuffer = make([]string, len(this.buffer))
	this.valueBuffer = slice.ParallelTransform(this.buffer, runtime.NumCPU(), func(i int, v *univalue.Univalue) interface{} {
		this.keyBuffer[i] = *v.GetPath()
		return v.Value()
	})

	// Compress the keys if the keyCompressor is available
	this.keyBuffer = common.IfThenDo1st(
		this.ccstore.keyCompressor != nil,
		func() []string {
			return this.ccstore.keyCompressor.CompressOnTemp(codec.Strings(this.keyBuffer).Clone())
		},
		this.keyBuffer,
	)

	// Encode the keys and values to the buffer so that they can be written to calcualte the root hash.
	this.encodedBuffer = make([][]byte, len(this.valueBuffer))
	for i := 0; i < len(this.valueBuffer); i++ {
		if this.valueBuffer[i] != nil {
			this.encodedBuffer[i] = this.ccstore.encoder(this.keyBuffer[i], this.valueBuffer[i])
		}
	}
}

func (this *CCIndexer) Clear() {
	this.partitionIDs = this.partitionIDs[:0]
	this.keyBuffer = this.keyBuffer[:0]
	this.valueBuffer = this.valueBuffer[:0]
	this.encodedBuffer = this.encodedBuffer[:0]
}
