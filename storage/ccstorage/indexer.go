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
	"github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/univalue"
)

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
type CCIndexer struct {
	blockNum uint64
	trans    []*univalue.Univalue

	keyBuffer     []string
	valueBuffer   []interface{}
	encodedBuffer [][]byte //The encoded buffer contains the encoded values
}

func NewIndexer(store interfaces.Datastore) *CCIndexer {
	return &CCIndexer{
		blockNum:      0,
		trans:         []*univalue.Univalue{},
		keyBuffer:     []string{},
		valueBuffer:   []interface{}{},
		encodedBuffer: [][]byte{}, //The encoded buffer contains the encoded values
	}
}

func (this *CCIndexer) SetID(blockNum uint64) { this.blockNum = blockNum }

// An index by account address, transitions have the same Eth account address will be put together in a list
// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
func (this *CCIndexer) Add(transitions []*univalue.Univalue) {
	for _, v := range transitions {
		if v.GetPath() != nil || !platform.IsEthPath(*v.GetPath()) {
			this.trans = append(this.trans, v)
		}
	}
}

func (this *CCIndexer) Get() interface{} {
	this.keyBuffer = make([]string, len(this.trans))
	this.valueBuffer = slice.ParallelTransform(this.trans, runtime.NumCPU(), func(i int, v *univalue.Univalue) interface{} {
		this.keyBuffer[i] = *v.GetPath()
		if v.Value() != nil {
			return v.Value().(intf.Type)
		}
		return nil // A deletion
	})
	// return []interface{}{keys, tVals}
	return this
}

func (this *CCIndexer) Finalize(committable intf.CommittableStore) {
	slice.RemoveIf(&this.trans, func(_ int, v *univalue.Univalue) bool { return v.GetPath() == nil }) // Remove the transitions that are marked

	store := committable.(*DataStore)
	this.keyBuffer = common.IfThenDo1st( // Compress the keys if the keyCompressor is available
		store.keyCompressor != nil,
		func() []string { return store.keyCompressor.CompressOnTemp(codec.Strings(this.keyBuffer).Clone()) },
		this.keyBuffer,
	)

	// Encode the keys and values to the buffer so that they can be written to calcualte the root hash.
	encodedBuffer := make([][]byte, len(this.valueBuffer))
	for i := 0; i < len(this.valueBuffer); i++ {
		if this.valueBuffer[i] != nil {
			encodedBuffer[i] = store.encoder(this.keyBuffer[i], this.valueBuffer[i])
		}
	}
}

func (this *CCIndexer) Clear() {
	this.trans = this.trans[:0]
	this.keyBuffer = this.keyBuffer[:0]
	this.valueBuffer = this.valueBuffer[:0]
	this.encodedBuffer = this.encodedBuffer[:0]
}
