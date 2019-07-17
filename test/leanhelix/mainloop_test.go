package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVerifyWorkerContextNotCancelledIfNodeSyncBlockIsIgnored(t *testing.T) {
	t.Skip()
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]interfaces.Block{block1, block2}).
			LogToConsole().
			Build(ctx)

		node0 := net.Nodes[0]

		// the leader (node0) is suggesting block1 to node1 and node2 (not to node3)
		net.StartConsensus(ctx)

		// node0, node1 and node2 should reach consensus
		require.True(t, net.WaitForNodesToCommitASpecificBlock(ctx, block1))
		fmt.Print("----- done block1\n")
		net.SetNodesToPauseOnRequestNewBlock()
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // pause when proposing block2
		require.True(t, net.WaitForNodesToCommitASpecificBlock(ctx, block2))
		fmt.Print("----- done block2\n")

		//time.Sleep(1 * time.Second)

	})
}
