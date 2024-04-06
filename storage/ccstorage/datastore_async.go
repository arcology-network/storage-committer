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

func (this *DataStore) Start() {
	go func() {
		for {
			idxer := <-this.queue
			this.Precommit(idxer)
		}
	}()

	go func() {
		for {
			idxer := <-this.commitQueue
			this.CommitV2(idxer)
		}
	}()
}

func (this *DataStore) AsyncPrecommit(args ...interface{}) {
	idxer := args[0].(*CCIndexer)
	// this.queue <- idxer
	this.Precommit(idxer)
}

func (this *DataStore) AsyncCommit(args ...interface{}) {
	idxer := args[0].(*CCIndexer)
	this.queue <- idxer
	this.CommitV2(idxer)
}
