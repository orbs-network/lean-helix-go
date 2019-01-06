package blockheight

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

func GetBlockHeight(block interfaces.Block) primitives.BlockHeight {
	if block == interfaces.GenesisBlock {
		return 0
	} else {
		return block.Height()
	}
}
