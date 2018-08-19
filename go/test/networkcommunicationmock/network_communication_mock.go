package networkcommunicationmock

type mockNetworkCommunication struct {
}

type NetworkCommunication interface {
	sendToMembers(publicKeys []string, messageType string, message []byte)
}

func NewMockNetworkCommunication() *mockNetworkCommunication {
	return &mockNetworkCommunication{}
}

func (*mockNetworkCommunication) sendToMembers(publicKeys []string, messageType string, message []byte) {
	panic("implement me")
}
