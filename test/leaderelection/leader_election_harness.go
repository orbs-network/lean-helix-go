package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test/network"
	"testing"
)

type harness struct {
	t   *testing.T
	net *network.TestNetwork
}

func NewHarness(ctx context.Context, t *testing.T, blocksPool ...interfaces.Block) *harness {
	//net := builders.NewTestNetworkBuilder().WithNodeCount(4).WithBlocks(blocksPool).LogToConsole().Build()
	net := network.ATestNetwork(4, blocksPool...)
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
