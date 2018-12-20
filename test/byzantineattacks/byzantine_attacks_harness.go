package byzantineattacks

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
