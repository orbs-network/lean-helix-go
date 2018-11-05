package test

import (
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCallValidateBlockDuringConsensusConsensus(t *testing.T) {
	t.Skip()

	testNetwork := builders.ABasicTestNetwork()
	testNetwork.StartConsensusOnAllNodes()

	node0 := testNetwork.Nodes[0]
	actual := node0.BlockUtils.CounterOfValidation()

	require.Equal(t, 1, actual)

}
