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

func TestNoConsensusWhenValidationFailed(t *testing.T) {
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(block1)

	testNetwork := builders.ATestNetwork(4, []leanhelix.Block{block1, block2})
	testNetwork.Nodes[0].BlockUtils.FailValidations()
	testNetwork.Nodes[1].BlockUtils.FailValidations()
	testNetwork.Nodes[2].BlockUtils.FailValidations()
	testNetwork.Nodes[3].BlockUtils.FailValidations()
	testNetwork.StartConsensusOnAllNodes()

	require.False(t, testNetwork.AllNodesAgreeOnBlock(block1))
}
