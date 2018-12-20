package consensusmessagefilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
)

type termMessagesHandlerMock struct {
	history []leanhelix.ConsensusMessage
}

func NewTermMessagesHandlerMock() *termMessagesHandlerMock {
	return &termMessagesHandlerMock{}
}

func (t *termMessagesHandlerMock) HandleTermMessages(ctx context.Context, message leanhelix.ConsensusMessage) {
	t.history = append(t.history, message)
}
