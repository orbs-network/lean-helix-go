package builders

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTriggerMock struct {
	view            primitives.View
	cb              func(ctx context.Context, view primitives.View)
	electionChannel chan func(ctx context.Context)
}

func NewMockElectionTrigger() *ElectionTriggerMock {
	return &ElectionTriggerMock{
		electionChannel: make(chan func(ctx context.Context)),
	}
}

func (et *ElectionTriggerMock) RegisterOnElection(view primitives.View, cb func(ctx context.Context, view primitives.View)) {
	et.view = view
	et.cb = cb
}

func (et *ElectionTriggerMock) ElectionChannel() chan func(ctx context.Context) {
	return et.electionChannel
}

func (et *ElectionTriggerMock) ManualTrigger() {
	et.electionChannel <- func(ctx context.Context) {
		if et.cb != nil {
			et.cb(ctx, et.view)
		}
	}
}
