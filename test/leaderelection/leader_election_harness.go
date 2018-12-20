package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
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
