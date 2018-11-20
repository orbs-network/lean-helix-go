package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHappyFlow(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork(ctx)
		net.StartConsensus(ctx)
		net.WaitForNodesToValidate(net.Nodes[1], net.Nodes[2], net.Nodes[3])
		net.ResumeNodesValidation(net.Nodes[1], net.Nodes[2], net.Nodes[3])
		require.True(t, net.InConsensus())
	})
}

func TestOnlyLeaderIsSendingPrePrepareOnce(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork(ctx)

		net.NodesWaitOnValidate(net.Nodes[1], net.Nodes[2], net.Nodes[3])
		net.StartConsensus(ctx)
		net.WaitForNodesToValidate(net.Nodes[1], net.Nodes[2], net.Nodes[3])

		require.Equal(t, 1, net.Nodes[0].Gossip.CountSentMessages(leanhelix.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[1].Gossip.CountSentMessages(leanhelix.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[2].Gossip.CountSentMessages(leanhelix.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[3].Gossip.CountSentMessages(leanhelix.LEAN_HELIX_PREPREPARE))
	})
}

func TestConsensusFor8Blocks(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork(ctx).StartConsensus(ctx)
		for i := 0; i < 8; i++ {
			net.WaitForNodesToValidate(net.Nodes[1], net.Nodes[2], net.Nodes[3])
			net.ResumeNodesValidation(net.Nodes[1], net.Nodes[2], net.Nodes[3])
			net.InConsensus()
		}
	})
}

func TestHangingNode(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		net := builders.ATestNetwork(ctx, 4, block1, block2)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]
		net.StartConsensus(ctx)

		net.WaitForNodesToValidate(node1, node2, node3)
		net.ResumeNodesValidation(node2, node3)
		net.WaitForNodesToCommitABlock(node0, node2, node3)
		require.True(t, builders.BlocksAreEqual(node0.GetLatestCommittedBlock(), block1))
		require.True(t, builders.BlocksAreEqual(node1.GetLatestCommittedBlock(), builders.GenesisBlock))
		require.True(t, builders.BlocksAreEqual(node2.GetLatestCommittedBlock(), block1))
		require.True(t, builders.BlocksAreEqual(node3.GetLatestCommittedBlock(), block1))

		net.WaitForNodesToValidate(node2, node3)
		net.ResumeNodesValidation(node2, node3)
		net.WaitForNodesToCommitABlock(node0, node2, node3)
		require.True(t, builders.BlocksAreEqual(node0.GetLatestCommittedBlock(), block2))
		require.True(t, builders.BlocksAreEqual(node1.GetLatestCommittedBlock(), builders.GenesisBlock))
		require.True(t, builders.BlocksAreEqual(node2.GetLatestCommittedBlock(), block2))
		require.True(t, builders.BlocksAreEqual(node3.GetLatestCommittedBlock(), block2))

		net.ResumeNodesValidation(node1)
		net.WaitForNodesToCommitABlock(node1)
		require.True(t, builders.BlocksAreEqual(node1.GetLatestCommittedBlock(), block1))

		net.WaitForNodesToValidate(node1)
		net.ResumeNodesValidation(node1)
		net.WaitForNodesToCommitABlock(node1)
		require.True(t, builders.BlocksAreEqual(node1.GetLatestCommittedBlock(), block2))
	})
}
