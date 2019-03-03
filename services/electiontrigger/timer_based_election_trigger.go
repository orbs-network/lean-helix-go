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
		electionChannel: make(chan func(ctx context.Context)),
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
	timeoutMultiplier := time.Duration(int64(math.Pow(TIMEOUT_EXP_BASE, float64(view))))
	return timeoutMultiplier * t.minTimeout
}
