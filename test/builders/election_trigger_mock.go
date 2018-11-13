package builders

import (
	"context"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTriggerMock struct {
	cancel      func()
	viewChannel chan View
}

func NewMockElectionTrigger() *ElectionTriggerMock {
	return &ElectionTriggerMock{
		viewChannel: make(chan View),
	}
}

func (et *ElectionTriggerMock) CreateElectionContextForView(parentContext context.Context, view View) context.Context {
	ctx, cancel := context.WithCancel(parentContext)
	et.cancel = cancel
	et.viewChannel <- view
	return ctx
}

func (et *ElectionTriggerMock) ManualTrigger() {
	cancel := et.cancel
	if cancel == nil {
		panic("You triggered the election before term was initialized")
	}
	cancel()
}

func (et *ElectionTriggerMock) WaitForNextView() View {
	return <-et.viewChannel
}
