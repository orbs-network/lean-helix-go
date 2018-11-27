package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/orbs-network-go/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestElectionTriggerMockInitialization(t *testing.T) {
	actual := builders.NewMockElectionTrigger()
	require.NotNil(t, actual)
}

func TestCallingCallback(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := builders.NewMockElectionTrigger()
		var actualView primitives.View = 666
		var expectedView primitives.View = 10
		cb := func(ctx context.Context, view primitives.View) { actualView = view }
		et.RegisterOnElection(expectedView, cb)

		go et.ManualTrigger()
		trigger := <-et.ElectionChannel()
		trigger(ctx)

		require.Equal(t, expectedView, actualView)
	})
}

func TestIgnoreEmptyCallback(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := builders.NewMockElectionTrigger()

		go et.ManualTrigger()
		trigger := <-et.ElectionChannel()
		trigger(ctx)
	})
}
