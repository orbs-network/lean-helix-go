package leanhelix

import "github.com/orbs-network/lean-helix-go/types"

type ElectionTriggerMock struct {
	view types.ViewCounter
	cb   func(view types.ViewCounter)
}

func NewElectionTriggerMock() *ElectionTriggerMock {
	return &ElectionTriggerMock{}
}

func (tbet *ElectionTriggerMock) RegisterOnTrigger(view types.ViewCounter, cb func(view types.ViewCounter)) {
	tbet.view = view
	tbet.cb = cb
}

func (tbet *ElectionTriggerMock) UnregisterOnTrigger() {
	tbet.cb = nil
}

// TODO: Gil - what to put as arg to cb()
func (tbet *ElectionTriggerMock) Trigger() {
	if tbet.cb != nil {
		tbet.cb(tbet.view)
	}
}
