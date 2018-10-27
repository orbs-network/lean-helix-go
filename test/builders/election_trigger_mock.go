package builders

import (
	"context"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTriggerMock struct {
	view View
	cb   func(ctx context.Context, view View)
}

func NewMockElectionTrigger() *ElectionTriggerMock {
	return &ElectionTriggerMock{}
}

func (electionTrigger *ElectionTriggerMock) RegisterOnTrigger(view View, cb func(ctx context.Context, view View)) {
	electionTrigger.view = view
	electionTrigger.cb = cb
}

func (electionTrigger *ElectionTriggerMock) UnregisterOnTrigger() {
	electionTrigger.cb = nil
}

func (electionTrigger *ElectionTriggerMock) Trigger(ctx context.Context) {
	if electionTrigger.cb != nil {
		electionTrigger.cb(ctx, electionTrigger.view)
	}
}
