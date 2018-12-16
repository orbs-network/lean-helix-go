package builders

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type ElectionTriggerMock struct {
	blockHeight     primitives.BlockHeight
	view            primitives.View
	cb              func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View)
	electionChannel chan func(ctx context.Context)
}

func NewMockElectionTrigger() *ElectionTriggerMock {
	return &ElectionTriggerMock{
		electionChannel: make(chan func(ctx context.Context)),
	}
}

func (et *ElectionTriggerMock) RegisterOnElection(blockHeight primitives.BlockHeight, view primitives.View, cb func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View)) {
	et.view = view
	et.blockHeight = blockHeight
	et.cb = cb
}

func (et *ElectionTriggerMock) ElectionChannel() chan func(ctx context.Context) {
	return et.electionChannel
}

func (et *ElectionTriggerMock) ManualTrigger() {
	go func() {
		et.electionChannel <- func(ctx context.Context) {
			if et.cb != nil {
				et.cb(ctx, et.blockHeight, et.view)
			}
		}
	}()
}

func (et *ElectionTriggerMock) ManualTriggerSync(ctx context.Context) {
	if et.cb != nil {
		et.cb(ctx, et.blockHeight, et.view)
	}
}
