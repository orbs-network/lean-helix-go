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
	electionChannel  chan *interfaces.ElectionTrigger
	minTimeout       time.Duration
	electionHandler  func(blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback)
	callbackFromOrbs interfaces.OnElectionCallback
	timer            *time.Timer

	// mutable, mutex protected - better refactor into separate obj
	lock             sync.RWMutex
	blockHeight      primitives.BlockHeight
	view             primitives.View
	triggerCancelled chan struct{}
}

func NewTimerBasedElectionTrigger(minTimeout time.Duration, callbackFromOrbs interfaces.OnElectionCallback) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		electionChannel:  make(chan *interfaces.ElectionTrigger), // Caution - keep 0 to make election channel blocking
		minTimeout:       minTimeout,
		callbackFromOrbs: callbackFromOrbs,
	}
}

// on new view
func (t *TimerBasedElectionTrigger) RegisterOnElection(blockHeight primitives.BlockHeight, view primitives.View, moveToNextLeader func(blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback)) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.electionHandler != nil && t.view == view && t.blockHeight == blockHeight {
		return
	}

	timeout := t.CalcTimeout(view)
	t.view = view
	t.blockHeight = blockHeight
	t.Stop()

	triggerCancelled := make(chan struct{})
	t.triggerCancelled = triggerCancelled
	t.timer = time.AfterFunc(timeout, func() {
		triggerElections(t.ElectionChannel(), blockHeight, view, triggerCancelled, func(ctx context.Context) {
			if moveToNextLeader != nil {
				moveToNextLeader(blockHeight, view, t.callbackFromOrbs) // executed by LH worker loop
			}
		})
	})

	t.electionHandler = moveToNextLeader
}

func (t *TimerBasedElectionTrigger) ElectionChannel() chan *interfaces.ElectionTrigger {
	return t.electionChannel
}

func (t *TimerBasedElectionTrigger) Stop() {
	t.electionHandler = nil
	if t.timer != nil {
		onTimerTimeoutIsAlreadyRunning := !t.timer.Stop()
		if onTimerTimeoutIsAlreadyRunning {
			close(t.triggerCancelled) // so that we do not write an irrelevant trigger to the election channel
			select {                  // levison voodoo
			case <-t.timer.C:
			default:
			}
		}
		t.timer = nil
	}
}

func (t *TimerBasedElectionTrigger) CalcTimeout(view primitives.View) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(TIMEOUT_EXP_BASE, float64(view))))
	return timeoutMultiplier * t.minTimeout
}

func triggerElections(electionChannel chan *interfaces.ElectionTrigger, height primitives.BlockHeight, view primitives.View, triggerCancelled chan struct{}, electionsFunc func(ctx context.Context)) {
	select {
	case <-triggerCancelled:
		return
	default:
	}

	select {
	case electionChannel <- &interfaces.ElectionTrigger{
		MoveToNextLeader: electionsFunc,
		Hv:               state.NewHeightView(height, view),
	}:
	case <-triggerCancelled:
	}

}
