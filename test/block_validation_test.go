package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCallValidateBlockDuringConsensus(t *testing.T) {
	WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork(ctx)
		net.StartConsensus(ctx)
		net.WaitForNodesToValidate(net.Nodes[1], net.Nodes[2], net.Nodes[3])
		net.ResumeNodesValidation(net.Nodes[1], net.Nodes[2], net.Nodes[3])

		require.True(t, net.AllNodesValidatedNoMoreThanOnceBeforeCommit())
	})
}

// TODO: uncomment
//func TestNoConsensusWhenValidationFailed(t *testing.T) {
//	block1 := builders.CreateBlock(builders.GenesisBlock)
//	block2 := builders.CreateBlock(block1)
//
//	testNetwork := builders.ATestNetwork(4, []leanhelix.Block{block1, block2})
//	testNetwork.Nodes[0].BlockUtils.FailValidations()
//	testNetwork.Nodes[1].BlockUtils.FailValidations()
//	testNetwork.Nodes[2].BlockUtils.FailValidations()
//	testNetwork.Nodes[3].BlockUtils.FailValidations()
//	testNetwork.StartConsensus()
//
//	require.False(t, testNetwork.AllNodesAgreeOnBlock(block1))
//}
