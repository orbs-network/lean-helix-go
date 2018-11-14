package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHappyFlow(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork(ctx)
		net.StartConsensus(ctx)
		require.True(t, net.InConsensus())
	})
}

func TestOnlyLeaderIsSendingPrePrepareOnce(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork(ctx)

		waitUntilAllNodesCalledValidate := net.PauseNodesExecutionOnValidation(net.Nodes[1], net.Nodes[2], net.Nodes[3])
		net.StartConsensus(ctx)
		waitUntilAllNodesCalledValidate()

		require.Equal(t, 1, net.Nodes[0].Gossip.CountSentMessages(leanhelix.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[1].Gossip.CountSentMessages(leanhelix.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[2].Gossip.CountSentMessages(leanhelix.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[3].Gossip.CountSentMessages(leanhelix.LEAN_HELIX_PREPREPARE))
	})
}

func TestConsensusFor8Blocks(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork(ctx).StartConsensus(ctx)
		for i := 0; i < 8; i++ {
			net.InConsensus()
		}
	})
}
