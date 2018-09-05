package electiontrigger

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCallbackTrigger(t *testing.T) {
	et := lh.NewTimerBasedElectionTrigger(10)
	wasCalled := false
	cb := func(view lh.ViewCounter) { wasCalled = true }
	et.RegisterOnTrigger(0, cb)
	time.Sleep(time.Duration(15) * time.Millisecond)

	require.True(t, wasCalled, "Did not call the timer callback")
	et.UnregisterOnTrigger()
}

func TestCallbackTriggerOnce(t *testing.T) {
	et := lh.NewTimerBasedElectionTrigger(10)
	callCount := 0
	cb := func(view lh.ViewCounter) { callCount++ }
	et.RegisterOnTrigger(0, cb)
	time.Sleep(time.Duration(25) * time.Millisecond)

	require.Exactly(t, 1, callCount, "Trigger callback called more than once")
	et.UnregisterOnTrigger()
}

func TestIgnoreSameView(t *testing.T) {
	et := lh.NewTimerBasedElectionTrigger(30)
	callCount := 0
	cb := func(view lh.ViewCounter) { callCount++ }

	et.RegisterOnTrigger(0, cb)
	time.Sleep(time.Duration(10) * time.Millisecond)
	et.RegisterOnTrigger(0, cb)
	time.Sleep(time.Duration(10) * time.Millisecond)
	et.RegisterOnTrigger(0, cb)
	time.Sleep(time.Duration(20) * time.Millisecond)
	et.RegisterOnTrigger(0, cb)

	require.Exactly(t, 1, callCount, "Trigger callback called more than once")
	et.UnregisterOnTrigger()
}

func TestViewChanges(t *testing.T) {
	et := lh.NewTimerBasedElectionTrigger(20)
	wasCalled := false
	cb := func(view lh.ViewCounter) { wasCalled = true }

	et.RegisterOnTrigger(0, cb) // 2 ** 0 * 20 = 20
	time.Sleep(time.Duration(10) * time.Millisecond)

	et.RegisterOnTrigger(1, cb) // 2 ** 1 * 20 = 40
	time.Sleep(time.Duration(30) * time.Millisecond)

	et.RegisterOnTrigger(2, cb) // 2 ** 2 * 20 = 80
	time.Sleep(time.Duration(70) * time.Millisecond)

	et.RegisterOnTrigger(3, cb) // 2 ** 3 * 20 = 160

	require.False(t, wasCalled, "Trigger the callback even if a new Register was called with a new view")
	et.UnregisterOnTrigger()
}

func TestViewPowTimeout(t *testing.T) {
	et := lh.NewTimerBasedElectionTrigger(10)
	wasCalled := false
	cb := func(view lh.ViewCounter) { wasCalled = true }

	et.RegisterOnTrigger(2, cb) // 2 ** 2 * 10 = 40
	time.Sleep(time.Duration(35) * time.Millisecond)
	require.False(t, wasCalled, "Triggered the callback too early")
	time.Sleep(time.Duration(10) * time.Millisecond)
	require.True(t, wasCalled, "Did not trigger the callback after the required timeout")

	et.UnregisterOnTrigger()
}

func TestStoppingTrigger(t *testing.T) {
	et := lh.NewTimerBasedElectionTrigger(10)
	wasCalled := false
	cb := func(view lh.ViewCounter) { wasCalled = true }
	et.RegisterOnTrigger(0, cb)
	time.Sleep(time.Duration(5) * time.Millisecond)
	et.UnregisterOnTrigger()
	time.Sleep(time.Duration(15) * time.Millisecond)

	require.False(t, wasCalled, "Did not stop the timer")
	et.UnregisterOnTrigger()
}
