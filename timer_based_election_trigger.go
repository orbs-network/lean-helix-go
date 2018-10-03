package leanhelix

import (
	"math"
	"time"
)

func setTimeout(cb func(), milliseconds uint) chan bool {
	interval := time.Duration(milliseconds) * time.Millisecond
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
	minTimeout uint
	view       View
	isActive   bool
	cb         func(view View)
	clearTimer chan bool
}

func NewTimerBasedElectionTrigger(minTimeout uint) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		minTimeout: minTimeout,
	}
}

func (tbet *TimerBasedElectionTrigger) RegisterOnTrigger(view View, cb func(view View)) {
	tbet.cb = cb
	if !tbet.isActive || tbet.view != view {
		tbet.isActive = true
		tbet.view = view
		tbet.stop()
		timeout := uint(math.Pow(2, float64(view))) * tbet.minTimeout
		tbet.clearTimer = setTimeout(tbet.onTimeout, timeout)
	}
}

func (tbet *TimerBasedElectionTrigger) UnregisterOnTrigger() {
	tbet.cb = nil
	tbet.isActive = false
	tbet.stop()
}

func (tbet *TimerBasedElectionTrigger) stop() {
	if tbet.clearTimer != nil {
		tbet.clearTimer <- true
		tbet.clearTimer = nil
	}
}

func (tbet *TimerBasedElectionTrigger) onTimeout() {
	if tbet.cb != nil {
		tbet.cb(tbet.view)
	}
}
