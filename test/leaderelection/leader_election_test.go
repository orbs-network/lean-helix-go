package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/orbs-network-go/test"
	"testing"
)

func waitForLeader(net *builders.TestNetwork, nodeIdx int, ctx context.Context) {
	node := net.Nodes[nodeIdx]
	for {
		if node.IsLeader() {
			break
		}
		node.Tick(ctx)
	}
}

func TestLeaderCircularOrdering(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork()
		net.StartConsensusSync()

		net.TriggerElection()
		waitForLeader(net, 0, ctx)

		net.TriggerElection()
		waitForLeader(net, 1, ctx)

		net.TriggerElection()
		waitForLeader(net, 2, ctx)

		net.TriggerElection()
		waitForLeader(net, 3, ctx)

		net.TriggerElection()
		waitForLeader(net, 0, ctx)
	})
}
