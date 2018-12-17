package byzantineattacks

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/orbs-network-go/test"
	"testing"
	"time"
)

func TestThatWeReachConsensusWhere1OutOf4NodeIsByzantine(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := builders.CreateBlock(builders.GenesisBlock)
		net := builders.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]leanhelix.Block{block}).
			Build()

		net.Nodes[3].Gossip.SetIncomingWhitelist([]primitives.MemberId{})

		net.StartConsensus(ctx)

		net.WaitForNodesToCommitABlock(net.Nodes[0], net.Nodes[1], net.Nodes[2])
	})
}

func TestThatWeReachConsensusWhere2OutOf7NodesAreByzantine(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := builders.CreateBlock(builders.GenesisBlock)
		net := builders.
			NewTestNetworkBuilder().
			WithNodeCount(7).
			WithBlocks([]leanhelix.Block{block}).
			Build()

		net.Nodes[1].Gossip.SetIncomingWhitelist([]primitives.MemberId{})
		net.Nodes[2].Gossip.SetIncomingWhitelist([]primitives.MemberId{})

		net.StartConsensus(ctx)

		net.WaitForNodesToCommitABlock(net.Nodes[0], net.Nodes[3], net.Nodes[4], net.Nodes[5], net.Nodes[6])
	})
}

func TestThatAByzantineLeaderCanNotCauseAForkBySendingTwoBlocks(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		net := builders.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]leanhelix.Block{block1}).
			Build()

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		//node3 := net.Nodes[3]

		node0.Gossip.SetOutgoingWhitelist([]primitives.MemberId{node1.MemberId, node2.MemberId})

		// the leader (node0) is suggesting block1 to node1 and node2 (not to node3)
		net.StartConsensus(ctx)

		// node0, node1 and node2 should reach consensus
		net.WaitForNodesToCommitASpecificBlock(block1, node0, node1, node2)

		node0.StartConsensus(ctx)
	})
}

func TestNoForkWhenAByzantineNodeSendsABadBlockSeveralTimes(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		goodBlock := builders.CreateBlock(builders.GenesisBlock)
		fakeBlock := builders.CreateBlock(builders.GenesisBlock)
		net := builders.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]leanhelix.Block{goodBlock}).
			Build()

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		byzantineNode := net.Nodes[3]
		net.NodesPauseOnRequestNewBlock(node0)
		net.StartConsensus(ctx)

		net.WaitForNodeToRequestNewBlock(node0)

		// fake a preprepare message from node3 (byzantineNode) that points to a unrelated block (Should be ignored)
		ppm := builders.APreprepareMessage(byzantineNode.KeyManager, byzantineNode.MemberId, 1, 1, fakeBlock)
		byzantineNode.Gossip.SendConsensusMessage(ctx, []primitives.MemberId{node0.MemberId, node1.MemberId, node2.MemberId}, ppm.ToConsensusRawMessage())
		byzantineNode.Gossip.SendConsensusMessage(ctx, []primitives.MemberId{node0.MemberId, node1.MemberId, node2.MemberId}, ppm.ToConsensusRawMessage())
		byzantineNode.Gossip.SendConsensusMessage(ctx, []primitives.MemberId{node0.MemberId, node1.MemberId, node2.MemberId}, ppm.ToConsensusRawMessage())
		byzantineNode.Gossip.SendConsensusMessage(ctx, []primitives.MemberId{node0.MemberId, node1.MemberId, node2.MemberId}, ppm.ToConsensusRawMessage())

		net.ResumeNodeRequestNewBlock(node0)

		net.WaitForAllNodesToCommitBlock(goodBlock)
	})
}

func TestThatAByzantineLeaderCanNotCauseAFork(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(builders.GenesisBlock)

		net := builders.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]leanhelix.Block{block1, block2}).
			Build()

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]
		node0.Gossip.SetOutgoingWhitelist([]primitives.MemberId{node1.MemberId, node2.MemberId})
		node1.Gossip.SetOutgoingWhitelist([]primitives.MemberId{node2.MemberId})
		node2.Gossip.SetOutgoingWhitelist([]primitives.MemberId{})
		node3.Gossip.SetOutgoingWhitelist([]primitives.MemberId{})

		net.StartConsensus(ctx)

		// Because we only allow node0 (The leader) to talk to node1 and node2
		// and node1 only to talk to node2,
		// we can expect (only) node2 to be prepared on block1
		test.Eventually(time.Duration(100)*time.Millisecond, func() bool {
			_, ppOk := node2.Storage.GetPreprepareMessage(1, 1)
			p, _ := node2.Storage.GetPrepareMessages(1, 1, builders.CalculateBlockHash(block1))
			return ppOk && len(p) == 2
		})

		// now that node2 is prepared on block1, we'll close any communication
		// to it, and open all the other nodes communication.
		// then, we trigger an election. Node2's prepared block will not get sent in a view-change
		node0.Gossip.ClearOutgoingWhitelist()
		node1.Gossip.ClearOutgoingWhitelist()
		node2.Gossip.ClearOutgoingWhitelist()
		node3.Gossip.ClearOutgoingWhitelist()

		node2.Gossip.SetOutgoingWhitelist([]primitives.MemberId{})
		node2.Gossip.SetIncomingWhitelist([]primitives.MemberId{})

		// selection node 1 as the leader
		node0.TriggerElection()
		node1.TriggerElection()
		node2.TriggerElection()
		node3.TriggerElection()

		net.WaitForNodesToCommitASpecificBlock(block2, node0, node1, node3)
	})
}
