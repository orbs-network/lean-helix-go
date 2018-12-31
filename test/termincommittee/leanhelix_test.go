package termincommittee

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
		require.True(t, net.WaitForAllNodesToCommitTheSameBlock())
	})
}

func TestOnlyLeaderIsSendingPrePrepareOnce(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()

		net.StartConsensusSync(ctx)
		net.Nodes[0].Tick(ctx)
		net.Nodes[1].Tick(ctx)
		net.Nodes[2].Tick(ctx)
		net.Nodes[3].Tick(ctx)

		require.Equal(t, 1, net.Nodes[0].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[1].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[2].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[3].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
	})
}

func TestHappyFlowMessages(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()
		net.NodesPauseOnRequestNewBlock()

		net.StartConsensus(ctx)

		// let the leader run on the first round
		net.WaitForNodeToRequestNewBlock(net.Nodes[0])
		net.ResumeNodeRequestNewBlock(net.Nodes[0])

		net.WaitForAllNodesToCommitTheSameBlock()

		// hang the leader before the next round
		net.WaitForNodeToRequestNewBlock(net.Nodes[0])

		require.Equal(t, 1, net.Nodes[0].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[1].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[2].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[3].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))

		require.Equal(t, 0, net.Nodes[0].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPARE))
		require.Equal(t, 1, net.Nodes[1].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPARE))
		require.Equal(t, 1, net.Nodes[2].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPARE))
		require.Equal(t, 1, net.Nodes[3].Gossip.CountSentMessages(protocol.LEAN_HELIX_PREPARE))

		require.Equal(t, 1, net.Nodes[0].Gossip.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
		require.Equal(t, 1, net.Nodes[1].Gossip.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
		require.Equal(t, 1, net.Nodes[2].Gossip.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
		require.Equal(t, 1, net.Nodes[3].Gossip.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
	})
}

func TestConsensusFor8Blocks(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork().StartConsensus(ctx)
		for i := 0; i < 8; i++ {
			net.WaitForAllNodesToCommitTheSameBlock()
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

		net.WaitForNodesToValidate(node1, node2, node3)
		net.ResumeNodesValidation(node1, node2)
		net.WaitForNodesToCommitABlock(node0, node1, node2)
		require.True(t, matchers.BlocksAreEqual(node0.GetLatestBlock(), block1))
		require.True(t, matchers.BlocksAreEqual(node1.GetLatestBlock(), block1))
		require.True(t, matchers.BlocksAreEqual(node2.GetLatestBlock(), block1))
		require.True(t, node3.GetLatestBlock() == interfaces.GenesisBlock)

		net.WaitForNodesToValidate(node1, node2)
		net.ResumeNodesValidation(node1, node2)
		net.WaitForNodesToCommitABlock(node0, node1, node2)
		require.True(t, matchers.BlocksAreEqual(node0.GetLatestBlock(), block2))
		require.True(t, matchers.BlocksAreEqual(node1.GetLatestBlock(), block2))
		require.True(t, matchers.BlocksAreEqual(node2.GetLatestBlock(), block2))
		require.True(t, node3.GetLatestBlock() == interfaces.GenesisBlock)

		net.ResumeNodesValidation(node3)
		net.WaitForNodesToCommitABlock(node3)
		require.True(t, matchers.BlocksAreEqual(node3.GetLatestBlock(), block1))

		net.WaitForNodesToValidate(node3)
		net.ResumeNodesValidation(node3)
		net.WaitForNodesToCommitABlock(node3)
		require.True(t, matchers.BlocksAreEqual(node3.GetLatestBlock(), block2))
	})
}
