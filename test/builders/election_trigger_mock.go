package builders

import (
	"context"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTriggerMock struct {
	TickSns     *Sns
	PauseOnTick bool
	cancel      func()
}

func NewMockElectionTrigger(pauseOnTick bool) *ElectionTriggerMock {
	return &ElectionTriggerMock{
		TickSns:     NewSignalAndStop(),
		PauseOnTick: pauseOnTick,
	}
}

func (et *ElectionTriggerMock) CreateElectionContextForView(parentContext context.Context, view View) context.Context {
	ctx, cancel := context.WithCancel(parentContext)
	et.cancel = cancel
	if et.PauseOnTick {
		et.TickSns.SignalAndStop()
	}

	return ctx
}

func (et *ElectionTriggerMock) ManualTrigger() {
	cancel := et.cancel
	if cancel == nil {
		panic("You triggered the election before term was initialized")
	}
	cancel()
}
