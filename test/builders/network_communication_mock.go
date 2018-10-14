package builders

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type MockNetworkCommunication struct {
	mock.Mock
}

func (n *MockNetworkCommunication) Send(publicKeys []Ed25519PublicKey, message []byte) error {
	ret := n.Called(publicKeys, message)
	return ret.Error(0)
}

func (n *MockNetworkCommunication) SendWithBlock(publicKeys []Ed25519PublicKey, message []byte, block lh.Block) error {
	ret := n.Called(publicKeys, message, block)
	return ret.Error(0)
}

func (n *MockNetworkCommunication) SendToMembers(publicKeys []Ed25519PublicKey, messageType string, message []lh.MessageTransporter) {
	panic("implement me")
}

func (n *MockNetworkCommunication) RequestOrderedCommittee(seed uint64) []Ed25519PublicKey {
	ret := n.Called(seed)
	return ret.Get(0).([]Ed25519PublicKey)
}

func (n *MockNetworkCommunication) IsMember(pk Ed25519PublicKey) bool {
	panic("implement me")
}

func (n *MockNetworkCommunication) SendPreprepare(pks []Ed25519PublicKey, message lh.PreprepareMessage) {
	n.Called(pks, message)
}

func (n *MockNetworkCommunication) SendPrepare(pks []Ed25519PublicKey, message lh.PrepareMessage) {
	panic("implement me")
}

func (n *MockNetworkCommunication) SendCommit(pks []Ed25519PublicKey, message lh.CommitMessage) {
	panic("implement me")
}

func (n *MockNetworkCommunication) SendViewChange(pk Ed25519PublicKey, message lh.ViewChangeMessage) {
	panic("implement me")
}

func (n *MockNetworkCommunication) SendNewView(pks []Ed25519PublicKey, message lh.NewViewMessage) {
	panic("implement me")
}

func (n *MockNetworkCommunication) RegisterToPreprepare(cb func(message lh.PreprepareMessage)) {
	panic("implement me")
}

func (n *MockNetworkCommunication) RegisterToPrepare(cb func(message lh.PrepareMessage)) {
	panic("implement me")
}

func (n *MockNetworkCommunication) RegisterToCommit(cb func(message lh.CommitMessage)) {
	panic("implement me")
}

func (n *MockNetworkCommunication) RegisterToViewChange(cb func(message lh.ViewChangeMessage)) {
	panic("implement me")
}

func (n *MockNetworkCommunication) RegisterToNewView(cb func(message lh.NewViewMessage)) {
	panic("implement me")
}

func NewMockNetworkCommunication() *MockNetworkCommunication {

	return &MockNetworkCommunication{}
}
