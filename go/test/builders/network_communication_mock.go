package builders

import (
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/test/network"
)

type MockNetworkCommunication struct {
	mock.Mock
	nodes []network.Node
}

func (net *MockNetworkCommunication) Nodes() []network.Node {
	return net.nodes
}

func (net *MockNetworkCommunication) SendToMembers(publicKeys []string, messageType string, message []byte) {
	panic("implement me")
}

func NewMockNetworkCommunication(nodeCount int) leanhelix.NetworkCommunication {

	nodes := make([]network.Node, 0)

	return &MockNetworkCommunication{
		nodes: nodes,
	}
}
