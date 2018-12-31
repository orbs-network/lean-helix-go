package blockheight

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

func GetBlockHeight(prevBlock interfaces.Block) primitives.BlockHeight {
	if prevBlock == interfaces.GenesisBlock {
		return 0
	} else {
		return prevBlock.Height()
	}
}
