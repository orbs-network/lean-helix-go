// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package Electiontrigger

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/metrics"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
	"math"
	"time"
)

var TIMEOUT_EXP_BASE = 2.0

type TimerBasedElectionTrigger struct {
	electionChannel chan *interfaces.ElectionTrigger
	minTimeout      time.Duration
	view            primitives.View
	blockHeight     primitives.BlockHeight
	electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics))
	onElectionCB    func(m metrics.ElectionMetrics)
	timer           *time.Timer
}

func NewTimerBasedElectionTrigger(minTimeout time.Duration, onElectionCB func(m metrics.ElectionMetrics)) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		electionChannel: make(chan *interfaces.ElectionTrigger, 0), // Caution - keep 0 to make election channel blocking
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
		t.timer = time.AfterFunc(timeout, func() {
			if ctx.Err() == nil {
				t.onTimerTimeout() // prevent running this after test code is complete (due to test error: "Log in goroutine after test has completed")
			}
		})
	}
	t.electionHandler = electionHandler
}

func (t *TimerBasedElectionTrigger) ElectionChannel() chan *interfaces.ElectionTrigger {
	return t.electionChannel
}

func (t *TimerBasedElectionTrigger) Stop() {
	t.electionHandler = nil
	if t.timer != nil {
		active := t.timer.Stop()
		if !active {
			select {
			case <-t.timer.C:
			default:
			}
		}
		t.timer = nil
	}
}

func (t *TimerBasedElectionTrigger) runOnReadElectionChannel(ctx context.Context) {
	if t.electionHandler != nil {
		t.electionHandler(ctx, t.blockHeight, t.view, t.onElectionCB)
	}
}

func (t *TimerBasedElectionTrigger) onTimerTimeout() {
	t.electionChannel <- &interfaces.ElectionTrigger{
		MoveToNextLeader: t.runOnReadElectionChannel,
		Hv:               state.NewHeightView(t.blockHeight, t.view),
	}

}

func (t *TimerBasedElectionTrigger) CalcTimeout(view primitives.View) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(2, float64(view))))
	return timeoutMultiplier * t.minTimeout
}
