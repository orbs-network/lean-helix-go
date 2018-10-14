package builders

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTriggerMock struct {
	view View
	cb   func(view View)
}

func NewMockElectionTrigger() *ElectionTriggerMock {
	return &ElectionTriggerMock{}
}

func (tbet *ElectionTriggerMock) RegisterOnTrigger(view View, cb func(view View)) {
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
