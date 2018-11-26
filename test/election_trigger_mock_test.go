package test

import (
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestElectionTriggerMockInitialization(t *testing.T) {
	actual := builders.NewMockElectionTrigger()
	require.NotNil(t, actual)
}

func TestCallingCallback(t *testing.T) {
	et := builders.NewMockElectionTrigger()
	var actualView primitives.View = 666
	var expectedView primitives.View = 10
	cb := func(view primitives.View) { actualView = view }
	et.RegisterOnElection(expectedView, cb)

	go et.ManualTrigger()
	trigger := <-et.ElectionChannel()
	trigger()

	require.Equal(t, expectedView, actualView)
}

func TestIgnoreEmptyCallback(t *testing.T) {
	et := builders.NewMockElectionTrigger()

	go et.ManualTrigger()
	trigger := <-et.ElectionChannel()
	trigger()
}
