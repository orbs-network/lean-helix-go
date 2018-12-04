package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

// Leader election //
func TestViewIncrementedAfterElectionTrigger(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		h.checkView(0)
		h.triggerElection(ctx)
		h.checkView(1)
	})
}

func TestNewViewNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(startView primitives.View, view primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			block := builders.CreateBlock(builders.GenesisBlock)

			h.receiveNewView(ctx, 2, 1, view, block)

			if shouldAcceptMessage {
				h.checkView(view)
			} else {
				h.checkView(startView)
			}
		}

		// notify node2 (view=6, future) as the leader
		sendNewView(5, 6, true)

		// notify node2 (view=2, past) as the leader
		sendNewView(5, 2, false)
	})
}

func TestNewViewNotAcceptMessageIfNotFromTheLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(fromNodeIdx int, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			block := builders.CreateBlock(builders.GenesisBlock)

			h.receiveNewView(ctx, fromNodeIdx, 1, 1, block)
			if shouldAcceptMessage {
				h.checkView(1)
			} else {
				h.checkView(0)
			}
		}

		// getting a new view message from node1 (the new leader)
		sendNewView(1, true)

		// getting a new view message from node2 about node1 as the new leader
		sendNewView(2, false)
	})
}

func TestViewChangeNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendViewChange := func(startView primitives.View, view primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			block := builders.CreateBlock(builders.GenesisBlock)

			viewChangeCountBefore := h.countViewChange(1, view)
			h.receiveViewChange(ctx, 3, 1, view, block)
			viewChangeCountAfter := h.countViewChange(1, view)

			isMessageAccepted := viewChangeCountAfter == viewChangeCountBefore+1
			if shouldAcceptMessage {
				require.True(t, isMessageAccepted)
			} else {
				require.False(t, isMessageAccepted)
			}
		}

		// re-voting me (node0, view=12 -> future) as the leader
		sendViewChange(8, 12, true)

		// re-voting me (node0, view=8 -> present) as the leader
		sendViewChange(8, 8, true)

		// re-voting me (node0, view=4 -> past) as the leader
		sendViewChange(8, 4, false)
	})
}

func TestViewChangeIsRejectedIfTargetIsNotTheNewLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendViewChange := func(startView primitives.View, view primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, view)

			block1 := builders.CreateBlock(builders.GenesisBlock)
			block2 := builders.CreateBlock(block1)

			viewChangeCountBefore := h.countViewChange(1, view)
			h.receiveViewChange(ctx, 3, 1, view, block2)
			viewChangeCountAfter := h.countViewChange(1, view)

			isMessageAccepted := viewChangeCountAfter == viewChangeCountBefore+1
			if shouldAcceptMessage {
				require.True(t, isMessageAccepted)
			} else {
				require.False(t, isMessageAccepted)
			}
		}

		// voting me (node0, view=4) as the leader
		sendViewChange(1, 4, true)

		// voting node2 (view=2) as the leader
		sendViewChange(1, 2, false)
	})
}

func TestPrepareNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendPrepare := func(startView primitives.View, view primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			block := builders.CreateBlock(builders.GenesisBlock)

			prepareCountBefore := h.countPrepare(1, view, block)
			h.receivePrepare(ctx, 1, 1, view, block)
			prepareCountAfter := h.countPrepare(1, view, block)

			isMessageAccepted := prepareCountAfter == prepareCountBefore+1
			if shouldAcceptMessage {
				require.True(t, isMessageAccepted)
			} else {
				require.False(t, isMessageAccepted)
			}
		}

		// sending a valid prepare (On view 12, future)
		sendPrepare(8, 12, true)

		// sending a valid prepare (On view 8, present)
		sendPrepare(8, 8, true)

		// sending a bad prepare (On view 4, past)
		sendPrepare(8, 4, false)
	})
}

func TestPrepareNotAcceptingMessagesFromTheLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendPrepare := func(startView primitives.View, view primitives.View, fromNode int, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, 1)

			block := builders.CreateBlock(builders.GenesisBlock)

			prepareCountBefore := h.countPrepare(1, view, block)
			h.receivePrepare(ctx, fromNode, 1, view, block)
			prepareCountAfter := h.countPrepare(1, view, block)

			isMessageAccepted := prepareCountAfter == prepareCountBefore+1
			if shouldAcceptMessage {
				require.True(t, isMessageAccepted)
			} else {
				require.False(t, isMessageAccepted)
			}

			h.receivePrepare(ctx, 2, 2, 1, block)
			prepareCount := h.countPrepare(2, 1, block)
			require.Equal(t, 1, prepareCount, "Term should not ignore Prepare message from node2")
		}

		// sending a valid prepare (From node2)
		sendPrepare(1, 1, 2, true)

		// sending an invalid prepare (From node1 - the leader)
		sendPrepare(1, 1, 1, false)
	})
}

func TestPreprepareNotAcceptedIfBlockHashDoesNotMatch(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendPreprepare := func(startView primitives.View, block leanhelix.Block, blockHash primitives.Uint256, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			ppm := h.createPreprepareMessage(1, 1, 1, block, blockHash)
			h.receivePreprepareMessage(ctx, ppm)

			hasPreprepare := h.hasPreprepare(1, 1, block)
			if shouldAcceptMessage {
				require.True(t, hasPreprepare, "Term should not ignore the Preprepare message")
			} else {
				require.False(t, hasPreprepare, "Term should ignore the Preprepare message")
			}
		}

		block := builders.CreateBlock(builders.GenesisBlock)

		// sending a valid preprepare
		matchingBlockHash := builders.CalculateBlockHash(block)
		sendPreprepare(1, block, matchingBlockHash, true)

		// sending an invalid preprepare (Mismatching block hash)
		mismatchingBlockHash := builders.CalculateBlockHash(builders.GenesisBlock)
		sendPreprepare(1, block, mismatchingBlockHash, false)
	})
}

func TestNewViewNotAcceptedWithWrongPPView(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(view primitives.View, preprepareView primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			block := builders.CreateBlock(builders.GenesisBlock)

			h.checkView(0)
			h.receiveCustomNewViewMessage(ctx, 1, 1, view, block, preprepareView)
			if shouldAcceptMessage {
				h.checkView(1)
			} else {
				h.checkView(0)
			}
		}

		sendNewView(1, 1, true)
		sendNewView(1, 2, false)
	})
}

func TestPreprepareAcceptOnlyMatchingViews(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendPreprepare := func(startView primitives.View, view primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			block := builders.CreateBlock(builders.GenesisBlock)

			hasPreprepare := h.hasPreprepare(1, startView, block)
			require.False(t, hasPreprepare, "No preprepare should exist in the storage")

			// current view (5) => valid
			h.receivePreprepare(ctx, 1, 1, view, block)
			hasPreprepare = h.hasPreprepare(1, view, block)
			if shouldAcceptMessage {
				require.True(t, hasPreprepare, "Term should not ignore the Preprepare message")
			} else {
				require.False(t, hasPreprepare, "Term should ignore the Preprepare message")
			}
		}

		// current view (5) => valid
		sendPreprepare(5, 5, true)

		// view from the future (9) => invalid, should be ignored
		sendPreprepare(5, 9, false)

		// view from the future (1) => invalid, should be ignored
		sendPreprepare(5, 1, false)
	})
}

func TestPrepare2fPlus1ForACommit(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		h.setNode1AsTheLeader(ctx, 1, 1, block1)

		require.Equal(t, 0, h.countCommits(2, 1, block2), "No commits should exist in the storage")
		h.receivePreprepare(ctx, 1, 2, 1, block2)

		require.Equal(t, 0, h.countCommits(2, 1, block2), "No commits should exist in the storage")
		h.receivePrepare(ctx, 2, 2, 1, block2)

		require.Equal(t, 1, h.countCommits(2, 1, block2), "There should be 1 commit in the storage")
		h.receivePrepare(ctx, 3, 2, 1, block2)

		require.Equal(t, 1, h.countCommits(2, 1, block2), "There should be 1 commit in the storage")
	})
}
