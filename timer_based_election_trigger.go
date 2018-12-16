package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"math"
	"time"
)

func setTimeout(cb func(), timeout time.Duration) chan bool {
	timer := time.NewTimer(timeout)
	clear := make(chan bool)

	go func() {
		for {
			select {
			case <-timer.C:
				cb()
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
	cb              func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View)
	clearTimer      chan bool
}

func NewTimerBasedElectionTrigger(minTimeout time.Duration) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		electionChannel: make(chan func(ctx context.Context)),
		minTimeout:      minTimeout,
		firstTime:       true,
	}
}

func (t *TimerBasedElectionTrigger) RegisterOnElection(blockHeight primitives.BlockHeight, view primitives.View, cb func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View)) {
	t.cb = cb
	if t.firstTime || t.view != view || t.blockHeight != blockHeight {
		t.firstTime = false
		t.view = view
		t.blockHeight = blockHeight
		t.stop()
		t.clearTimer = setTimeout(t.onTimeout, t.calcTimeout(view))
	}
}

func (t *TimerBasedElectionTrigger) ElectionChannel() chan func(ctx context.Context) {
	return t.electionChannel
}

func (t *TimerBasedElectionTrigger) stop() {
	if t.clearTimer != nil {
		t.clearTimer <- true
		t.clearTimer = nil
	}
}

func (t *TimerBasedElectionTrigger) trigger(ctx context.Context) {
	if t.cb != nil {
		t.cb(ctx, t.blockHeight, t.view)
	}
}

func (t *TimerBasedElectionTrigger) onTimeout() {
	t.electionChannel <- t.trigger
}

func (t *TimerBasedElectionTrigger) calcTimeout(view primitives.View) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(2, float64(view))))
	return timeoutMultiplier * t.minTimeout
}
