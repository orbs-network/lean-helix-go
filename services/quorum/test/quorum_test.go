package test

import (
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenesisBlockHeight(t *testing.T) {
	actual := blockheight.GetBlockHeight(interfaces.GenesisBlock)
	require.Equal(t, primitives.BlockHeight(0), actual)
}

func TestBasicBlockHeight(t *testing.T) {
	block1 := mocks.ABlock(interfaces.GenesisBlock)
	block2 := mocks.ABlock(block1)
	block3 := mocks.ABlock(block2)
	actual := blockheight.GetBlockHeight(block3)
	require.Equal(t, primitives.BlockHeight(3), actual)
}