package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const LOG_TO_CONSOLE = true

func TestVerifyWorkerContextNotCancelledIfNodeSyncBlockIsIgnored(t *testing.T) {
	t.Skip()
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks(block1, block2, block3).
			LogToConsole().
			Build(ctx)

		node0 := net.Nodes[0]
		net.SetNodesToPauseOnRequestNewBlock(node0)
		net.StartConsensus(ctx)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1, should be agreed by all nodes
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		require.True(t, net.WaitForAllNodesToCommitBlockAndReturnWhetherEqualToGiven(ctx, block1))
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // pause when proposing block2
		node0.Sync(ctx, block1, nil, nil)
		go func(ctx context.Context) {
			// Check that request new block was cancelled - this should NOT happen

			t.Fatal("RequestNewBlockProposal was cancelled although an old and irrelevant Block was provided to Node Sync")
		}(ctx)
		net.ResumeRequestNewBlockOnNodes(ctx, node0)

		time.Sleep(100 * time.Millisecond)
		require.True(t, net.WaitForAllNodesToCommitBlockAndReturnWhetherEqualToGiven(ctx, block2))
	})
}
