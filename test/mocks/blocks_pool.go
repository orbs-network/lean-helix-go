// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"sync"
)

type BlocksPool struct {
	upcomingBlocks []interfaces.Block
	mutex          *sync.Mutex
}

func (bp *BlocksPool) PopBlock(prevBlock interfaces.Block) interfaces.Block {
	bp.mutex.Lock()
	var nextBlock interfaces.Block
	if len(bp.upcomingBlocks) > 0 {
		// Simple queue impl, see https://github.com/golang/go/wiki/SliceTricks
		nextBlock, bp.upcomingBlocks = bp.upcomingBlocks[0], bp.upcomingBlocks[1:]
	} else {
		nextBlock = ABlock(prevBlock)
	}
	bp.mutex.Unlock()
	return nextBlock
}

func NewBlocksPool(upcomingBlocks []interfaces.Block) *BlocksPool {
	return &BlocksPool{
		upcomingBlocks: upcomingBlocks,
		mutex:          &sync.Mutex{},
	}
}
