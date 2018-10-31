package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/primitives"
	"math"
	"time"
)

// TODO What to do in the infinite loop when context is cancelled?
func setTimeoutMillis(cb func(), timeoutMillis uint64) chan bool {
	interval := time.Duration(timeoutMillis) * time.Millisecond
	timer := time.NewTimer(interval)
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
	minTimeoutMillis uint64
	view             primitives.View
	isActive         bool
	cb               func(view primitives.View)
	clearTimer       chan bool
}

func NewTimerBasedElectionTrigger(minTimeout uint64) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		minTimeoutMillis: minTimeout,
	}
}

func (t *TimerBasedElectionTrigger) RegisterOnTrigger(view primitives.View, cb func(view primitives.View)) {
	t.cb = cb
	if !t.isActive || t.view != view {
		t.isActive = true
		t.view = view
		t.stop()
		timeoutMillis := uint64(math.Pow(2, float64(view))) * t.minTimeoutMillis
		t.clearTimer = setTimeoutMillis(t.onTimeout, timeoutMillis)
	}
}

func (t *TimerBasedElectionTrigger) UnregisterOnTrigger() {
	t.cb = nil
	t.isActive = false
	t.stop()
}

func (t *TimerBasedElectionTrigger) stop() {
	if t.clearTimer != nil {
		t.clearTimer <- true
		t.clearTimer = nil
	}
}

func (t *TimerBasedElectionTrigger) onTimeout() {
	if t.cb != nil {
		t.cb(t.view)
	}
}
