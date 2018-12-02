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
		h.sendViewChange(ctx, 1, 12, block)
		viewChangeCount := h.countViewChange(1, 12)
		require.Equal(t, 1, viewChangeCount, "Term should not ignore ViewChange message on view 12")

		// re-voting me (node0, view=4 -> past) as the leader
		block = builders.CreateBlock(builders.GenesisBlock)
		h.sendViewChange(ctx, 1, 4, block)
		viewChangeCount = h.countViewChange(1, 4)
		require.Equal(t, 0, viewChangeCount, "Term should not ignore ViewChange message on view 4 (From the past)")
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

func TestPreprepareSignature(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)
		block3 := builders.CreateBlock(block2)

		h.setNode1AsTheLeader(ctx, 1, 1, block1)

		// start with 0 preprepare
		hasPreprepare := h.hasPreprepare(2, 1, block2)
		require.False(t, hasPreprepare, "No preprepare should exist in the storage")

		// sending a preprepare (height 2)
		h.sendPreprepare(ctx, 1, 2, 1, block2)

		// Expect the storage to have it
		hasPreprepare = h.hasPreprepare(2, 1, block2)
		require.True(t, hasPreprepare, "A preprepare should exist in the storage")

		// sending another preprepare (height 3)
		h.failFutureVerifications()
		h.sendPreprepare(ctx, 1, 3, 1, block3)

		// Expect the storage NOT to have it
		hasPreprepare = h.hasPreprepare(3, 1, block3)
		require.False(t, hasPreprepare, "preprepare should NOT exist in the storage")
	})
}

func TestPrepareSignature(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block := builders.CreateBlock(builders.GenesisBlock)

		// start with 0 prepare
		prepareCount := h.countPrepare(1, 0, block)
		require.Equal(t, 0, prepareCount, "No prepare should exist in the storage")

		// sending a prepare
		h.sendPrepare(ctx, 1, 1, 0, block)

		// Expect the storage to have it
		prepareCount = h.countPrepare(1, 0, block)
		require.Equal(t, 1, prepareCount, "1 prepare should exist in the storage")

		// sending another prepare (From a different node)
		h.failFutureVerifications()
		h.sendPrepare(ctx, 2, 1, 0, block)

		// Expect the storage NOT to have it
		prepareCount = h.countPrepare(1, 0, block)
		require.Equal(t, 1, prepareCount, "(Still) 1 prepare should exist in the storage")
	})
}
