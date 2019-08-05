// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package Electiontrigger

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
	"math"
	"sync"
	"time"
)

var TIMEOUT_EXP_BASE = 2.0

type TimerBasedElectionTrigger struct {
	electionChannel chan *interfaces.ElectionTrigger
	minTimeout      time.Duration
	electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback)
	onElectionCB    interfaces.OnElectionCallback
	timer           *time.Timer

	// mutable, mutex protected - better refactor into separate obj
	inRegister  bool
	blockHeight primitives.BlockHeight
	view        primitives.View
	lock        sync.RWMutex
}

func NewTimerBasedElectionTrigger(minTimeout time.Duration, onElectionCB interfaces.OnElectionCallback) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		electionChannel: make(chan *interfaces.ElectionTrigger, 0), // Caution - keep 0 to make election channel blocking
		minTimeout:      minTimeout,
		onElectionCB:    onElectionCB,
	}
}

func (t *TimerBasedElectionTrigger) RegisterOnElection(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback)) {
	if t.inRegister {
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.inRegister = true

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
	t.inRegister = false
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
	t.lock.RLock()
	h := t.blockHeight
	v := t.view
	handler := t.electionHandler
	t.lock.RUnlock()

	if handler != nil {
		handler(ctx, h, v, t.onElectionCB)
	}
}

func (t *TimerBasedElectionTrigger) onTimerTimeout() {
	t.lock.RLock()
	h := t.blockHeight
	v := t.view
	t.lock.RUnlock()

	t.electionChannel <- &interfaces.ElectionTrigger{
		MoveToNextLeader: t.runOnReadElectionChannel,
		Hv:               state.NewHeightView(h, v),
	}

}

func (t *TimerBasedElectionTrigger) CalcTimeout(view primitives.View) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(TIMEOUT_EXP_BASE, float64(view))))
	return timeoutMultiplier * t.minTimeout
}
