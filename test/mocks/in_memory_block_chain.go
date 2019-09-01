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
	"sync"
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
	memberId primitives.MemberId
	items    []*chainItem
	lock     sync.RWMutex
}

func NewInMemoryBlockchain() *InMemoryBlockchain {
	return &InMemoryBlockchain{
		items: []*chainItem{
			//{interfaces.GenesisBlock, nil},
		},
	}
}

func (bs *InMemoryBlockchain) WithMemberId(memberId primitives.MemberId) *InMemoryBlockchain {
	bs.memberId = memberId
	return bs
}
func (bs *InMemoryBlockchain) AppendBlockToChain(block interfaces.Block, blockProof []byte) {
	bs.lock.Lock()
	defer bs.lock.Unlock()
	bs.items = append(bs.items, &chainItem{block, blockProof})
}

func (bs *InMemoryBlockchain) LastBlock() interfaces.Block {
	bs.lock.RLock()
	defer bs.lock.RUnlock()

	if len(bs.items) == 0 {
		return nil
	}
	item := bs.items[len(bs.items)-1]
	return item.block
}

func (bs *InMemoryBlockchain) LastBlockProof() []byte {
	bs.lock.RLock()
	defer bs.lock.RUnlock()

	if len(bs.items) == 0 {
		return nil
	}
	item := bs.items[len(bs.items)-1]
	return item.blockProof
}

func (bs *InMemoryBlockchain) BlockAndProofAt(height primitives.BlockHeight) (interfaces.Block, []byte) {
	bs.lock.RLock()
	defer bs.lock.RUnlock()

	if int(height) >= len(bs.items) {
		panic(fmt.Sprintf("BlockAndProofAt() ID=%s requested H=%d but blockchain has only %d blocks", bs.memberId, height, len(bs.items)))
	}

	item := bs.items[height]
	return item.block, item.blockProof
}

func (bs *InMemoryBlockchain) Count() int {
	bs.lock.RLock()
	defer bs.lock.RUnlock()
	return len(bs.items)
}
