package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
	"math"
	"time"
)

type TimerBasedElectionTrigger struct {
	minTimeout        time.Duration
	electionTimeoutAt time.Time
	view              primitives.View
}

func NewTimerBasedElectionTrigger(minTimeout time.Duration) *TimerBasedElectionTrigger {
	return &TimerBasedElectionTrigger{
		minTimeout: minTimeout,
		view:       100,
	}
}

func (t *TimerBasedElectionTrigger) CreateElectionContext(parentContext context.Context, view primitives.View) context.Context {
	if t.view != view {
		t.view = view
		t.electionTimeoutAt = time.Now().Add(t.calcTimeout(view))
	}

	ctx, _ := context.WithDeadline(parentContext, t.electionTimeoutAt)
	return ctx
}

func (t *TimerBasedElectionTrigger) calcTimeout(view primitives.View) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(2, float64(view))))
	return timeoutMultiplier * t.minTimeout
}
