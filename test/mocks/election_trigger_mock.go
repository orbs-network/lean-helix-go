package mocks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/metrics"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"time"
)

type ElectionTriggerMock struct {
	blockHeight     primitives.BlockHeight
	view            primitives.View
	electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics))
	electionChannel chan func(ctx context.Context)
}

func (et *ElectionTriggerMock) CalcTimeout(view primitives.View) time.Duration {
	return 1 * time.Millisecond // dummy
}

func NewMockElectionTrigger() *ElectionTriggerMock {
	return &ElectionTriggerMock{
		electionChannel: make(chan func(ctx context.Context)),
	}
}

func (et *ElectionTriggerMock) RegisterOnElection(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics))) {
	et.view = view
	et.blockHeight = blockHeight
	et.electionHandler = electionHandler
}

func (et *ElectionTriggerMock) Stop() {
	et.electionHandler = nil
}

func (et *ElectionTriggerMock) ElectionChannel() chan func(ctx context.Context) {
	return et.electionChannel
}

func (et *ElectionTriggerMock) ManualTrigger(ctx context.Context) {
	go func() {
		select {
		case <-ctx.Done():
			return
		case et.electionChannel <- func(ctx context.Context) {
			if et.electionHandler != nil {
				et.electionHandler(ctx, et.blockHeight, et.view, nil)
			}
		}:
		}
	}()
}

func (et *ElectionTriggerMock) ManualTriggerSync(ctx context.Context) {
	if et.electionHandler != nil {
		et.electionHandler(ctx, et.blockHeight, et.view, nil)
	}
}
