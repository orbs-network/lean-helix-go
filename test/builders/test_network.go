package builders

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type TestNetwork struct {
	Nodes      []*Node
	BlocksPool []leanhelix.Block
	Discovery  *gossip.Discovery
}

func (net *TestNetwork) GetNodeGossip(publicKey Ed25519PublicKey) *gossip.Gossip {
	return net.Discovery.GetGossipByPK(publicKey)
}

func (net *TestNetwork) TriggerElection() {
	for _, node := range net.Nodes {
		node.TriggerElection()
	}
}

func (net *TestNetwork) StartConsensusOnAllNodes(ctx context.Context) error {
	for _, node := range net.Nodes {
		node.StartConsensus(ctx)
	}
	return nil
}

func (net *TestNetwork) RegisterNode(node *Node) {
	net.Nodes = append(net.Nodes, node)
}

func (net *TestNetwork) RegisterNodes(nodes []*Node) {
	for _, node := range nodes {
		net.RegisterNode(node)
	}
}

func (net *TestNetwork) ShutDown() {
	for _, node := range net.Nodes {
		node.Dispose()
	}

}

func NewTestNetwork(discovery *gossip.Discovery, blocksPool []leanhelix.Block) *TestNetwork {
	return &TestNetwork{
		Nodes:      []*Node{},
		BlocksPool: blocksPool,
		Discovery:  discovery,
	}
}
