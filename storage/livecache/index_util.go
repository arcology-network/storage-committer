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

package livecache

import (
	"github.com/emirpasic/gods/v2/trees/redblacktree"
	"golang.org/x/exp/constraints"
)

func between[K constraints.Ordered, V any](node *redblacktree.Node[K, V], low, high K, result *[]V) {
	if node == nil {
		return
	}

	// If the current node's key is within the range, add it to the result
	if node.Key >= low && node.Key <= high {
		*result = append(*result, node.Value)
	}

	// If the current node's key is greater than the lower bound, search the left subtree
	if node.Key >= low {
		between(node.Left, low, high, result)
	}

	// If the current node's key is less than the upper bound, search the right subtree
	if node.Key <= high {
		between(node.Right, low, high, result)
	}
}

func greaterOrEqualThan[K constraints.Ordered, V any](node *redblacktree.Node[K, V], threshold K, result *[]V) {
	if node == nil {
		return
	}

	// If the current node's key is greater than the threshold, add it to the result
	if node.Key >= threshold {
		*result = append(*result, node.Value)
		greaterOrEqualThan(node.Left, threshold, result)
	}

	// Always search the right subtree (it may contain values larger than the threshold)
	greaterOrEqualThan(node.Right, threshold, result)
}

func lessOrEqualThan[K constraints.Ordered, V any](node *redblacktree.Node[K, V], threshold K, result *[]V) {
	if node == nil {
		return
	}

	// If the current node's key is greater than the threshold, add it to the result
	if node.Key <= threshold {
		*result = append(*result, node.Value)
		lessOrEqualThan(node.Right, threshold, result)
	}

	// Always search the right subtree (it may contain values larger than the threshold)
	lessOrEqualThan(node.Left, threshold, result)
}
