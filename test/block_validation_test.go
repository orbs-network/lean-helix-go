package test

import (
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCallValidateBlockDuringConsensusConsensus(t *testing.T) {
	testNetwork := builders.ABasicTestNetwork()
	testNetwork.StartConsensusOnAllNodes()

	node1 := testNetwork.Nodes[1]
	actual := node1.BlockUtils.CounterOfValidation()

	require.Equal(t, uint(1), actual)
}
