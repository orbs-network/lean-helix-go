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
	leaderBlockUtils := testNetwork.Nodes[0].BlockUtils

	leaderBlockUtils.ProvideNextBlock()

	node1 := testNetwork.Nodes[1]
	require.Equal(t, uint(1), node1.BlockUtils.CounterOfValidation())

	node2 := testNetwork.Nodes[2]
	require.Equal(t, uint(1), node2.BlockUtils.CounterOfValidation())

	node3 := testNetwork.Nodes[3]
	require.Equal(t, uint(1), node3.BlockUtils.CounterOfValidation())
}

func TestHappyFlowWithBlockValidation(t *testing.T) {
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(block1)
	testNetwork := builders.ATestNetwork(4, []leanhelix.Block{block1, block2})

	testNetwork.StartConsensusOnAllNodes()
	leaderBlockUtils := testNetwork.Nodes[0].BlockUtils
	leaderBlockUtils.ProvideNextBlock()

	require.True(t, testNetwork.AllNodesAgreeOnBlock(block1))
}
