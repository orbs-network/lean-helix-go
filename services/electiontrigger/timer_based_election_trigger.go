package electiontrigger

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/metrics"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"math"
	"time"
)

var TIMEOUT_EXP_BASE = float64(1.1) // By default it is 2.0

func setTimeout(ctx context.Context, cb func(ctx context.Context), timeout time.Duration) chan bool {
	timer := time.NewTimer(timeout)
	clear := make(chan bool)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				cb(ctx)
				return
			case <-clear:
				timer.Stop()
				return
			}

		}
	}()

	return clear
}

type TimerBasedElectionTrigger struct {
	electionChannel chan func(ctx context.Context)
	minTimeout      time.Duration
	view            primitives.View
	blockHeight     primitives.BlockHeight
	firstTime       bool
	electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics))
	onElectionCB    func(m metrics.ElectionMetrics)
	clearTimer      chan bool
}

func NewTimerBasedElectionTrigger(minTimeout time.Duration, onElectionCB func(m metrics.ElectionMetrics)) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		electionChannel: make(chan func(ctx context.Context)),
		minTimeout:      minTimeout,
		firstTime:       true,
		onElectionCB:    onElectionCB,
	}
}

func (t *TimerBasedElectionTrigger) RegisterOnElection(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics))) {
	t.electionHandler = electionHandler
	if t.firstTime || t.view != view || t.blockHeight != blockHeight {
		t.firstTime = false
		t.view = view
		t.blockHeight = blockHeight
		t.stop(ctx)
		t.clearTimer = setTimeout(ctx, t.onTimeout, t.CalcTimeout(view))
	}
}

func (t *TimerBasedElectionTrigger) ElectionChannel() chan func(ctx context.Context) {
	return t.electionChannel
}

func (t *TimerBasedElectionTrigger) stop(ctx context.Context) {
	if t.clearTimer != nil {
		select {
		case <-ctx.Done():
			return
		case t.clearTimer <- true:
			t.clearTimer = nil
		}
	}
}

func (t *TimerBasedElectionTrigger) trigger(ctx context.Context) {
	if t.electionHandler != nil {
		t.electionHandler(ctx, t.blockHeight, t.view, t.onElectionCB)
	}
}

func (t *TimerBasedElectionTrigger) onTimeout(ctx context.Context) {
	t.clearTimer = nil
	select {
	case <-ctx.Done():
		return
	case t.electionChannel <- t.trigger:
	}
}

func (t *TimerBasedElectionTrigger) CalcTimeout(view primitives.View) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(TIMEOUT_EXP_BASE, float64(view))))
	return timeoutMultiplier * t.minTimeout
}
