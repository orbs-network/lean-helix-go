package leanhelixterm

import (
	"context"
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
		h := NewHarness(ctx, t)

		h.checkView(0)
		h.triggerElection(ctx)
		h.checkView(1)

		// voting node0 as the leader
		block := builders.CreateBlock(builders.GenesisBlock)
		h.setMeAsTheLeader(ctx, 1, 8, block)
		h.checkView(8)

		// re-voting node0 as the leader, but with a view from the past (4)
		block = builders.CreateBlock(builders.GenesisBlock)
		h.setMeAsTheLeader(ctx, 1, 4, block)
		h.checkView(8) // unchanged
	})
}

func TestNewViewNotAcceptMessageIfNotFromTheLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		// node0 (me) is the leader
		// getting a new view message from node2 about node1 as the new leader
		block := builders.CreateBlock(builders.GenesisBlock)
		h.checkView(0)
		h.sendNewView(ctx, 2, 1, 1, block)
		h.checkView(0)
	})
}

func TestViewChangeNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		// jumping to view=8 me (node0) as the leader
		h.checkView(0)
		block := builders.CreateBlock(builders.GenesisBlock)
		h.setMeAsTheLeader(ctx, 1, 8, block)
		h.checkView(8)

		// re-voting me (node0, view=12 -> future) as the leader
		block = builders.CreateBlock(builders.GenesisBlock)
		h.sendViewChange(ctx, 3, 1, 12, block)
		viewChangeCount := h.countViewChange(1, 12)
		require.Equal(t, 1, viewChangeCount, "Term should not ignore ViewChange message on view 12")

		// re-voting me (node0, view=4 -> past) as the leader
		block = builders.CreateBlock(builders.GenesisBlock)
		h.sendViewChange(ctx, 3, 1, 4, block)
		viewChangeCount = h.countViewChange(1, 4)
		require.Equal(t, 0, viewChangeCount, "Term should not ignore ViewChange message on view 4 (From the past)")
	})
}

func TestViewChangeIsRejectedIfTargetIsNotTheNewLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		// setting node1 as the leader
		h.checkView(0)
		h.triggerElection(ctx)
		h.checkView(1)

		// voting me (node0, view=4) as the leader
		h.sendViewChange(ctx, 3, 1, 4, block2)
		viewChangeCount := h.countViewChange(1, 4)
		require.Equal(t, 1, viewChangeCount, "Term should not ignore ViewChange message on view 4")

		// voting node2 (view=2) as the leader
		h.sendViewChange(ctx, 3, 1, 2, block2)
		viewChangeCount = h.countViewChange(1, 2)
		require.Equal(t, 0, viewChangeCount, "Term should not ignore ViewChange message on view 4")
	})
}

func TestPrepareNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block := builders.CreateBlock(builders.GenesisBlock)

		// jumping to view=8 me (node0) as the leader
		h.setMeAsTheLeader(ctx, 1, 8, block)

		// sending a valid prepare (On view 12)
		h.sendPrepare(ctx, 1, 1, 12, block)
		prepareCount := h.countPrepare(1, 12, block)
		require.Equal(t, 1, prepareCount, "Term should not ignore Prepare message on view 8 (Current view)")

		// sending a bad prepare (On view 4, from the past)
		h.sendPrepare(ctx, 2, 1, 4, block)
		prepareCount = h.countPrepare(1, 4, block)
		require.Equal(t, 0, prepareCount, "Term should ignore Prepare message on view 4 (From the past)")
	})
}

func TestPrepareNotAcceptingMessagesFromTheLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block := builders.CreateBlock(builders.GenesisBlock)

		h.setNode1AsTheLeader(ctx, 1, 1, block)

		// sending a valid prepare (From node2)
		h.sendPrepare(ctx, 2, 2, 1, block)
		prepareCount := h.countPrepare(2, 1, block)
		require.Equal(t, 1, prepareCount, "Term should not ignore Prepare message from node2")

		// sending an invalid prepare (From node1 - the leader)
		h.sendPrepare(ctx, 1, 2, 1, block)
		prepareCount = h.countPrepare(2, 1, block)
		require.Equal(t, 1, prepareCount, "Term should ignore Prepare message from node1, the leader")
	})
}

func TestPreprepareNotAcceptedIfBlockHashDoesNotMatch(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block := builders.CreateBlock(builders.GenesisBlock)

		h.setNode1AsTheLeader(ctx, 1, 1, block)

		// sending a valid preprepare
		h.sendPreprepare(ctx, 1, 2, 1, block)
		hasPreprepare := h.hasPreprepare(2, 1, block)
		require.True(t, hasPreprepare, "Term should not ignore Prepare message from node2")

		// sending an invalid preprepare (Mismatching block hash)
		mismatchingBlockHash := builders.CalculateBlockHash(builders.GenesisBlock)
		h.sendPreprepareWithSpecificBlockHash(ctx, 1, 3, 1, block, mismatchingBlockHash)
		hasPreprepare = h.hasPreprepare(3, 1, block)
		require.False(t, hasPreprepare, "Term should ignore Prepare message from node1, the leader")
	})
}

func TestPreprepareAcceptOnlyMatchingViews(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		// set node 1 as the leader (view 5)
		h.checkView(0)
		h.triggerElection(ctx)
		h.triggerElection(ctx)
		h.triggerElection(ctx)
		h.triggerElection(ctx)
		h.triggerElection(ctx)
		h.checkView(5)

		hasPreprepare := h.hasPreprepare(1, 5, block2)
		require.False(t, hasPreprepare, "No preprepare should exist in the storage")

		// current view (5) => valid
		h.sendPreprepare(ctx, 1, 1, 5, block2)
		hasPreprepare = h.hasPreprepare(1, 5, block2)
		require.True(t, hasPreprepare, "A preprepare should exist in the storage")

		// view from the future (9) => invalid, should be ignored
		h.sendPreprepare(ctx, 1, 1, 9, block2)
		hasPreprepare = h.hasPreprepare(1, 9, block2)
		require.False(t, hasPreprepare, "No preprepare should exist in the storage")

		// view from the future (1) => invalid, should be ignored
		h.sendPreprepare(ctx, 1, 1, 1, block2)
		hasPreprepare = h.hasPreprepare(1, 1, block2)
		require.False(t, hasPreprepare, "No preprepare should exist in the storage")
	})
}

func TestPrepare2fPlus1ForACommit(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		h.setNode1AsTheLeader(ctx, 1, 1, block1)

		commitChangeCount := h.countCommits(2, 1, block2)
		require.Equal(t, 0, commitChangeCount, "No commits should exist in the storage")

		h.sendPreprepare(ctx, 1, 2, 1, block2)
		h.sendPrepare(ctx, 2, 2, 1, block2)
		h.sendPrepare(ctx, 3, 2, 1, block2)

		commitChangeCount = h.countCommits(2, 1, block2)
		require.Equal(t, 1, commitChangeCount, "There should be 1 commit in the storage")
	})
}
