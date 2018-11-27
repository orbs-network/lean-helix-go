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

		h.checkView(0)
		h.triggerElection(ctx)
		h.checkView(1)
	})
}

func TestNewViewNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(t)

		h.checkView(0)
		h.triggerElection(ctx)
		h.checkView(1)

		// voting node0 as the leader
		block := builders.CreateBlock(builders.GenesisBlock)
		h.sendNewView(ctx, 1, 8, block)
		h.checkView(8)

		// re-voting node0 as the leader, but with a view from the past (4)
		block = builders.CreateBlock(builders.GenesisBlock)
		h.sendNewView(ctx, 1, 4, block)
		h.checkView(8) // unchanged
	})
}

func TestViewChangeNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(t)

		// jumping to view=8 me (node0) as the leader
		h.checkView(0)
		block := builders.CreateBlock(builders.GenesisBlock)
		h.sendNewView(ctx, 1, 8, block)
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
