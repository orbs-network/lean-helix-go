// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package electiontrigger

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/metrics"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"math"
	"time"
)

var TIMEOUT_EXP_BASE = float64(2.0)

type TimerBasedElectionTrigger struct {
	electionChannel chan func(ctx context.Context)
	minTimeout      time.Duration
	view            primitives.View
	blockHeight     primitives.BlockHeight
	electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics))
	onElectionCB    func(m metrics.ElectionMetrics)
	triggerTimer    *time.Timer
}

func NewTimerBasedElectionTrigger(minTimeout time.Duration, onElectionCB func(m metrics.ElectionMetrics)) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		electionChannel: make(chan func(ctx context.Context), 0), // Caution - keep 0 to make election channel blocking
		minTimeout:      minTimeout,
		onElectionCB:    onElectionCB,
	}
}

func (t *TimerBasedElectionTrigger) RegisterOnElection(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics))) {
	if t.electionHandler == nil || t.view != view || t.blockHeight != blockHeight {
		timeout := t.CalcTimeout(view)
		t.view = view
		t.blockHeight = blockHeight
		t.Stop()
		t.triggerTimer = time.AfterFunc(timeout, t.sendTrigger)
	}
	t.electionHandler = electionHandler
}

func (t *TimerBasedElectionTrigger) ElectionChannel() chan func(ctx context.Context) {
	return t.electionChannel
}

func (t *TimerBasedElectionTrigger) Stop() {
	t.electionHandler = nil
	if t.triggerTimer != nil {
		active := t.triggerTimer.Stop()
		if !active {
			select {
			case <-t.triggerTimer.C:
			default:
			}
		}
		t.triggerTimer = nil
	}
}

func (t *TimerBasedElectionTrigger) trigger(ctx context.Context) {
	if t.electionHandler != nil {
		t.electionHandler(ctx, t.blockHeight, t.view, t.onElectionCB)
	}
}

func (t *TimerBasedElectionTrigger) sendTrigger() {
	t.electionChannel <- t.trigger
}

func (t *TimerBasedElectionTrigger) CalcTimeout(view primitives.View) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(2, float64(view))))
	return timeoutMultiplier * t.minTimeout
}
