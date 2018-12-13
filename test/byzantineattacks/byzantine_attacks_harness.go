package byzantineattacks

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

type harness struct {
	t   *testing.T
	net *builders.TestNetwork
}

func NewHarness(ctx context.Context, t *testing.T, totalNodes int, blocksPool ...leanhelix.Block) *harness {
	net := builders.ATestNetwork(totalNodes, blocksPool...)
	net.NodesPauseOnRequestNewBlock()
	net.StartConsensus(ctx)

	return &harness{
		t:   t,
		net: net,
	}
}

func (h *harness) TriggerElection() {
	h.net.TriggerElection()
}

func (h *harness) verifyNodeIsLeader(nodeIdx int) {
	for idx, node := range h.net.Nodes {
		if idx == nodeIdx {
			require.True(h.t, node.IsLeader())
		} else {
			require.False(h.t, node.IsLeader())
		}
	}
}
