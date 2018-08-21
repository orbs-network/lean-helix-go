package builders

import (
	"github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/stretchr/testify/mock"
)

type MockNetworkCommunication struct {
	mock.Mock
	nodes []leanhelix.Node
}

func (net *MockNetworkCommunication) Nodes() []leanhelix.Node {
	return net.nodes
}

func (net *MockNetworkCommunication) SendToMembers(publicKeys []string, messageType string, message []byte) {
	panic("implement me")
}

func NewMockNetworkCommunication(nodeCount int) leanhelix.NetworkCommunication {

	nodes := make([]leanhelix.Node, 0)

	return &MockNetworkCommunication{
		nodes: nodes,
	}
}
