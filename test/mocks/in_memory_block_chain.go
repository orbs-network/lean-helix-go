// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/pkg/errors"
)

type chainItem struct {
	block      interfaces.Block
	blockProof []byte
}

func (i *chainItem) Block() interfaces.Block {
	return i.block
}

func (i *chainItem) Proof() []byte {
	return i.blockProof
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

func (bs *InMemoryBlockchain) GetFirstXItems(count int) (*InMemoryBlockchain, error) {
	if bs == nil {
		return nil, errors.New("GetFirstXItems(): bs in nil")
	}
	newItems := make([]*chainItem, count, count)
	copied := copy(newItems, bs.items)
	if copied != count {
		return nil, errors.Errorf("GetFirstXItems(): bad copy: bs.items=%v newItems=%v copied=%d count=%d", bs.items, newItems, copied, count)
	}
	return &InMemoryBlockchain{
		items: newItems,
	}, nil
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

// Do we want to strictly maintain escapulation here? probably only if we decide to RWLock/RLock the "bs.items" field
func (bs *InMemoryBlockchain) Items() []*chainItem {
	return bs.items
}
