package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/matchers"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHappyFlow(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()
		net.StartConsensus(ctx)
		require.True(t, net.WaitForAllNodesToCommitTheSameBlock(ctx))
	})
}

func TestHappyFlowMessages(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()
		net.NodesPauseOnRequestNewBlock()

		net.StartConsensus(ctx)

		// let the leader run on the first round
		net.WaitForNodeToRequestNewBlock(ctx, net.Nodes[0])
		net.ResumeNodeRequestNewBlock(ctx, net.Nodes[0])

		net.WaitForAllNodesToCommitTheSameBlock(ctx)

		// hang the leader before the next round
		net.WaitForNodeToRequestNewBlock(ctx, net.Nodes[0])

		require.Equal(t, 1, net.Nodes[0].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[1].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[2].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[3].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))

		require.Equal(t, 0, net.Nodes[0].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPARE))
		require.Equal(t, 1, net.Nodes[1].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPARE))
		require.Equal(t, 1, net.Nodes[2].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPARE))
		require.Equal(t, 1, net.Nodes[3].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPARE))

		require.Equal(t, 1, net.Nodes[0].Communication.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
		require.Equal(t, 1, net.Nodes[1].Communication.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
		require.Equal(t, 1, net.Nodes[2].Communication.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
		require.Equal(t, 1, net.Nodes[3].Communication.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
	})
}

func TestConsensusFor8Blocks(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork().StartConsensus(ctx)
		for i := 0; i < 8; i++ {
			net.WaitForAllNodesToCommitTheSameBlock(ctx)
		}
	})
}

func TestHangingNode(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		net := network.ATestNetwork(4, block1, block2)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]
		net.NodesPauseOnValidate()
		net.StartConsensus(ctx)

		net.WaitForNodesToValidate(ctx, node1, node2, node3)
		net.ResumeNodesValidation(ctx, node1, node2)
		net.WaitForNodesToCommitABlock(ctx, node0, node1, node2)
		require.True(t, matchers.BlocksAreEqual(node0.GetLatestBlock(), block1))
		require.True(t, matchers.BlocksAreEqual(node1.GetLatestBlock(), block1))
		require.True(t, matchers.BlocksAreEqual(node2.GetLatestBlock(), block1))
		require.True(t, node3.GetLatestBlock() == interfaces.GenesisBlock)

		net.WaitForNodesToValidate(ctx, node1, node2)
		net.ResumeNodesValidation(ctx, node1, node2)
		net.WaitForNodesToCommitABlock(ctx, node0, node1, node2)
		require.True(t, matchers.BlocksAreEqual(node0.GetLatestBlock(), block2))
		require.True(t, matchers.BlocksAreEqual(node1.GetLatestBlock(), block2))
		require.True(t, matchers.BlocksAreEqual(node2.GetLatestBlock(), block2))
		require.True(t, node3.GetLatestBlock() == interfaces.GenesisBlock)

		net.ResumeNodesValidation(ctx, node3)
		net.WaitForNodesToCommitABlock(ctx, node3)
		require.True(t, matchers.BlocksAreEqual(node3.GetLatestBlock(), block1))

		net.WaitForNodesToValidate(ctx, node3)
		net.ResumeNodesValidation(ctx, node3)
		net.WaitForNodesToCommitABlock(ctx, node3)
		require.True(t, matchers.BlocksAreEqual(node3.GetLatestBlock(), block2))
	})
}
