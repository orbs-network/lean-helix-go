package electiontrigger

import (
	"github.com/orbs-network/lean-helix-go/go/electiontrigger"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCallbackTrigger(t *testing.T) {
	et := electiontrigger.NewTimerBasedElectionTrigger(10)
	wasCalled := false
	cb := func() { wasCalled = true }
	et.Start(cb)
	time.Sleep(time.Duration(15) * time.Millisecond)
	et.Stop()

	require.True(t, wasCalled, "Did not call the timer callback")
}

func TestCallbackTriggerTwice(t *testing.T) {
	et := electiontrigger.NewTimerBasedElectionTrigger(10)
	callCount := 0
	cb := func() { callCount++ }
	et.Start(cb)
	time.Sleep(time.Duration(25) * time.Millisecond)
	et.Stop()

	require.Exactly(t, 2, callCount, "Did not tigger the timer twice")
}

func TestStoppingTrigger(t *testing.T) {
	et := electiontrigger.NewTimerBasedElectionTrigger(10)
	callCount := 0
	cb := func() { callCount++ }
	et.Start(cb)
	time.Sleep(time.Duration(15) * time.Millisecond)
	et.Stop()
	time.Sleep(time.Duration(20) * time.Millisecond)

	require.Exactly(t, 1, callCount, "Did not stop the timer")
}
