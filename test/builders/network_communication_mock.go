package builders

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
)

type MockNetworkCommunication struct {
	mock.Mock
}

func (n *MockNetworkCommunication) Send(publicKeys []lh.PublicKey, message []byte) error {
	ret := n.Called(publicKeys, message)
	return ret.Error(0)
}

func (n *MockNetworkCommunication) SendWithBlock(publicKeys []lh.PublicKey, message []byte, block lh.Block) error {
	ret := n.Called(publicKeys, message, block)
	return ret.Error(0)
}

func (n *MockNetworkCommunication) SendToMembers(publicKeys []lh.PublicKey, messageType string, message []lh.MessageTransporter) {
	panic("implement me")
}

func (n *MockNetworkCommunication) RequestOrderedCommittee(seed uint64) []lh.PublicKey {
	ret := n.Called(seed)
	return ret.Get(0).([]lh.PublicKey)
}

func (n *MockNetworkCommunication) IsMember(pk lh.PublicKey) bool {
	panic("implement me")
}

func (n *MockNetworkCommunication) SendPreprepare(pks []lh.PublicKey, message lh.PreprepareMessage) {
	n.Called(pks, message)
}

func (n *MockNetworkCommunication) SendPrepare(pks []lh.PublicKey, message lh.PrepareMessage) {
	panic("implement me")
}

func (n *MockNetworkCommunication) SendCommit(pks []lh.PublicKey, message lh.CommitMessage) {
	panic("implement me")
}

func (n *MockNetworkCommunication) SendViewChange(pk lh.PublicKey, message lh.ViewChangeMessage) {
	panic("implement me")
}

func (n *MockNetworkCommunication) SendNewView(pks []lh.PublicKey, message lh.NewViewMessage) {
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
