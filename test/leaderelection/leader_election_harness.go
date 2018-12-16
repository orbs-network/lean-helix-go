package leaderelection

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

type harness struct {
	t   *testing.T
	net *builders.TestNetwork
}

func NewHarness(ctx context.Context, t *testing.T, blocksPool ...leanhelix.Block) *harness {
	//net := builders.NewTestNetworkBuilder().WithNodeCount(4).WithBlocks(blocksPool).LogToConsole().Build()
	net := builders.ATestNetwork(4, blocksPool...)
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
			require.True(h.t, node.IsLeader(), fmt.Sprintf("node %d should be IsLeader=true", nodeIdx))
		} else {
			require.False(h.t, node.IsLeader(), fmt.Sprintf("node %d should be IsLeader=false", idx))
		}
	}
}
