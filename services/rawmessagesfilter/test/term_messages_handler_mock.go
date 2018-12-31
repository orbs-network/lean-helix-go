package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type termMessagesHandlerMock struct {
	history []interfaces.ConsensusMessage
}

func NewTermMessagesHandlerMock() *termMessagesHandlerMock {
	return &termMessagesHandlerMock{}
}

func (t *termMessagesHandlerMock) HandleConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage) {
	t.history = append(t.history, message)
}
