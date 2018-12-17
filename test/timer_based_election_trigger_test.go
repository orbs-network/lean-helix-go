package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/orbs-network-go/test"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func buildElectionTrigger(ctx context.Context, timeout time.Duration) *leanhelix.TimerBasedElectionTrigger {
	et := leanhelix.NewTimerBasedElectionTrigger(timeout)
	go func() {
		for {
			trigger := <-et.ElectionChannel()
			trigger(ctx)
		}
	}()

	return et
}

func TestCallbackTrigger(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 10*time.Millisecond)

		wasCalled := false
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View) { wasCalled = true }
		et.RegisterOnElection(10, 0, cb)

		time.Sleep(time.Duration(15) * time.Millisecond)

		require.True(t, wasCalled, "Did not call the timer callback")
	})
}

func TestCallbackTriggerOnce(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 10*time.Millisecond)

		callCount := 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View) { callCount++ }
		et.RegisterOnElection(10, 0, cb)

		time.Sleep(time.Duration(25) * time.Millisecond)

		require.Exactly(t, 1, callCount, "Trigger callback called more than once")
	})
}

func TestIgnoreSameViewOrHeight(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 30*time.Millisecond)

		callCount := 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View) { callCount++ }

		et.RegisterOnElection(10, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(10, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(10, 0, cb)
		time.Sleep(time.Duration(20) * time.Millisecond)
		et.RegisterOnElection(10, 0, cb)

		require.Exactly(t, 1, callCount, "Trigger callback called more than once")
	})
}

func TestNotTriggerIfSameViewButDifferentHeight(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 30*time.Millisecond)

		callCount := 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View) { callCount++ }

		et.RegisterOnElection(10, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(11, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(12, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(13, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(14, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(15, 0, cb)

		require.Exactly(t, 0, callCount, "Trigger callback called")
	})
}

func TestNotTriggerIfSameHeightButDifferentView(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 30*time.Millisecond)

		callCount := 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View) { callCount++ }

		et.RegisterOnElection(10, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(10, 1, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(10, 2, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(10, 3, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(10, 4, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(10, 5, cb)

		require.Exactly(t, 0, callCount, "Trigger callback called")
	})
}

func TestViewChanges(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 20*time.Millisecond)

		wasCalled := false
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View) { wasCalled = true }

		et.RegisterOnElection(10, 0, cb) // 2 ** 0 * 20 = 20
		time.Sleep(time.Duration(10) * time.Millisecond)

		et.RegisterOnElection(10, 1, cb) // 2 ** 1 * 20 = 40
		time.Sleep(time.Duration(30) * time.Millisecond)

		et.RegisterOnElection(10, 2, cb) // 2 ** 2 * 20 = 80
		time.Sleep(time.Duration(70) * time.Millisecond)

		et.RegisterOnElection(10, 3, cb) // 2 ** 3 * 20 = 160

		require.False(t, wasCalled, "Trigger the callback even if a new Register was called with a new view")
	})
}

func TestViewPowTimeout(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 10*time.Millisecond)

		wasCalled := false
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View) { wasCalled = true }

		et.RegisterOnElection(10, 2, cb) // 2 ** 2 * 10 = 40
		time.Sleep(time.Duration(30) * time.Millisecond)
		require.False(t, wasCalled, "Triggered the callback too early")
		time.Sleep(time.Duration(30) * time.Millisecond)
		require.True(t, wasCalled, "Did not trigger the callback after the required timeout")
	})
}
