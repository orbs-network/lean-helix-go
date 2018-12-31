package blockvalidation

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCallValidateBlockDuringConsensus(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()

		net.StartConsensus(ctx)
		require.True(t, net.AllNodesValidatedNoMoreThanOnceBeforeCommit())
	})
}
