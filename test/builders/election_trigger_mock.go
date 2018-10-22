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

func (electionTrigger *ElectionTriggerMock) RegisterOnTrigger(view View, cb func(view View)) {
	electionTrigger.view = view
	electionTrigger.cb = cb
}

func (electionTrigger *ElectionTriggerMock) UnregisterOnTrigger() {
	electionTrigger.cb = nil
}

// TODO: Gil - what to put as arg to cb()
func (electionTrigger *ElectionTriggerMock) Trigger() {
	if electionTrigger.cb != nil {
		electionTrigger.cb(electionTrigger.view)
	}
}
