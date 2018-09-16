package builders

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
)

type MockNetworkCommunication struct {
	mock.Mock
	nodes []Node
}

func (net *MockNetworkCommunication) Nodes() []Node {
	return net.nodes
}

func (net *MockNetworkCommunication) SendToMembers(publicKeys []string, messageType string, message []byte) {
	panic("implement me")
}

func NewMockNetworkCommunication(nodeCount int) lh.NetworkCommunication {

	nodes := make([]Node, 0)

	return &MockNetworkCommunication{
		nodes: nodes,
	}
}
