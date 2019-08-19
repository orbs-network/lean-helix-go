// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/electiontrigger"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func buildElectionTrigger(ctx context.Context, timeout time.Duration) *Electiontrigger.TimerBasedElectionTrigger {
	et := Electiontrigger.NewTimerBasedElectionTrigger(timeout, nil)
	ready := make(chan struct{})
	go func() {
		close(ready)
		for {
			select {
			case <-ctx.Done():
				return
			case trigger := <-et.ElectionChannel():
				trigger.MoveToNextLeader(ctx)
			}
		}
	}()
	<-ready
	return et
}

func TestCallbackTriggerOnce(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 1*time.Nanosecond)

		triggerReached := make(chan struct{})
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback) {
			triggerReached <- struct{}{}
		}
		et.RegisterOnElection(10, 0, cb)

		<-triggerReached

		timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 25*time.Millisecond)
		defer cancelTimeout()

		select {
		case <-triggerReached:
			t.Fatal("Trigger callback called more than once")
		case <-timeoutCtx.Done():

		}
	})
}

func TestCallbackTriggerTwiceInARow(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 1*time.Nanosecond)

		triggerReached := make(chan struct{})
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback) {
			triggerReached <- struct{}{}
		}
		et.RegisterOnElection(10, 0, cb)

		<-triggerReached

		et.RegisterOnElection(11, 0, cb)

		<-triggerReached

		timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 25*time.Millisecond)
		defer cancelTimeout()

		select {
		case <-triggerReached:
			t.Fatal("Trigger callback called more than once")
		case <-timeoutCtx.Done():
		}
	})
}

func TestIgnoreSameViewOrHeight(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 1*time.Nanosecond)

		var callCount int32 = 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback) {
			atomic.AddInt32(&callCount, 1)
		}

		et.RegisterOnElection(10, 0, cb)
		time.Sleep(10 * time.Millisecond)
		et.RegisterOnElection(10, 0, cb)
		time.Sleep(10 * time.Millisecond)
		et.RegisterOnElection(10, 0, cb)
		time.Sleep(20 * time.Millisecond)
		et.RegisterOnElection(10, 0, cb)

		atomic.LoadInt32(&callCount)
		require.Exactly(t, 1, int(callCount), "Trigger callback called more than once")
	})
}

func TestNotTriggeredIfSameViewButDifferentHeight(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		electionTimeout := 10 * time.Millisecond
		et := buildElectionTrigger(ctx, electionTimeout)

		cbNeverTriggered := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback) {
			t.Fatalf("Callback for H=%d V=%d", blockHeight, view)
		}

		cbNoop := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback) {
		}

		et.RegisterOnElection(1, 0, cbNeverTriggered)
		et.RegisterOnElection(2, 0, cbNoop)

		time.Sleep(2 * electionTimeout)
	})
}

func TestNotTriggerIfSameHeightButDifferentView(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 50*time.Millisecond)

		var callCount int32 = 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback) {
			atomic.AddInt32(&callCount, 1)
		}

		for i := 0; i < 5; i++ {
			et.RegisterOnElection(10, primitives.View(i), cb)
			time.Sleep(10 * time.Millisecond)
		}

		count := atomic.LoadInt32(&callCount)
		require.True(t, count <= 1, "Trigger callback called multiple times while expected 1 at most")
	})
}

func TestTimerBasedElectionTrigger_DidNotTriggerBeforeTimeout(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 10*time.Hour)

		var wasCalled int32 = 0
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback) {
			atomic.StoreInt32(&wasCalled, 1)
		}

		et.RegisterOnElection(10, 2, cb) // 2 ** 2 * 10h = 40h
		time.Sleep(30 * time.Millisecond)
		atomic.LoadInt32(&wasCalled)
		require.Equal(t, 0, int(wasCalled), "Triggered the callback too early")
	})
}

func TestViewPowTimeout_DidTriggerAfterTimeout(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 1*time.Millisecond)

		triggered := make(chan struct{})
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB interfaces.OnElectionCallback) {
			close(triggered)
		}

		et.RegisterOnElection(10, 2, cb) // 2 ** 2 * 1ms = 4ms

		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		select {
		case <-ctx.Done():
			return
		case <-timeoutCtx.Done():
			t.Fatal("Did not trigger the callback after the required timeout")
		case <-triggered:
			return
		}
	})
}

func TestElectionTrigger_Stress_FrequentRegisters(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 1*time.Microsecond)
		var counter int32
		for h := primitives.BlockHeight(1); h < primitives.BlockHeight(1000); h++ {
			et.RegisterOnElection(h, 0, nil)
			counter++
			time.Sleep(1 * time.Microsecond)
		}
	})

}

func TestElectionTrigger_StuckOnTimerTimeout_GetsReleasedByStop(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := Electiontrigger.NewTimerBasedElectionTrigger(100*time.Millisecond, nil)
		et.RegisterOnElection(0, 0, nil)
		time.Sleep(200 * time.Millisecond) // strictly higher than timeout
		et.RegisterOnElection(0, 1, nil)
		time.Sleep(200 * time.Millisecond) // strictly higher than timeout

		select {
		case trigger := <-et.ElectionChannel():
			require.EqualValues(t, 1, trigger.Hv.View(), "only view 1's trigger should have been written to chan, as second register should have cancelled the first trigger")
		case <-time.After(1 * time.Second):
			require.Fail(t, "election channel should have contained a single message")
		}

		select {
		case <-et.ElectionChannel():
			require.Fail(t, "election channel should have only contained one message")
		default:
		}
	})
}
