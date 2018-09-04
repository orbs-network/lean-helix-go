package electiontrigger

import lh "github.com/orbs-network/lean-helix-go/go/leanhelix"

type ElectionTriggerMock struct {
	cb func(view lh.ViewCounter)
}

func NewElectionTriggerMock() *ElectionTriggerMock {
	return &ElectionTriggerMock{}
}

func (tbet *ElectionTriggerMock) RegisterOnTrigger(view lh.ViewCounter, cb func(view lh.ViewCounter)) {
	tbet.cb = cb
}

func (tbet *ElectionTriggerMock) UnregisterOnTrigger() {
	tbet.cb = nil
}

// TODO: Gil - what to put as arg to cb()
func (tbet *ElectionTriggerMock) Trigger() {
	if tbet.cb != nil {
		tbet.cb()
	}
}
