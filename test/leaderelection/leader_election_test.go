package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test2fPlus1ViewChangeToBeElected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		h := NewHarness(ctx, t, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]
		node3 := h.net.Nodes[3]

		// hang the leader (node0)
		h.net.WaitForNodeToRequestNewBlock(node0)
		node0.Gossip.SetOutgoingWhitelist([]primitives.MemberId{})

		// manually cause new-view with 3 view-changes
		node0VCMessage := builders.AViewChangeMessage(node0.KeyManager, node0.MemberId, 1, 1, nil)
		node2VCMessage := builders.AViewChangeMessage(node2.KeyManager, node2.MemberId, 1, 1, nil)
		node3VCMessage := builders.AViewChangeMessage(node3.KeyManager, node3.MemberId, 1, 1, nil)
		node1.Gossip.OnRemoteMessage(ctx, node0VCMessage.ToConsensusRawMessage())
		node1.Gossip.OnRemoteMessage(ctx, node2VCMessage.ToConsensusRawMessage())
		node1.Gossip.OnRemoteMessage(ctx, node3VCMessage.ToConsensusRawMessage())

		// now that we caused node1 to be the new leader, he'll ask for a new block (block2)
		h.net.WaitForNodeToRequestNewBlock(node1)
		h.net.ResumeNodeRequestNewBlock(node1)

		// release the hanged the leader (node0)
		h.net.ResumeNodeRequestNewBlock(node0)

		// make sure that we're on block2
		h.net.WaitForAllNodesToCommitBlock(block2)
	})
}

func TestBlockIsNotUsedWhenElectionHappened(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)
		block3 := builders.CreateBlock(block1)

		h := NewHarness(ctx, t, block1, block2, block3)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]

		// processing block1, should be agreed by all nodes
		h.net.WaitForNodeToRequestNewBlock(node0)
		h.net.ResumeNodeRequestNewBlock(node0)
		h.net.WaitForAllNodesToCommitBlock(block1)

		// processing block 2
		h.net.WaitForNodeToRequestNewBlock(node0)
		node0.Gossip.SetOutgoingWhitelist([]primitives.MemberId{})
		// selection node 1 as the leader (dropping block2)
		h.net.Nodes[0].TriggerElection()
		h.net.Nodes[1].TriggerElection()
		h.net.Nodes[2].TriggerElection()
		h.net.Nodes[3].TriggerElection()

		node0.Gossip.ClearOutgoingWhitelist()

		h.net.WaitForNodeToRequestNewBlock(node1)
		h.net.ResumeNodeRequestNewBlock(node0)
		require.True(h.t, node1.IsLeader())

		// processing block 3
		h.net.ResumeNodeRequestNewBlock(node1)
		h.net.WaitForAllNodesToCommitBlock(block3)
	})
}

func TestThatNewLeaderSendsNewViewWhenElected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)
		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]
		node3 := h.net.Nodes[3]

		h.net.WaitForNodeToRequestNewBlock(node0)
		node0.Gossip.SetOutgoingWhitelist([]primitives.MemberId{})

		// selection node 1 as the leader
		node0.TriggerElection()
		node1.TriggerElection()
		node2.TriggerElection()
		node3.TriggerElection()

		h.net.ResumeNodeRequestNewBlock(node0)
		h.net.WaitForNodeToRequestNewBlock(node1)
		node0.Gossip.ClearOutgoingWhitelist()

		h.net.ResumeNodeRequestNewBlock(node1)
		h.net.WaitForAllNodesToCommitTheSameBlock()

		require.Equal(t, 1, node0.Gossip.CountSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE))
		require.Equal(t, 1, node2.Gossip.CountSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE))
		require.Equal(t, 1, node3.Gossip.CountSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE))
		require.Equal(t, 1, node1.Gossip.CountSentMessages(protocol.LEAN_HELIX_NEW_VIEW))
	})
}

func TestNotCountingViewChangeFromTheSameNode(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		h := NewHarness(ctx, t, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]

		// hang the leader (node0)
		h.net.WaitForNodeToRequestNewBlock(node0)

		// sending only 4 view-change from the same node
		node1.Gossip.OnRemoteMessage(ctx, builders.AViewChangeMessage(node2.KeyManager, node2.MemberId, 1, 1, nil).ToConsensusRawMessage())
		node1.Gossip.OnRemoteMessage(ctx, builders.AViewChangeMessage(node2.KeyManager, node2.MemberId, 1, 1, nil).ToConsensusRawMessage())
		node1.Gossip.OnRemoteMessage(ctx, builders.AViewChangeMessage(node2.KeyManager, node2.MemberId, 1, 1, nil).ToConsensusRawMessage())
		node1.Gossip.OnRemoteMessage(ctx, builders.AViewChangeMessage(node2.KeyManager, node2.MemberId, 1, 1, nil).ToConsensusRawMessage())

		node1.Gossip.CountSentMessages(protocol.LEAN_HELIX_NEW_VIEW)
	})
}

func TestNoNewViewIfLessThan2fPlus1ViewChange(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		h := NewHarness(ctx, t, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]

		// hang the leader (node0)
		h.net.WaitForNodeToRequestNewBlock(node0)

		// sending only 2 view-change (not enough to be elected)
		node0VCMessage := builders.AViewChangeMessage(node0.KeyManager, node0.MemberId, 1, 1, nil)
		node2VCMessage := builders.AViewChangeMessage(node2.KeyManager, node2.MemberId, 1, 1, nil)
		node1.Gossip.OnRemoteMessage(ctx, node0VCMessage.ToConsensusRawMessage())
		node1.Gossip.OnRemoteMessage(ctx, node2VCMessage.ToConsensusRawMessage())

		// release the hanged the leader (node0)
		h.net.ResumeNodeRequestNewBlock(node0)

		// make sure that we're on block2
		h.net.WaitForAllNodesToCommitBlock(block2)
	})
}

func TestLeaderCircularOrdering(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		// Nodes might get into prepared state, and send their block in the view-change
		// meaning that the new leader will not request new block and we can't hang him.
		// to prevent nodes from getting prepared, we just don't validate the block

		h.net.Nodes[0].BlockUtils.ValidationResult = false
		h.net.Nodes[1].BlockUtils.ValidationResult = false
		h.net.Nodes[2].BlockUtils.ValidationResult = false
		h.net.Nodes[3].BlockUtils.ValidationResult = false

		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[0])

		// Making sure that node 0 is the leader
		h.verifyNodeIsLeader(0)

		// selecting node 1 as the leader
		h.net.Nodes[0].TriggerElection()
		h.net.Nodes[2].TriggerElection()
		h.net.Nodes[3].TriggerElection()

		h.net.ResumeNodeRequestNewBlock(h.net.Nodes[0])
		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[1])

		h.verifyNodeIsLeader(1)

		// selecting node 2 as the leader
		h.net.Nodes[0].TriggerElection()
		h.net.Nodes[1].TriggerElection()
		h.net.Nodes[3].TriggerElection()

		h.net.ResumeNodeRequestNewBlock(h.net.Nodes[1])
		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[2])

		h.verifyNodeIsLeader(2)

		// selecting node 3 as the leader
		h.net.Nodes[0].TriggerElection()
		h.net.Nodes[1].TriggerElection()
		h.net.Nodes[2].TriggerElection()

		h.net.ResumeNodeRequestNewBlock(h.net.Nodes[2])
		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[3])

		h.verifyNodeIsLeader(3)

		// back to node 0 as the leader
		h.net.Nodes[1].TriggerElection()
		h.net.Nodes[2].TriggerElection()
		h.net.Nodes[3].TriggerElection()

		h.net.ResumeNodeRequestNewBlock(h.net.Nodes[3])
		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[0])

		h.verifyNodeIsLeader(0)
	})
}
