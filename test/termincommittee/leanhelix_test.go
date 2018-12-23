package termincommittee

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHappyFlow(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork()
		net.StartConsensus(ctx)
		require.True(t, net.WaitForAllNodesToCommitTheSameBlock())
	})
}

func TestOnlyLeaderIsSendingPrePrepareOnce(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork()

		net.StartConsensusSync()
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
		net := builders.ABasicTestNetwork()
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
		net := builders.ABasicTestNetwork().StartConsensus(ctx)
		for i := 0; i < 8; i++ {
			net.WaitForAllNodesToCommitTheSameBlock()
		}
	})
}

func TestHangingNode(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(leanhelix.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		net := builders.ATestNetwork(4, block1, block2)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]
		net.NodesPauseOnValidate()
		net.StartConsensus(ctx)

		net.WaitForNodesToValidate(node1, node2, node3)
		net.ResumeNodesValidation(node1, node2)
		net.WaitForNodesToCommitABlock(node0, node1, node2)
		require.True(t, builders.BlocksAreEqual(node0.GetLatestBlock(), block1))
		require.True(t, builders.BlocksAreEqual(node1.GetLatestBlock(), block1))
		require.True(t, builders.BlocksAreEqual(node2.GetLatestBlock(), block1))
		require.True(t, node3.GetLatestBlock() == leanhelix.GenesisBlock)

		net.WaitForNodesToValidate(node1, node2)
		net.ResumeNodesValidation(node1, node2)
		net.WaitForNodesToCommitABlock(node0, node1, node2)
		require.True(t, builders.BlocksAreEqual(node0.GetLatestBlock(), block2))
		require.True(t, builders.BlocksAreEqual(node1.GetLatestBlock(), block2))
		require.True(t, builders.BlocksAreEqual(node2.GetLatestBlock(), block2))
		require.True(t, node3.GetLatestBlock() == leanhelix.GenesisBlock)

		net.ResumeNodesValidation(node3)
		net.WaitForNodesToCommitABlock(node3)
		require.True(t, builders.BlocksAreEqual(node3.GetLatestBlock(), block1))

		net.WaitForNodesToValidate(node3)
		net.ResumeNodesValidation(node3)
		net.WaitForNodesToCommitABlock(node3)
		require.True(t, builders.BlocksAreEqual(node3.GetLatestBlock(), block2))
	})
}

func TestNodeSync(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(leanhelix.GenesisBlock)
		block2 := builders.CreateBlock(block1)
		block3 := builders.CreateBlock(block2)

		net := builders.ATestNetwork(4, block1, block2, block3)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]

		net.NodesPauseOnRequestNewBlock()
		net.StartConsensus(ctx)

		// closing node3's network to messages (To make it out of sync)
		node3.Gossip.SetIncomingWhitelist([]primitives.MemberId{})

		// node0, node1, and node2 are progressing to block2
		net.WaitForNodeToRequestNewBlock(node0)
		net.ResumeNodeRequestNewBlock(node0)
		net.WaitForNodesToCommitASpecificBlock(block1, node0, node1, node2)

		net.WaitForNodeToRequestNewBlock(node0)
		net.ResumeNodeRequestNewBlock(node0)
		net.WaitForNodesToCommitASpecificBlock(block2, node0, node1, node2)

		// node3 is still "stuck" on the genesis block
		require.True(t, node3.GetLatestBlock() == leanhelix.GenesisBlock)

		net.WaitForNodeToRequestNewBlock(node0)

		// opening node3's network to messages
		node3.Gossip.ClearIncomingWhitelist()

		// syncing node3
		latestBlock := node0.GetLatestBlock()
		//lastestBlockProof := node0.GetLatestBlockProof()
		node3.Sync(latestBlock, []byte{1, 2, 3}) // TODO: create a real block proof

		net.ResumeNodeRequestNewBlock(node0)

		// now that node3 is synced, they all should progress to block3
		net.WaitForNodesToCommitASpecificBlock(block3, node0, node1, node2, node3)
	})
}

func TestThatWeDoNotAcceptNilBlockProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork()

		net.StartConsensus(ctx)

		block1 := builders.CreateBlock(leanhelix.GenesisBlock)
		block2 := builders.CreateBlock(block1)
		block3 := builders.CreateBlock(block2)
		require.False(t, net.Nodes[0].ValidateBlockConsensus(block3, nil))
	})
}
