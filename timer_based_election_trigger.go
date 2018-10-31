package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/primitives"
	"math"
	"time"
)

// TODO What to do in the infinite loop when context is cancelled?
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
	minTimeout time.Duration
	view       primitives.View
	isActive   bool
	cb         func(view primitives.View)
	clearTimer chan bool
}

func NewTimerBasedElectionTrigger(minTimeout time.Duration) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		minTimeout: minTimeout,
	}
}

func (t *TimerBasedElectionTrigger) RegisterOnTrigger(view primitives.View, cb func(view primitives.View)) {
	t.cb = cb
	if !t.isActive || t.view != view {
		t.isActive = true
		t.view = view
		t.stop()
		timeoutMultiplier := time.Duration(int64(math.Pow(2, float64(view))))
		timeoutForView := timeoutMultiplier * t.minTimeout
		t.clearTimer = setTimeout(t.onTimeout, timeoutForView)
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
