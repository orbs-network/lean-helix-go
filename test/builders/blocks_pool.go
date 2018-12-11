package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
)

type BlocksPool struct {
	upcomingBlocks []lh.Block
	latestBlock    lh.Block
}

func (bp *BlocksPool) PopBlock() lh.Block {
	var nextBlock lh.Block
	if len(bp.upcomingBlocks) > 0 {
		// Simple queue impl, see https://github.com/golang/go/wiki/SliceTricks
		nextBlock, bp.upcomingBlocks = bp.upcomingBlocks[0], bp.upcomingBlocks[1:]
	} else {
		nextBlock = CreateBlock(bp.latestBlock)
	}
	bp.latestBlock = nextBlock
	return nextBlock
}

func NewBlocksPool(upcomingBlocks []lh.Block) *BlocksPool {
	return &BlocksPool{
		latestBlock:    GenesisBlock,
		upcomingBlocks: upcomingBlocks,
	}
}
