package builders

import lh "github.com/orbs-network/lean-helix-go"

type ElectionTriggerMock struct {
	view lh.View
	cb   func(view lh.View)
}

func NewMockElectionTrigger() *ElectionTriggerMock {
	return &ElectionTriggerMock{}
}

func (tbet *ElectionTriggerMock) RegisterOnTrigger(view lh.View, cb func(view lh.View)) {
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
