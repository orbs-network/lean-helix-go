package electiontrigger

import (
	"time"
)

func setInterval(cb func(), milliseconds uint) chan bool {
	interval := time.Duration(milliseconds) * time.Millisecond
	ticker := time.NewTicker(interval)
	clear := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				cb()
			case <-clear:
				ticker.Stop()
				return
			}

		}
	}()

	return clear
}

type TimerBasedElectionTrigger struct {
	timeout    uint
	clearTimer chan bool
}

func NewTimerBasedElectionTrigger(timeout uint) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		timeout: timeout,
	}
}

func (tbet *TimerBasedElectionTrigger) Start(cb func()) {
	tbet.clearTimer = setInterval(cb, tbet.timeout)
}

func (tbet *TimerBasedElectionTrigger) Stop() {
	if tbet.clearTimer != nil {
		tbet.clearTimer <- true
		tbet.clearTimer = nil
	}
}
