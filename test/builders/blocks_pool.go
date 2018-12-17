package builders

import (
	"github.com/orbs-network/lean-helix-go"
)

type BlocksPool struct {
	upcomingBlocks []leanhelix.Block
	latestBlock    leanhelix.Block
}

func (bp *BlocksPool) PopBlock() leanhelix.Block {
	var nextBlock leanhelix.Block
	if len(bp.upcomingBlocks) > 0 {
		// Simple queue impl, see https://github.com/golang/go/wiki/SliceTricks
		nextBlock, bp.upcomingBlocks = bp.upcomingBlocks[0], bp.upcomingBlocks[1:]
	} else {
		nextBlock = CreateBlock(bp.latestBlock)
	}
	bp.latestBlock = nextBlock
	return nextBlock
}

func NewBlocksPool(upcomingBlocks []leanhelix.Block) *BlocksPool {
	return &BlocksPool{
		latestBlock:    GenesisBlock,
		upcomingBlocks: upcomingBlocks,
	}
}
