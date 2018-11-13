package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test"
	"testing"
)

// Leader election //
func TestViewIncrementedAfterElectionTrigger(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)
		h.startConsensus(ctx)
		h.waitForView(0)

		h.triggerElection()
		h.waitForView(1)
	})
}

// TODO: uncomment
//// View Change messages //
//func TestViewIncrementedAfterEnoughViewChangeMessages(t *testing.T) {
//	test.WithContext(func(ctx context.Context) {
//		h := NewHarness(ctx, t)
//		h.startConsensus(ctx)
//		h.waitForView(0)
//
//		h.sendLeaderChange(ctx, 1) // next view
//		h.waitForView(1)
//	})
//}
//
//func TestRejectNewViewMessagesFromPast(t *testing.T) {
//	test.WithContext(func(ctx context.Context) {
//		h := NewHarness(ctx, t)
//		h.startConsensus(ctx)
//		h.waitForView(0)
//
//		h.sendLeaderChange(ctx, 1) // next view, good
//		h.waitForView(1)
//
//		h.sendLeaderChange(ctx, 1) // same view, ignored
//		h.verifyViewDoesNotChange(1)
//	})
//}

// TODO: uncomment
//func TestRejectNewViewMessagesFromPast(t *testing.T) {
//	WithContext(func(ctx context.Context) {
//		height := BlockHeight(0)
//		view := View(0)
//		block := builders.CreateBlock(builders.GenesisBlock)
//		net := builders.ABasicTestNetwork()
//
//		node := net.Nodes[0]
//		messageFactory := lh.NewMessageFactory(node.KeyManager)
//		ppmContentBuilder := messageFactory.CreatePreprepareMessageContentBuilder(height, view, block)
//		termConfig := net.Nodes[0].BuildConfig()
//		filter := lh.NewConsensusMessageFilter(termConfig.KeyManager.MyPublicKey())
//		term := lh.NewLeanHelixTerm(termConfig, filter, 0)
//
//		require.Equal(t, View(0), term.GetView(), "Term should have view=0 on init")
//		net.TriggerElection()
//		require.Equal(t, View(1), term.GetView(), "Term should have view=1 after one election")
//		nvm := builders.ANewViewMessage(node.KeyManager, height, view, ppmContentBuilder, nil, block)
//		term.OnReceiveNewView(ctx, nvm)
//		require.Equal(t, View(1), term.GetView(), "Term should have view=1")
//	})
//}
