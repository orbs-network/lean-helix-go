package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const LOG_TO_CONSOLE = true

// TODO Add state object to mainloop and workerloop so the correct currentheight will be reported on both loops
func TestMainloopReportsCorrectHeight(t *testing.T) {
	t.Skip("Unskip when State object makes mainlop.currentHeight return correct result")
	test.WithContext(func(ctx context.Context) {
		nodeCount := 4
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(nodeCount).
			WithBlocks(block1, block2).
			LogToConsole().
			Build(ctx)

		node0 := net.Nodes[0]
		net.SetNodesToPauseOnRequestNewBlock()
		net.StartConsensus(ctx)
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitForNodesToCommitABlock(ctx)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitForNodesToCommitABlock(ctx)

		//net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block2
		require.Equal(t, block2.Height()+1, node0.GetCurrentHeight(), "node0 should be on height 1")

	})
}

func TestVerifyPreprepareMessageSentByLeader_HappyFlow(t *testing.T) {
	t.Skip("Unskip when State object makes mainlop.currentHeight return correct result")
	test.WithContext(func(ctx context.Context) {
		nodeCount := 4
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(nodeCount).
			WithBlocks(block1, block2).
			LogToConsole().
			Build(ctx)

		node0 := net.Nodes[0]
		net.StartConsensus(ctx)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1, should be agreed by all nodes
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitForNodesToCommitASpecificBlock(ctx, block1)
		require.Equal(t, nodeCount-1, node0.Communication.CountMessagesSent(protocol.LEAN_HELIX_PREPREPARE, mocks.BLOCK_HEIGHT_DONT_CARE, mocks.VIEW_DONT_CARE, nil), "node0 should have sent %d PREPREPARE messages", nodeCount-1)

		//net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block2, should be agreed by all nodes
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitForNodesToCommitASpecificBlock(ctx, block2)
		require.Equal(t, (nodeCount-1)*2, node0.Communication.CountMessagesSent(protocol.LEAN_HELIX_PREPREPARE, mocks.BLOCK_HEIGHT_DONT_CARE, mocks.VIEW_DONT_CARE, nil), "node0 should have sent total of %d PREPREPARE messages", (nodeCount-1)*2)
	})
}

func TestPreprepareMessageNotSentByLeaderIfRequestNewBlockProposalContextCancelled(t *testing.T) {
	t.Skip("Unskip when State object makes mainlop.currentHeight return correct result")
	test.WithContext(func(ctx context.Context) {
		nodeCount := 4
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(nodeCount).
			WithBlocks(block1, block2, block3).
			LogToConsole().
			Build(ctx)

		node0 := net.Nodes[0]
		//net.SetNodesPauseCounterOnRequestNewBlock(2)
		net.SetNodesToPauseOnRequestNewBlock()
		net.StartConsensus(ctx)
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1, should be agreed by all nodes
		net.ResumeRequestNewBlockOnNodes(ctx, node0)

		require.True(t, net.WaitForNodesToCommitASpecificBlock(ctx, block1))
		require.Equal(t, nodeCount-1, node0.Communication.CountMessagesSent(protocol.LEAN_HELIX_PREPREPARE, mocks.BLOCK_HEIGHT_DONT_CARE, mocks.VIEW_DONT_CARE, nil), "node0 sent PREPREPARE despite having its worker context cancelled during RequestNewBlockProposal")

		latestBlock := node0.GetLatestBlock() // block1
		latestBlockProof := node0.GetLatestBlockProof()
		prevBlockProof := node0.GetBlockProofAt(latestBlock.Height())
		fmt.Printf("Node0 is on H=%d\n", latestBlock.Height())

		for _, node := range net.Nodes {
			node.Sync(ctx, latestBlock, latestBlockProof, prevBlockProof) // block2 has H=2 so next block is H=3
		}
		// Sync closed the context of previous Pause of RequestNewBlock so now it's time to pause on it again
		net.WaitForNodesToCommitABlock(ctx)
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block3

		// TODO This line will work only when State is implemented
		//require.Equal(t, latestBlock.Height()+1, node0.GetCurrentHeight(), "node0 should be on height %d", latestBlock.Height()+1)

		//net.ResumeRequestNewBlockOnNodes(ctx, node0)

		// Only 2 block are closed with PREPREPARE - one was provided with sync
		require.Equal(t, (nodeCount-1)*2, node0.Communication.CountMessagesSent(protocol.LEAN_HELIX_PREPREPARE, mocks.BLOCK_HEIGHT_DONT_CARE, mocks.VIEW_DONT_CARE, nil), "node0 sent PREPREPARE despite having its worker context cancelled during RequestNewBlockProposal")
	})
}

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
