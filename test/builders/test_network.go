package builders

import (
	"context"
	"fmt"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

const MINIMUM_NODES = 2

type TestNetwork struct {
	Nodes     []*Node
	Discovery gossip.Discovery
}

func (net *TestNetwork) GetNodeGossip(pk Ed25519PublicKey) *gossip.Gossip {
	return net.Discovery.GetGossipByPK(pk)
}

func (net *TestNetwork) TriggerElection(ctx context.Context) {
	for _, node := range net.Nodes {
		node.TriggerElection(ctx)
	}
}

func (net *TestNetwork) StartConsensusOnAllNodes() error {
	if len(net.Nodes) < MINIMUM_NODES {
		return fmt.Errorf("not enough nodes in test network - found %d but minimum is %d", len(net.Nodes), MINIMUM_NODES)
	}
	for _, node := range net.Nodes {
		node.StartConsensus()
	}
	return nil
}

func (net *TestNetwork) ShutDown() {
	// TODO Do we need this??
	for _, node := range net.Nodes {
		node.Dispose()
	}

}
