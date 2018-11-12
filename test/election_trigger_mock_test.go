package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestElectionTriggerMockInitialization(t *testing.T) {
	actual := builders.NewMockElectionTrigger()
	require.NotNil(t, actual)
}

func TestElectionTriggerMockContextCreation(t *testing.T) {
	WithContext(func(ctx context.Context) {
		et := builders.NewMockElectionTrigger()
		resultContext := et.CreateElectionContext(ctx, 10)
		require.NotNil(t, resultContext)
	})
}

func TestElectionTriggerMockTriggerCancellation(t *testing.T) {
	WithContext(func(ctx context.Context) {
		et := builders.NewMockElectionTrigger()
		resultContext := et.CreateElectionContext(ctx, 10)
		et.Trigger()
		require.Error(t, resultContext.Err())
	})
}
