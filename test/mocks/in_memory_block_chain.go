// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type chainItem struct {
	block      interfaces.Block
	blockProof []byte
}
type InMemoryBlockchain struct {
	items []*chainItem
}

func NewInMemoryBlockchain() *InMemoryBlockchain {
	return &InMemoryBlockchain{
		items: []*chainItem{
			{interfaces.GenesisBlock, nil},
		},
	}
}

func (bs *InMemoryBlockchain) GetFirstXItems(count int) *InMemoryBlockchain {
	newItems := make([]*chainItem, count, count)
	copied := copy(newItems, bs.items)
	if copied != count {
		panic(fmt.Sprintf("GetFirstXItems(): bad copy: bs.items=%v newItems=%v copied=%d count=%d", bs.items, newItems, copied, count))
	}
	return &InMemoryBlockchain{
		items: newItems,
	}
}

func (bs *InMemoryBlockchain) AppendBlockToChain(block interfaces.Block, blockProof []byte) {
	bs.items = append(bs.items, &chainItem{block, blockProof})
}

func (bs *InMemoryBlockchain) LastBlock() interfaces.Block {
	item := bs.items[len(bs.items)-1]
	return item.block
}

func (bs *InMemoryBlockchain) LastBlockProof() []byte {
	item := bs.items[len(bs.items)-1]
	return item.blockProof
}

func (bs *InMemoryBlockchain) BlockProofAt(height primitives.BlockHeight) []byte {
	item := bs.items[height]
	return item.blockProof
}

func (bs *InMemoryBlockchain) BlockAndProofAt(height primitives.BlockHeight) (interfaces.Block, []byte) {
	item := bs.items[height]
	return item.block, item.blockProof
}
