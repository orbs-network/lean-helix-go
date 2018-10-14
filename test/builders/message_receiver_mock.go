package builders

import (
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
)

type MockMessageReceiver struct {
	mock.Mock
}

func NewMockMessageReceiver() *MockMessageReceiver {
	return &MockMessageReceiver{}
}

func (rec *MockMessageReceiver) OnReceive(message []byte) error {
	ret := rec.Called(message)
	return ret.Error(0)
}

func (rec *MockMessageReceiver) OnReceiveWithBlock(message []byte, block leanhelix.Block) error {
	ret := rec.Called(message, block)
	return ret.Error(0)
}
