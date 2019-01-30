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
