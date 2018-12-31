package mocks

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type BlocksPool struct {
	upcomingBlocks []interfaces.Block
	latestBlock    interfaces.Block
}

func (bp *BlocksPool) PopBlock() interfaces.Block {
	var nextBlock interfaces.Block
	if len(bp.upcomingBlocks) > 0 {
		// Simple queue impl, see https://github.com/golang/go/wiki/SliceTricks
		nextBlock, bp.upcomingBlocks = bp.upcomingBlocks[0], bp.upcomingBlocks[1:]
	} else {
		nextBlock = ABlock(bp.latestBlock)
	}
	bp.latestBlock = nextBlock
	return nextBlock
}

func NewBlocksPool(upcomingBlocks []interfaces.Block) *BlocksPool {
	return &BlocksPool{
		latestBlock:    interfaces.GenesisBlock,
		upcomingBlocks: upcomingBlocks,
	}
}
