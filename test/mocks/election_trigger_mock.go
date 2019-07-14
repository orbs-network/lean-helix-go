// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

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
		electionChannel: make(chan func(ctx context.Context), 0), // Caution - keep 0 to make elections channel blocking
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

func (et *ElectionTriggerMock) ManualTrigger(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			close(done)
		case et.electionChannel <- et.electionTriggerHandler:
			close(done)
		}
	}()
	return done
}

func (et *ElectionTriggerMock) electionTriggerHandler(ctx context.Context) {
	if et.electionHandler != nil {
		et.electionHandler(ctx, et.blockHeight, et.view, nil)
	}
}

func (et *ElectionTriggerMock) ManualTriggerSync(ctx context.Context) {
	if et.electionHandler != nil {
		et.electionHandler(ctx, et.blockHeight, et.view, nil)
	}
}
