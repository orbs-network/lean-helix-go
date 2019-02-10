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
	firstTime       bool
	electionHandler func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics))
	onElectionCB    func(m metrics.ElectionMetrics)
	timer           *time.Timer
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
		t.restartTimer(ctx, t.onTimeout, t.CalcTimeout(view))
	}
}

func (t *TimerBasedElectionTrigger) ElectionChannel() chan func(ctx context.Context) {
	return t.electionChannel
}

func (t *TimerBasedElectionTrigger) tryStop() {
	if t.timer != nil {
		t.timer.Stop()
	}
}

func (t *TimerBasedElectionTrigger) trigger(ctx context.Context) {
	if t.electionHandler != nil {
		t.electionHandler(ctx, t.blockHeight, t.view, t.onElectionCB)
	}
}

func (t *TimerBasedElectionTrigger) onTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case t.electionChannel <- t.trigger:
	}
}

func (t *TimerBasedElectionTrigger) restartTimer(ctx context.Context, cb func(ctx context.Context), timeout time.Duration) {

	t.tryStop()
	t.timer = time.NewTimer(timeout)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.timer.C:
				cb(ctx)
				return
			}
		}
	}()
}

func (t *TimerBasedElectionTrigger) CalcTimeout(view primitives.View) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(TIMEOUT_EXP_BASE, float64(view))))
	return timeoutMultiplier * t.minTimeout
}
