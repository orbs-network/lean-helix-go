package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNodeSync(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.ATestNetwork(4, block1, block2, block3)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]

		net.NodesPauseOnRequestNewBlock()
		net.StartConsensus(ctx)

		// closing node3's network to messages (To make it out of sync)
		node3.Communication.SetIncomingWhitelist([]primitives.MemberId{})

		// node0, node1, and node2 are progressing to block2
		net.WaitForNodeToRequestNewBlock(ctx, node0)
		net.ResumeNodeRequestNewBlock(ctx, node0)
		net.WaitForNodesToCommitASpecificBlock(ctx, block1, node0, node1, node2)

		net.WaitForNodeToRequestNewBlock(ctx, node0)
		net.ResumeNodeRequestNewBlock(ctx, node0)
		net.WaitForNodesToCommitASpecificBlock(ctx, block2, node0, node1, node2)

		// node3 is still "stuck" on the genesis block
		require.True(t, node3.GetLatestBlock() == interfaces.GenesisBlock)

		net.WaitForNodeToRequestNewBlock(ctx, node0)

		// opening node3's network to messages
		node3.Communication.ClearIncomingWhitelist()

		// syncing node3
		latestBlock := node0.GetLatestBlock()
		latestBlockProof := node0.GetLatestBlockProof()
		prevBlockProof := node0.GetBlockProofAt(latestBlock.Height())
		node3.Sync(ctx, latestBlock, latestBlockProof, prevBlockProof)

		net.ResumeNodeRequestNewBlock(ctx, node0)

		// now that node3 is synced, they all should progress to block3
		net.WaitForNodesToCommitASpecificBlock(ctx, block3, node0, node1, node2, node3)
	})
}
