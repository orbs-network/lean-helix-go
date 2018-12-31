package byzantineattacks

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

func NewHarness(ctx context.Context, t *testing.T, totalNodes int, blocksPool ...interfaces.Block) *harness {
	net := network.ATestNetwork(totalNodes, blocksPool...)
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
