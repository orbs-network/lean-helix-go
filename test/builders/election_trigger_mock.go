package builders

import "github.com/orbs-network/lean-helix-go/primitives"

type ElectionTriggerMock struct {
	view            primitives.View
	cb              func(view primitives.View)
	electionChannel chan func()
}

func NewMockElectionTrigger() *ElectionTriggerMock {
	return &ElectionTriggerMock{
		electionChannel: make(chan func()),
	}
}

func (et *ElectionTriggerMock) RegisterOnElection(view primitives.View, cb func(view primitives.View)) {
	et.view = view
	et.cb = cb
}

func (et *ElectionTriggerMock) ElectionChannel() chan func() {
	return et.electionChannel
}

func (et *ElectionTriggerMock) ManualTrigger() {
	et.electionChannel <- func() {
		if et.cb != nil {
			et.cb(et.view)
		}
	}
}
