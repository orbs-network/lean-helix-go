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
		h := NewHarness(t)
		h.startConsensus(ctx)
		h.waitForView(0)

		h.triggerElection()
		h.waitForView(1)
	})
}

func TestNewViewNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(t)

		h.startConsensus(ctx)

		// moving to node1 as the leader
		h.waitForView(0)
		h.triggerElection()
		h.waitForView(1)

		// voting node0 as the leader
		block := builders.CreateBlock(builders.GenesisBlock)
		h.sendLeaderChanged(ctx, 1, 8, block)
		h.waitForView(8)

		// re-voting node0 as the leader, but with a view from the past (4)
		block = builders.CreateBlock(builders.GenesisBlock)
		h.sendLeaderChanged(ctx, 1, 4, block)
		h.waitForView(8) // unchanged
	})
}

func TestViewChangeNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(t)

		h.startConsensus(ctx)

		// moving to node1 as the leader
		h.waitForView(0)
		h.triggerElection()
		h.waitForView(1)
		h.triggerElection()
		h.waitForView(2)

		// voting node1 (view=5) as the leader
		block := builders.CreateBlock(builders.GenesisBlock)
		h.waitForTick()
		h.sendChangeLeader(ctx, 1, 5, block)
		h.resume()
		h.waitForTick()
		viewChangeCount := h.countViewChange(1, 9)
		require.Equal(t, viewChangeCount, 1, "Term should not ignore ViewChange message on view 9")
		h.resume()

		// re-voting node1 as the leader, but with a view from the past (1)
		block = builders.CreateBlock(builders.GenesisBlock)
		h.sendChangeLeader(ctx, 1, 1, block)
		h.waitForTick()
		viewChangeCount = h.countViewChange(1, 1)
		require.Equal(t, viewChangeCount, 0, "Term should not ignore ViewChange message on view 1 (From the past)")
		h.resume()
	})
}
