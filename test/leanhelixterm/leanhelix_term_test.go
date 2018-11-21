package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
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
		h.sendLeaderChanged(ctx, 8, block)
		h.waitForView(8)

		// re-voting node0 as the leader, but with a view from the past (4)
		block = builders.CreateBlock(builders.GenesisBlock)
		h.sendLeaderChanged(ctx, 4, block)
		h.waitForView(8) // unchanged
	})
}

//func TestNoConsensusWhenValidationFailed(t *testing.T) {
//	test.WithContext(func(ctx context.Context) {
//		block1 := builders.CreateBlock(builders.GenesisBlock)
//		block2 := builders.CreateBlock(block1)
//
//		net := builders.ATestNetwork(ctx, 4, block1, block2)
//		node1 := net.Nodes[1]
//		node2 := net.Nodes[2]
//		node3 := net.Nodes[3]
//
//		net.NodesPauseOnValidate()
//		net.StartConsensus(ctx)
//
//		// Block1, should pass
//		node1.BlockUtils.ValidationResult = true
//		node2.BlockUtils.ValidationResult = true
//		node3.BlockUtils.ValidationResult = true
//		net.WaitForNodesToValidate(node1, node2, node3)
//		net.ResumeNodesValidation(node1, node2, node3)
//		require.True(t, net.WaitForAllNodesToCommitBlock(block1))
//
//		// Block2, should fail
//		node1.BlockUtils.ValidationResult = false
//		node2.BlockUtils.ValidationResult = false
//		node3.BlockUtils.ValidationResult = false
//		net.WaitForNodesToValidate(node1, node2, node3)
//		node1.PauseOnTick()
//		node2.PauseOnTick()
//		node3.PauseOnTick()
//		net.ResumeNodesValidation(node1, node2, node3)
//
//		node1.WaitForPause()
//		node2.WaitForPause()
//		node3.WaitForPause()
//		require.True(t, net.AllNodesChainEndsWithABlock(block1))
//	})
//}

// TODO: uncomment
//// View Change messages //
//func TestViewIncrementedAfterEnoughViewChangeMessages(t *testing.T) {
//	test.WithContext(func(ctx context.Context) {
//		h := NewHarness(ctx, t)
//		h.startConsensus(ctx)
//		h.waitForView(0)
//
//		h.sendLeaderChanged(ctx, 1) // next view
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
//		h.sendLeaderChanged(ctx, 1) // next view, good
//		h.waitForView(1)
//
//		h.sendLeaderChanged(ctx, 1) // same view, ignored
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
