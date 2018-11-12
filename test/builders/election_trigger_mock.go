package builders

import (
	"context"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTriggerMock struct {
	cancel func()
}

func NewMockElectionTrigger() *ElectionTriggerMock {
	return &ElectionTriggerMock{}
}

func (electionTrigger *ElectionTriggerMock) CreateElectionContext(parentContext context.Context, view View) context.Context {
	ctx, cancel := context.WithCancel(parentContext)
	electionTrigger.cancel = cancel
	return ctx
}

func (electionTrigger *ElectionTriggerMock) Trigger() {
	cancel := electionTrigger.cancel
	if cancel != nil {
		cancel()
	}
}
