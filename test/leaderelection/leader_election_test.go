package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/orbs-network-go/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestThatNewLeaderSendsNewViewWhenElected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)
		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[0])

		// selection node 1 as the leader
		h.net.Nodes[0].TriggerElection()
		h.net.Nodes[2].TriggerElection()
		h.net.Nodes[3].TriggerElection()

		h.net.ResumeNodeRequestNewBlock(h.net.Nodes[0])
		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[1])

		require.Equal(t, 1, h.net.Nodes[0].Gossip.CountSentMessages(leanhelix.LEAN_HELIX_VIEW_CHANGE))
	})
}

func TestLeaderCircularOrdering(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)
		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[0])

		// Making sure that node 1 is the leader
		h.verifyNodeIsLeader(0)

		// selection node 1 as the leader
		h.net.Nodes[0].TriggerElection()
		h.net.Nodes[2].TriggerElection()
		h.net.Nodes[3].TriggerElection()

		h.net.ResumeNodeRequestNewBlock(h.net.Nodes[0])
		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[1])

		h.verifyNodeIsLeader(1)

		// selection node 2 as the leader
		h.net.Nodes[0].TriggerElection()
		h.net.Nodes[1].TriggerElection()
		h.net.Nodes[3].TriggerElection()

		h.net.ResumeNodeRequestNewBlock(h.net.Nodes[1])
		h.net.WaitForNodeToRequestNewBlock(h.net.Nodes[2])

		h.verifyNodeIsLeader(2)

		// selection node 3 as the leader
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
