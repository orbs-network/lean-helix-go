package builders

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type MockNetworkCommunication struct {
	mock.Mock
}

func (n *MockNetworkCommunication) Send(publicKeys []Ed25519PublicKey, message lh.ConsensusRawMessage) error {
	ret := n.Called(publicKeys, message)
	return ret.Error(0)
}

func (n *MockNetworkCommunication) RequestOrderedCommittee(seed uint64) []Ed25519PublicKey {
	ret := n.Called(seed)
	return ret.Get(0).([]Ed25519PublicKey)
}

func (n *MockNetworkCommunication) IsMember(pk Ed25519PublicKey) bool {
	panic("implement me")
}

func NewMockNetworkCommunication() *MockNetworkCommunication {

	return &MockNetworkCommunication{}
}
