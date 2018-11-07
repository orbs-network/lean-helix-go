package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCallValidateBlockDuringConsensusConsensus(t *testing.T) {
	testNetwork := builders.ABasicTestNetwork()
	testNetwork.StartConsensusOnAllNodes()

	require.True(t, testNetwork.AllNodesValidatedOnceBeforeCommit())
}

func TestHappyFlow(t *testing.T) {
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(block1)

	testNetwork := builders.ATestNetwork(4, []leanhelix.Block{block1, block2})
	testNetwork.StartConsensusOnAllNodes()

	require.True(t, testNetwork.AllNodesAgreeOnBlock(block1))
}
