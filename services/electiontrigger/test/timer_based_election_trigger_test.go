// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/metrics"
	"github.com/orbs-network/lean-helix-go/services/electiontrigger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func buildElectionTrigger(ctx context.Context, timeout time.Duration) *electiontrigger.TimerBasedElectionTrigger {
	et := electiontrigger.NewTimerBasedElectionTrigger(timeout, nil)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case trigger := <-et.ElectionChannel():
				trigger(ctx)
			}
		}
	}()

	return et
}

// TODO Consider removing this test entirely - sleeps in tests are bad
func TestCallbackTrigger(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 50*time.Millisecond)

		wasCalled := false
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
			wasCalled = true
		}
		et.RegisterOnElection(ctx, 20, 0, cb)

		time.Sleep(time.Duration(80) * time.Millisecond)

		require.True(t, wasCalled, "Did not call the timer callback")
	})
}

// TODO v1 this is sometimes flaky
func TestCallbackTriggerOnce(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 10*time.Millisecond)

		callCount := 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
			callCount++
		}
		et.RegisterOnElection(ctx, 10, 0, cb)

		time.Sleep(time.Duration(25) * time.Millisecond)

		require.Exactly(t, 1, callCount, "Trigger callback called more than once")
	})
}

func TestCallbackTriggerTwiceInARow(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 10*time.Millisecond)

		callCount := 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
			callCount++
		}
		et.RegisterOnElection(ctx, 10, 0, cb)

		time.Sleep(time.Duration(25) * time.Millisecond)

		et.RegisterOnElection(ctx, 11, 0, cb)
		time.Sleep(time.Duration(25) * time.Millisecond)

		require.Exactly(t, 2, callCount, "Trigger callback twice without getting stuck")
	})
}

func TestIgnoreSameViewOrHeight(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 30*time.Millisecond)

		callCount := 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
			callCount++
		}

		et.RegisterOnElection(ctx, 10, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 10, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 10, 0, cb)
		time.Sleep(time.Duration(20) * time.Millisecond)
		et.RegisterOnElection(ctx, 10, 0, cb)

		require.Exactly(t, 1, callCount, "Trigger callback called more than once")
	})
}

// TODO: Flaky
func TestNotTriggerIfSameViewButDifferentHeight(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 30*time.Millisecond)

		callCount := 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
			callCount++
		}

		et.RegisterOnElection(ctx, 10, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 11, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 12, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 13, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 14, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 15, 0, cb)

		require.Exactly(t, 0, callCount, "Trigger callback called")
	})
}

func TestNotTriggerIfSameHeightButDifferentView(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 30*time.Millisecond)

		callCount := 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
			callCount++
		}

		et.RegisterOnElection(ctx, 10, 0, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 10, 1, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 10, 2, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 10, 3, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 10, 4, cb)
		time.Sleep(time.Duration(10) * time.Millisecond)
		et.RegisterOnElection(ctx, 10, 5, cb)

		require.Exactly(t, 0, callCount, "Trigger callback called")
	})
}

func TestViewChanges(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 50*time.Millisecond)

		wasCalled := false
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
			wasCalled = true
		}

		et.RegisterOnElection(ctx, 10, 0, cb) // 2 ** 0 * 20 = 20
		time.Sleep(time.Duration(10) * time.Millisecond)

		et.RegisterOnElection(ctx, 10, 1, cb) // 2 ** 1 * 20 = 40
		time.Sleep(time.Duration(30) * time.Millisecond)

		et.RegisterOnElection(ctx, 10, 2, cb) // 2 ** 2 * 20 = 80
		time.Sleep(time.Duration(70) * time.Millisecond)

		et.RegisterOnElection(ctx, 10, 3, cb) // 2 ** 3 * 20 = 160

		require.False(t, wasCalled, "Trigger the callback even if a new Register was called with a new view")
	})
}

func TestViewPowTimeout(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 10*time.Millisecond)

		wasCalled := false
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
			wasCalled = true
		}

		et.RegisterOnElection(ctx, 10, 2, cb) // 2 ** 2 * 10 = 40
		time.Sleep(time.Duration(30) * time.Millisecond)
		require.False(t, wasCalled, "Triggered the callback too early")
		time.Sleep(time.Duration(30) * time.Millisecond)
		require.True(t, wasCalled, "Did not trigger the callback after the required timeout")
	})
}
