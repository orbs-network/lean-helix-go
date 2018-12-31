package matchers

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

func BlocksAreEqual(block1 interfaces.Block, block2 interfaces.Block) bool {
	return mocks.CalculateBlockHash(block1).Equal(mocks.CalculateBlockHash(block2))
}
