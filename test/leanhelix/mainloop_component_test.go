package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/leaderelection"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const LOG_TO_CONSOLE = false

func TestMainloopReportsCorrectHeight(t *testing.T) {
	test.WithContext(func(ctx context.Context) {

		net := network.ABasicTestNetwork(ctx)
		node0 := net.Nodes[0]

		net.SetNodesToPauseOnRequestNewBlock()
		net.StartConsensus(ctx)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 2)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block2
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 3)
	})
}

func TestVerifyPreprepareMessageSentByLeader_HappyFlow(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		nodeCount := 4
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(nodeCount).
			WithBlocks(block1, block2).
			//LogToConsole(t).
			Build(ctx)

		node0 := net.Nodes[0]
		net.SetNodesToPauseOnRequestNewBlock()
		net.StartConsensus(ctx)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1, should be agreed by all nodes
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 2)
		require.Equal(t, nodeCount-1, node0.Communication.CountMessagesSent(protocol.LEAN_HELIX_PREPREPARE, mocks.BLOCK_HEIGHT_DONT_CARE, mocks.VIEW_DONT_CARE, nil), "node0 should have sent %d PREPREPARE messages", nodeCount-1)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block2, should be agreed by all nodes
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 3)
		require.Equal(t, (nodeCount-1)*2, node0.Communication.CountMessagesSent(protocol.LEAN_HELIX_PREPREPARE, mocks.BLOCK_HEIGHT_DONT_CARE, mocks.VIEW_DONT_CARE, nil), "node0 should have sent total of %d PREPREPARE messages", (nodeCount-1)*2)
	})
}

func TestPreprepareMessageNotSentByLeaderIfRequestNewBlockProposalContextCancelled(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		nodeCount := 4
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(nodeCount).
			WithBlocks(block1, block2, block3).
			LogToConsole(t).
			Build(ctx)

		bc, err := leaderelection.GenerateBlocksWithProofsForTest([]interfaces.Block{block1, block2, block3}, net.Nodes)
		if err != nil {
			t.Fatalf("Error creating mock blockchain for tests - %s", err)
			return
		}
		node0 := net.Nodes[0]
		consensusRoundChan := make(chan primitives.BlockHeight, 10)

		//net.SetNodesPauseOnRequestNewBlockWhenCounterIsZero(2)
		net.SetNodesToPauseOnRequestNewBlock()
		net.StartConsensus(ctx)
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1, should be agreed by all nodes
		net.ResumeRequestNewBlockOnNodes(ctx, node0)

		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 2)
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)

		require.Equal(t, nodeCount-1, node0.Communication.CountMessagesSent(protocol.LEAN_HELIX_PREPREPARE, mocks.BLOCK_HEIGHT_DONT_CARE, mocks.VIEW_DONT_CARE, nil), "node0 sent PREPREPARE despite having its worker context cancelled during RequestNewBlockProposal")

		blockToSync, blockProofToSync := bc.BlockAndProofAt(2)
		prevBlockProofToSync := bc.BlockProofAt(1)

		require.Equal(t, blockToSync.Height(), node0.GetCurrentHeight())
		node0.SetPauseOnNewConsensusRoundUntilReadingFrom(consensusRoundChan)
		for _, node := range net.Nodes {
			if err := node.Sync(ctx, blockToSync, blockProofToSync, prevBlockProofToSync); err != nil { // block2 has H=2 so next block is H=3
				t.Fatalf("Sync failed for node %s - %s", node.MemberId, err)
			}
		}

		expectedHeightOfNewTermAfterSuccessfulSync := blockToSync.Height() + 1
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, expectedHeightOfNewTermAfterSuccessfulSync, node0)
		ppmSent := node0.Communication.CountMessagesSent(protocol.LEAN_HELIX_PREPREPARE, mocks.BLOCK_HEIGHT_DONT_CARE, mocks.VIEW_DONT_CARE, nil)
		require.Equal(t, nodeCount-1, ppmSent, "node0 sent PREPREPARE despite having its worker context cancelled by UpdateState during RequestNewBlockProposal")
	})
}

func TestVerifyWorkerContextNotCancelledIfNodeSyncBlockIsIgnored(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks(block1, block2, block3).
			//LogToConsole(t).
			Build(ctx)

		node0 := net.Nodes[0]
		net.SetNodesToPauseOnRequestNewBlock(node0)
		net.StartConsensus(ctx)
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1, should be agreed by all nodes
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitUntilNodesEventuallyCommitASpecificBlock(ctx, t, 0, block1)
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // pause when proposing block2
		bc, err := leaderelection.GenerateBlocksWithProofsForTest([]interfaces.Block{block1, block2, block3}, net.Nodes)
		if err != nil {
			t.Fatalf("Error creating mock blockchain for tests - %s", err)
			return
		}

		blockToSync, blockProofToSync := bc.BlockAndProofAt(1)
		prevBlockProofToSync := bc.BlockProofAt(0)

		if err := node0.Sync(ctx, blockToSync, blockProofToSync, prevBlockProofToSync); err != nil {
			t.Fatalf("Sync failed for node %s - %s", node0.MemberId, err)
		}

		time.Sleep(100 * time.Millisecond) // let the above go func run

		require.Equal(t, primitives.BlockHeight(2), node0.GetCurrentHeight())
	})
}
