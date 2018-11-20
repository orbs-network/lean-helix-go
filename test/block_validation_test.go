package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCallValidateBlockDuringConsensus(t *testing.T) {
	WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork()

		net.StartConsensus(ctx)
		require.True(t, net.AllNodesValidatedNoMoreThanOnceBeforeCommit())
	})
}
