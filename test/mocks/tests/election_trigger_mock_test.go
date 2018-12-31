package tests

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestElectionTriggerMockInitialization(t *testing.T) {
	actual := mocks.NewMockElectionTrigger()
	require.NotNil(t, actual)
}

func TestCallingCallback(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := mocks.NewMockElectionTrigger()
		var actualView primitives.View = 666
		var actualHeight primitives.BlockHeight = 666
		var expectedView primitives.View = 10
		var expectedHeight primitives.BlockHeight = 20
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View) {
			actualHeight = blockHeight
			actualView = view
		}
		et.RegisterOnElection(ctx, expectedHeight, expectedView, cb)

		go et.ManualTrigger()
		trigger := <-et.ElectionChannel()
		trigger(ctx)

		require.Equal(t, expectedView, actualView)
		require.Equal(t, expectedHeight, actualHeight)
	})
}

func TestIgnoreEmptyCallback(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := mocks.NewMockElectionTrigger()

		go et.ManualTrigger()
		trigger := <-et.ElectionChannel()
		trigger(ctx)
	})
}
