package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/gossip"
	"github.com/stretchr/testify/mock"
)

type TestNetwork struct {
	mock.Mock
	Nodes      []Node
	BlockUtils *MockBlockUtils
	Transport  *MockNetworkCommunication
	discovery  gossip.Discovery
}

func CreateTestNetwork(nodeCount int) *TestNetwork {

	nodes := make([]Node, nodeCount)
	discovery := gossip.NewGossipDiscovery()

	return &TestNetwork{
		Nodes:      nodes,
		BlockUtils: NewMockBlockUtils(),
		discovery:  discovery,
	}
}

func (net *TestNetwork) GetNodeGossip(pk leanhelix.PublicKey) (*gossip.Gossip, bool) {
	return net.discovery.GetGossipByPK(pk)
}

func (net *TestNetwork) Start() {

}

func (net *TestNetwork) Stop() {

}
