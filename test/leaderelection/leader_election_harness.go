package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test/builders"
)

type harness struct {
	net *builders.TestNetwork
}

func NewHarness() *harness {
	net := builders.ABasicTestNetwork()
	net.StartConsensusSync()

	return &harness{
		net: net,
	}
}

func (h *harness) TriggerElection() {
	h.net.TriggerElection()
}

func (h *harness) waitForLeader(nodeIdx int, ctx context.Context) {
	node := h.net.Nodes[nodeIdx]
	for {
		if node.IsLeader() {
			break
		}
		node.Tick(ctx)
	}
}
