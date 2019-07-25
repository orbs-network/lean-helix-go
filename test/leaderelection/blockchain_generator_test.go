package leaderelection

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBlockchainGenerator(t *testing.T) {

	block1 := mocks.ABlock(interfaces.GenesisBlock)
	block2 := mocks.ABlock(block1)
	block3 := mocks.ABlock(block2)

	bc := GenerateBlockChainFor([]interfaces.Block{block1, block2, block3})

	require.Equal(t, primitives.BlockHeight(3), bc.LastBlock().Height())
}
