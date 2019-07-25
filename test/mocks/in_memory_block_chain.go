// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type chainItem struct {
	block      interfaces.Block
	blockProof []byte
}
type InMemoryBlockChain struct {
	blockChain []*chainItem
}

func NewInMemoryBlockChain() *InMemoryBlockChain {
	return &InMemoryBlockChain{
		blockChain: []*chainItem{
			{interfaces.GenesisBlock, nil},
		},
	}
}

func (bs *InMemoryBlockChain) AppendBlockToChain(block interfaces.Block, blockProof []byte) {
	bs.blockChain = append(bs.blockChain, &chainItem{block, blockProof})
}

func (bs *InMemoryBlockChain) LastBlock() interfaces.Block {
	item := bs.blockChain[len(bs.blockChain)-1]
	return item.block
}

func (bs *InMemoryBlockChain) LastBlockProof() []byte {
	item := bs.blockChain[len(bs.blockChain)-1]
	return item.blockProof
}

func (bs *InMemoryBlockChain) BlockProofAt(height primitives.BlockHeight) []byte {
	item := bs.blockChain[height]
	return item.blockProof
}

func (bs *InMemoryBlockChain) BlockAndProofAt(height primitives.BlockHeight) (interfaces.Block, []byte) {
	item := bs.blockChain[height]
	return item.block, item.blockProof
}
