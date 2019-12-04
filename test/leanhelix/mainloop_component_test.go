package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/leaderelection"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

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
			//LogToConsole(t).
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
		_, prevBlockProofToSync := bc.BlockAndProofAt(1)

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
		_, prevBlockProofToSync := bc.BlockAndProofAt(0)

		if err := node0.Sync(ctx, blockToSync, blockProofToSync, prevBlockProofToSync); err != nil {
			t.Fatalf("Sync failed for node %s - %s", node0.MemberId, err)
		}

		time.Sleep(100 * time.Millisecond) // let the above go func run

		require.Equal(t, primitives.BlockHeight(2), node0.GetCurrentHeight())
	})
}

// see https://github.com/orbs-network/lean-helix-go/issues/74
func TestViewChangeRaceWithElectionLeader(t *testing.T) {

	test.WithContext(func(ctx context.Context) {
		l := logger.NewConsoleLogger(test.NameHashPrefix(t, 4))

		d := newDriver(l, 1, 4, nil, nil)
		d.start(ctx, t)

		// receive VIEW_CHANGE messages form other committee members
		nextView := state.NewHeightView(1, 1)
		d.handleViewChangeMessage(ctx, nextView, 0)
		d.handleViewChangeMessage(ctx, nextView, 2)

		// for another flavor of this test uncomment this:
		//d.handleViewChangeMessage(ctx, nextView, 3)

		// trigger elections
		d.electionTriggerMock.ManualTrigger(ctx, state.NewHeightView(1, 0))

		require.True(t, test.Eventually(1*time.Second, func() bool {
			newViewSentCount := d.communication.CountSentMessages(protocol.LEAN_HELIX_NEW_VIEW)
			return newViewSentCount == 1
		}), "expect to send NEW_VIEW after at least 2 VIEW_CHANGEs and an election trigger")
	})
}

func TestCommitCallbackErrorDetectedAndPreservesState(t *testing.T) {

	test.WithContext(func(ctx context.Context) {
		l := logger.NewConsoleLogger(test.NameHashPrefix(t, 4))

		onCommitCalledOnce := make(chan struct{})

		d := newDriver(l, 0, 4, func(ctx context.Context, block interfaces.Block, blockProof []byte, view primitives.View) error {
			close(onCommitCalledOnce)
			return fmt.Errorf("intentionally failing commit callback")
		}, nil)
		d.start(ctx, t)

		preprepareBlock := d.waitForSentPreprepareMessage(t, 1).Block()

		d.handlePrepareMessage(ctx, d.leadersByView[1], primitives.BlockHeight(1), primitives.View(0), preprepareBlock)
		d.handlePrepareMessage(ctx, d.leadersByView[2], primitives.BlockHeight(1), primitives.View(0), preprepareBlock)

		d.waitForSentCommitMessage(t, 1)

		genesisRandomSeed := calcGenesisBlockRandomSeed()

		d.handleCommitMessage(ctx, d.leadersByView[1], primitives.BlockHeight(1), primitives.View(0), preprepareBlock, genesisRandomSeed)
		d.handleCommitMessage(ctx, d.leadersByView[2], primitives.BlockHeight(1), primitives.View(0), preprepareBlock, genesisRandomSeed)

		requireChanClosedWithinTimeout(t, ctx, onCommitCalledOnce)

		require.True(t, test.Consistently(100*time.Millisecond, func() bool {
			return d.mainLoop.State().Height() < 2
		}), "expected onNewConsensusRound to not be called due to block commit failure")
	})
}

func TestPreparedNodeCommitsInOlderViewAfterElectionTrigger(t *testing.T) {

	test.WithContext(func(ctx context.Context) {
		l := logger.NewConsoleLogger(test.NameHashPrefix(t, 4))

		const blockHeight = 1
		const view = 0

		commitCallbackCalledChan := make(chan interface{})

		d := newDriver(l, 3, 4, func(ctx context.Context, block interfaces.Block, blockProof []byte, committedAtView primitives.View) error {
			require.Equal(t, primitives.View(view), committedAtView, "expected block to commit at view %d", view)
			close(commitCallbackCalledChan)
			return nil
		}, nil)
		d.start(ctx, t)

		block := mocks.ABlock(interfaces.GenesisBlock)
		randomSeed := calcGenesisBlockRandomSeed()
		leaderMemberId := d.leadersByView[0]
		d.handlePreprepareMessage(ctx, leaderMemberId, blockHeight, view, block, randomSeed)

		d.handlePrepareMessage(ctx, d.leadersByView[1], primitives.BlockHeight(1), primitives.View(0), block)

		d.waitForSentCommitMessage(t, 1)

		// Advance to next view

		d.electionTriggerMock.ManualTrigger(ctx, state.NewHeightView(blockHeight, view))
		require.True(t, test.Eventually(100*time.Millisecond, func() bool {
			return d.mainLoop.State().View() == view+1
		}), "expected node to advance to view %d", view+1)

		// Send commits in previous view

		require.EqualValues(t, d.mainLoop.State().Height(), blockHeight, "expected height to remain  %d", blockHeight)

		d.handleCommitMessage(ctx, d.leadersByView[0], primitives.BlockHeight(1), primitives.View(0), block, randomSeed)
		d.handleCommitMessage(ctx, d.leadersByView[1], primitives.BlockHeight(1), primitives.View(0), block, randomSeed)

		require.True(t, test.Eventually(100*time.Millisecond, func() bool {
			return d.mainLoop.State().Height() == blockHeight+1
		}), "expected node to commit block at height %d", blockHeight)
	})
}

func TestNodePassesCorrectViewToOnCommitCallback(t *testing.T) {

	test.WithContext(func(ctx context.Context) {
		l := logger.NewConsoleLogger(test.NameHashPrefix(t, 4))

		const blockHeight = 1
		const view = 1

		commitCallbackCalledChan := make(chan interface{})

		d := newDriver(l, 3, 4, func(ctx context.Context, block interfaces.Block, blockProof []byte, committedAtView primitives.View) error {
			require.Equal(t, primitives.View(view), committedAtView, "expected block to commit at view %d", view)
			close(commitCallbackCalledChan)
			return nil
		}, nil)
		d.start(ctx, t)

		// Advance to next view

		d.electionTriggerMock.ManualTrigger(ctx, state.NewHeightView(blockHeight, view-1))
		require.True(t, test.Eventually(100*time.Millisecond, func() bool {
			return d.mainLoop.State().View() == view
		}), "expected node to advance to view %d", view)

		block := mocks.ABlock(interfaces.GenesisBlock)
		randomSeed := calcGenesisBlockRandomSeed()
		leaderMemberId := d.leadersByView[view]
		otherMemberId := d.leadersByView[view+1]

		// Complete consensus round
		d.handlePreprepareMessage(ctx, leaderMemberId, blockHeight, view, block, randomSeed)
		d.handlePrepareMessage(ctx, otherMemberId, primitives.BlockHeight(blockHeight), primitives.View(view), block)
		d.handleCommitMessage(ctx, leaderMemberId, primitives.BlockHeight(blockHeight), primitives.View(view), block, randomSeed)
		d.handleCommitMessage(ctx, otherMemberId, primitives.BlockHeight(blockHeight), primitives.View(view), block, randomSeed)

		require.True(t, test.Eventually(100*time.Millisecond, func() bool {
			return d.mainLoop.State().Height() == blockHeight+1
		}), "expected node to commit block at height %d", blockHeight)

		<-commitCallbackCalledChan
	})
}

func TestNodeReportsCorrectViewToOnNewViewCallback(t *testing.T) {

	test.WithContext(func(ctx context.Context) {
		l := logger.NewConsoleLogger(test.NameHashPrefix(t, 4))

		const blockHeight = 1
		const numNodes = 4

		reportedViews := make(chan primitives.View, 100)
		reportedLeaders := make(chan primitives.MemberId, 100)

		d := newDriver(l, 3, numNodes, nil, func(leader primitives.MemberId, newView primitives.View) {
			reportedViews <- newView
			reportedLeaders <- leader
		})
		d.start(ctx, t)

		require.Equal(t, <-reportedViews, primitives.View(0), "expected initial view to be reported")
		require.Equal(t, <-reportedLeaders, d.leadersByView[0], "expected initial leader to be reported")

		// Next view by election trigger
		const viewByTrigger = 1

		d.electionTriggerMock.ManualTrigger(ctx, state.NewHeightView(blockHeight, viewByTrigger-1))
		require.Equal(t, <-reportedViews, primitives.View(viewByTrigger), "expected triggered view to be reported")
		require.Equal(t, <-reportedLeaders, d.leadersByView[viewByTrigger], "expected leader of triggered view to be reported")

		// Next view by new-view message
		const viewByNewView = 2

		confirmations := []*interfaces.ViewChangeMessage{}
		for i := 0; i < numNodes; i++ {
			confirmations = append(confirmations, builders.AViewChangeMessage(d.instanceId, mocks.NewMockKeyManager(d.leadersByView[i]), d.leadersByView[i], blockHeight, viewByNewView, nil))
		}

		block := mocks.ABlock(interfaces.GenesisBlock)
		randomSeed := calcGenesisBlockRandomSeed()

		d.handleNewViewMessage(ctx, d.leadersByView[viewByNewView], blockHeight, viewByNewView, confirmations, block, randomSeed)

		require.Equal(t, <-reportedViews, primitives.View(viewByNewView), "expected new view to be reported")
		require.Equal(t, <-reportedLeaders, d.leadersByView[viewByNewView], "expected leader of new view to be reported")
	})
}

func TestUnpreparedNodeDoesNotSendCommitsInOlderViewAfterElectionTrigger(t *testing.T) {

	test.WithContext(func(ctx context.Context) {
		l := logger.NewConsoleLogger(test.NameHashPrefix(t, 4))

		d := newDriver(l, 3, 4, func(ctx context.Context, block interfaces.Block, blockProof []byte, view primitives.View) error {
			return nil
		}, nil)
		d.start(ctx, t)

		const blockHeight = 1
		const view = 0

		block := mocks.ABlock(interfaces.GenesisBlock)
		randomSeed := calcGenesisBlockRandomSeed()
		leaderMemberId := d.leadersByView[0]
		d.handlePreprepareMessage(ctx, leaderMemberId, blockHeight, view, block, randomSeed)

		// Advance to next view before node is prepared

		d.electionTriggerMock.ManualTrigger(ctx, state.NewHeightView(blockHeight, view))
		require.True(t, test.Eventually(100*time.Millisecond, func() bool {
			return d.mainLoop.State().View() == view+1
		}), "expected node to advance to view %d", view+1)

		// Send prepares and commits in previous view
		d.handlePrepareMessage(ctx, d.leadersByView[1], primitives.BlockHeight(1), primitives.View(0), block)

		d.handleCommitMessage(ctx, d.leadersByView[0], primitives.BlockHeight(1), primitives.View(0), block, randomSeed)
		d.handleCommitMessage(ctx, d.leadersByView[1], primitives.BlockHeight(1), primitives.View(0), block, randomSeed)

		// Node should not send commit messages or commit the block

		require.True(t, test.Consistently(100*time.Millisecond, func() bool {
			return len(d.communication.GetSentMessages(protocol.LEAN_HELIX_COMMIT)) == 0 &&
				d.mainLoop.State().Height() == blockHeight
		}), "expected node to never send any commit messages")

		require.EqualValues(t, d.mainLoop.State().Height(), blockHeight, "expected height to remain  %d", blockHeight)
	})
}

func calcGenesisBlockRandomSeed() uint64 {
	prevBlockProof := protocol.BlockProofReader(nil) // nil represents the block proof of the genesis block
	genesisRandomSeed := randomseed.CalculateRandomSeed(prevBlockProof.RandomSeedSignature())
	return genesisRandomSeed
}

func requireChanClosedWithinTimeout(t *testing.T, ctx context.Context, onCommitCalledOnce chan struct{}) {
	timeout, c := context.WithTimeout(ctx, 1*time.Second)
	defer c()
	select {
	case <-onCommitCalledOnce:
	case <-timeout.Done():
		t.FailNow()
	}
}
