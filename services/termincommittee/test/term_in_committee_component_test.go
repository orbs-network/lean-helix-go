// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

// Leader election //
func TestViewIncrementedAfterElectionTrigger(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)
		h.assertView(0)
		h.triggerElection(ctx)
		h.assertView(1)
	})
}

func TestNewViewNotAcceptedIfDidNotPassValidation(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(startView primitives.View, view primitives.View, failValidations bool, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			block := mocks.ABlock(interfaces.GenesisBlock)

			h.assertView(startView)
			if failValidations {
				h.failMyNodeBlockProposalValidations()
			}
			h.receiveAndHandleNewView(ctx, 2, 1, view, block)
			if shouldAcceptMessage {
				h.assertView(view)
			} else {
				h.assertView(startView)
			}
		}

		// a valid new view
		sendNewView(5, 6, false, true)

		// a failing validation new view
		sendNewView(5, 6, true, false)
	})
}

func TestNewViewNotAcceptViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(startView primitives.View, view primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			block := mocks.ABlock(interfaces.GenesisBlock)

			h.receiveAndHandleNewView(ctx, 2, 1, view, block)

			if shouldAcceptMessage {
				h.assertView(view)
			} else {
				h.assertView(startView)
			}
		}

		// notify node2 (view=6, future) as the leader
		sendNewView(5, 6, true)

		// notify node2 (view=2, past) as the leader
		sendNewView(5, 2, false)
	})
}

func TestNewViewIsSentWithTheHighestBlockFromTheViewChangeProofs(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		blockOnView3 := mocks.ABlock(interfaces.GenesisBlock)
		preparedMessagesOnView3 := builders.CreatePreparedMessages(
			h.instanceId,
			h.net.Nodes[3],
			[]builders.Sender{h.net.Nodes[0], h.net.Nodes[1], h.net.Nodes[2]},
			1,
			3,
			blockOnView3)

		blockOnView4 := mocks.ABlock(interfaces.GenesisBlock)
		preparedMessagesOnView4 := builders.CreatePreparedMessages(
			h.instanceId,
			h.net.Nodes[0],
			[]builders.Sender{h.net.Nodes[1], h.net.Nodes[2], h.net.Nodes[3]},
			1,
			4,
			blockOnView4)

		// voting node1 as the new leader (view 5)
		votes := builders.NewVotesBuilder(h.instanceId).
			WithVote(h.getMemberKeyManager(0), h.getNodeMemberId(0), 1, 5, preparedMessagesOnView3).
			WithVote(h.getMemberKeyManager(2), h.getNodeMemberId(2), 1, 5, preparedMessagesOnView4).
			WithVote(h.getMemberKeyManager(3), h.getNodeMemberId(3), 1, 5, nil).
			Build()

		h.assertView(0)

		nvm := builders.
			NewNewViewBuilder().
			LeadBy(h.getMemberKeyManager(1), h.getNodeMemberId(1)).
			WithViewChangeVotes(votes).
			OnBlock(blockOnView4).
			OnBlockHeight(1).
			OnView(5).
			Build()

		h.handleNewViewMessage(ctx, nvm)

		h.assertView(5)
		require.True(t, h.hasPreprepare(1, 5, blockOnView4))
	})
}

func TestNewViewIsRejectedIfVotesAreNotQuorum(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		// voting node1 as the new leader (view 5)
		votes := builders.NewVotesBuilder(h.instanceId).
			WithVote(h.getMemberKeyManager(0), h.getNodeMemberId(0), 1, 5, nil).
			WithVote(h.getMemberKeyManager(1), h.getNodeMemberId(1), 1, 5, nil).
			WithVote(h.getMemberKeyManager(2), h.getNodeMemberId(2), 1, 5, nil).
			Build()

		// total voting weight is 6 < 7 (quorum weight)

		h.assertView(0)

		nvm := builders.
			NewNewViewBuilder().
			LeadBy(h.getMemberKeyManager(1), h.getNodeMemberId(1)).
			WithViewChangeVotes(votes).
			OnBlockHeight(1).
			OnView(5).
			Build()

		h.handleNewViewMessage(ctx, nvm)
		h.assertView(0)

		votes = builders.NewVotesBuilder(h.instanceId).
			WithVote(h.getMemberKeyManager(0), h.getNodeMemberId(0), 1, 5, nil).
			WithVote(h.getMemberKeyManager(1), h.getNodeMemberId(1), 1, 5, nil).
			WithVote(h.getMemberKeyManager(3), h.getNodeMemberId(3), 1, 5, nil).
			Build()

		// total voting weight is 7 (quorum)

		nvm = builders.
			NewNewViewBuilder().
			LeadBy(h.getMemberKeyManager(1), h.getNodeMemberId(1)).
			WithViewChangeVotes(votes).
			OnBlockHeight(1).
			OnView(5).
			Build()

		h.handleNewViewMessage(ctx, nvm)
		h.assertView(5)
	})
}

func TestNewViewWithOlderBlockIsRejected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		blockOnView3 := mocks.ABlock(interfaces.GenesisBlock)
		preparedMessagesOnView3 := builders.CreatePreparedMessages(
			h.instanceId,
			h.net.Nodes[3],
			[]builders.Sender{h.net.Nodes[0], h.net.Nodes[1], h.net.Nodes[2]},
			1,
			3,
			blockOnView3)

		blockOnView4 := mocks.ABlock(interfaces.GenesisBlock)
		preparedMessagesOnView4 := builders.CreatePreparedMessages(
			h.instanceId,
			h.net.Nodes[0],
			[]builders.Sender{h.net.Nodes[1], h.net.Nodes[2], h.net.Nodes[3]},
			1,
			4,
			blockOnView4)

		// voting node1 as the new leader (view 5)
		votes := builders.NewVotesBuilder(h.instanceId).
			WithVote(h.getMemberKeyManager(0), h.getNodeMemberId(0), 1, 5, preparedMessagesOnView3).
			WithVote(h.getMemberKeyManager(2), h.getNodeMemberId(2), 1, 5, preparedMessagesOnView4).
			WithVote(h.getMemberKeyManager(3), h.getNodeMemberId(3), 1, 5, nil).
			Build()

		h.assertView(0)

		nvm := builders.
			NewNewViewBuilder().
			LeadBy(h.getMemberKeyManager(1), h.getNodeMemberId(1)).
			WithViewChangeVotes(votes).
			OnBlock(blockOnView3).
			OnBlockHeight(1).
			OnView(5).
			Build()

		h.handleNewViewMessage(ctx, nvm)

		h.assertView(0)
		require.False(t, h.hasPreprepare(1, 5, blockOnView3))
		require.False(t, h.hasPreprepare(1, 5, blockOnView4))
	})
}

func TestNewViewNotAcceptMessageIfNotFromTheLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(fromNodeIdx int, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			block := mocks.ABlock(interfaces.GenesisBlock)

			h.receiveAndHandleNewView(ctx, fromNodeIdx, 1, 1, block)
			if shouldAcceptMessage {
				h.assertView(1)
			} else {
				h.assertView(0)
			}
		}

		// getting a new view message from node1 (the new leader)
		sendNewView(1, true)

		// getting a new view message from node2 about node1 as the new leader
		sendNewView(2, false)
	})
}

func TestNewViewNotAcceptedWithWrongPPDetails(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(
			block interfaces.Block,
			blockHeight primitives.BlockHeight,
			view primitives.View,
			preprepareBlock interfaces.Block,
			preprepareBlockHeight primitives.BlockHeight,
			preprepareView primitives.View,
			shouldAcceptMessage bool,
		) {
			h := NewHarness(ctx, t)

			h.assertView(0)

			voters := []*builders.Voter{
				{KeyManager: h.getMemberKeyManager(0), MemberId: h.getNodeMemberId(0)},
				{KeyManager: h.getMemberKeyManager(2), MemberId: h.getNodeMemberId(2)},
				{KeyManager: h.getMemberKeyManager(3), MemberId: h.getNodeMemberId(3)},
			}
			votes := builders.ASimpleViewChangeVotes(h.instanceId, voters, blockHeight, view)

			newLeaderKeyManager := h.getMemberKeyManager(1)
			newLeaderId := h.getNodeMemberId(1)
			nvm := builders.NewNewViewBuilder().
				LeadBy(newLeaderKeyManager, newLeaderId).
				OnBlock(block).
				OnBlockHeight(blockHeight).
				OnView(view).
				WithCustomPreprepare(h.instanceId, newLeaderKeyManager, newLeaderId, preprepareBlockHeight, preprepareView, preprepareBlock).
				WithViewChangeVotes(votes).
				Build()

			h.handleNewViewMessage(ctx, nvm)

			if shouldAcceptMessage {
				h.assertView(1)
			} else {
				h.assertView(0)
			}
		}

		block := mocks.ABlock(interfaces.GenesisBlock)

		// good new view
		sendNewView(block, 10, 1, block, 10, 1, true)

		// mismatching preprepare view
		sendNewView(block, 10, 1, block, 10, 666, false)

		// mismatching preprepare block height
		sendNewView(block, 10, 1, block, 666, 1, false)
	})
}

func TestNewViewNotAcceptedWithWrongViewChangeDetails(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(blockHeight primitives.BlockHeight, view primitives.View, vcsBlockHeight [3]primitives.BlockHeight, vcsView [3]primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			block := mocks.ABlock(interfaces.GenesisBlock)

			h.assertView(0)

			votesBuilder := builders.NewVotesBuilder(h.instanceId)
			votesBuilder.WithVote(h.getMemberKeyManager(0), h.getNodeMemberId(0), vcsBlockHeight[0], vcsView[0], nil)
			votesBuilder.WithVote(h.getMemberKeyManager(2), h.getNodeMemberId(2), vcsBlockHeight[1], vcsView[1], nil)
			votesBuilder.WithVote(h.getMemberKeyManager(3), h.getNodeMemberId(3), vcsBlockHeight[2], vcsView[2], nil)
			votes := votesBuilder.Build()

			newLeaderKeyManager := h.getMemberKeyManager(1)
			newLeaderMemberId := h.getNodeMemberId(1)
			nvm := builders.NewNewViewBuilder().
				LeadBy(newLeaderKeyManager, newLeaderMemberId).
				OnBlock(block).
				OnBlockHeight(blockHeight).
				OnView(view).
				WithViewChangeVotes(votes).
				Build()

			h.handleNewViewMessage(ctx, nvm)

			if shouldAcceptMessage {
				h.assertView(1)
			} else {
				h.assertView(0)
			}
		}

		// good new view
		sendNewView(10, 1, [3]primitives.BlockHeight{10, 10, 10}, [3]primitives.View{1, 1, 1}, true)

		// mismatching view-change view
		sendNewView(10, 1, [3]primitives.BlockHeight{10, 10, 10}, [3]primitives.View{666, 1, 1}, false)
		sendNewView(10, 1, [3]primitives.BlockHeight{10, 10, 10}, [3]primitives.View{1, 666, 1}, false)
		sendNewView(10, 1, [3]primitives.BlockHeight{10, 10, 10}, [3]primitives.View{1, 1, 666}, false)

		// mismatching view-change block height
		sendNewView(10, 1, [3]primitives.BlockHeight{666, 10, 10}, [3]primitives.View{1, 1, 1}, false)
		sendNewView(10, 1, [3]primitives.BlockHeight{10, 666, 10}, [3]primitives.View{1, 1, 1}, false)
		sendNewView(10, 1, [3]primitives.BlockHeight{10, 10, 666}, [3]primitives.View{1, 1, 1}, false)
	})
}

func TestNewViewNotAcceptedWithBadVotes(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(leaderNodeIdx int, members []int, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			block := mocks.ABlock(interfaces.GenesisBlock)

			h.assertView(0)

			leaderKeyManager := h.getMemberKeyManager(leaderNodeIdx)
			leaderMemberId := h.getNodeMemberId(leaderNodeIdx)

			votesBuilder := builders.NewVotesBuilder(h.instanceId)
			for _, memberIdx := range members {
				votesBuilder.WithVote(h.net.Nodes[memberIdx].KeyManager, h.net.Nodes[memberIdx].MemberId, 10, 1, nil)
			}

			nvm := builders.
				NewNewViewBuilder().
				LeadBy(leaderKeyManager, leaderMemberId).
				WithViewChangeVotes(votesBuilder.Build()).
				OnBlock(block).
				OnBlockHeight(10).
				OnView(1).
				Build()
			h.handleNewViewMessage(ctx, nvm)

			if shouldAcceptMessage {
				h.assertView(1)
			} else {
				h.assertView(0)
			}
		}

		// good new view
		sendNewView(1, []int{0, 2, 3}, true)

		// duplicate voters
		sendNewView(1, []int{0, 2, 2}, false)

		// No votes
		sendNewView(1, []int{}, false)

		// Not enough votes
		sendNewView(1, []int{0, 2}, false)
	})
}

func TestViewChangeIgnoreViewsFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendViewChange := func(startView primitives.View, expectedEndView primitives.View, vcmView primitives.View) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			h.assertView(startView)
			h.receiveAndHandleViewChange(ctx, 1, 1, vcmView)
			h.receiveAndHandleViewChange(ctx, 2, 1, vcmView)
			h.receiveAndHandleViewChange(ctx, 3, 1, vcmView)
			h.assertView(expectedEndView)
		}

		// re-voting me (node0, view=12 -> future) as the leader
		t.Logf("startView=%d endView=%d messageView=%d", 8, 12, 12)
		sendViewChange(8, 12, 12)

		t.Logf("startView=%d endView=%d messageView=%d", 8, 8, 8)
		// re-voting me (node0, view=8 -> present) as the leader
		sendViewChange(8, 8, 8)

		// re-voting me (node0, view=4 -> past) as the leader
		t.Logf("startView=%d endView=%d messageView=%d", 8, 8, 4)
		sendViewChange(8, 8, 4)
	})
}

func TestViewChangeIsRejectedIfTargetIsNotTheNewLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendViewChange := func(startView primitives.View, view primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, view)

			viewChangeCountBefore := h.countViewChange(1, view)
			h.receiveAndHandleViewChange(ctx, 3, 1, view)
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

			block := mocks.ABlock(interfaces.GenesisBlock)

			prepareCountBefore := h.countPrepare(1, view, block)
			h.receiveAndHandlePrepare(ctx, 1, 1, view, block)
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

			block := mocks.ABlock(interfaces.GenesisBlock)

			prepareCountBefore := h.countPrepare(1, view, block)
			h.receiveAndHandlePrepare(ctx, fromNode, 1, view, block)
			prepareCountAfter := h.countPrepare(1, view, block)

			isMessageAccepted := prepareCountAfter == prepareCountBefore+1
			if shouldAcceptMessage {
				require.True(t, isMessageAccepted)
			} else {
				require.False(t, isMessageAccepted)
			}

			h.receiveAndHandlePrepare(ctx, 2, 2, 1, block)
			prepareCount := h.countPrepare(2, 1, block)
			require.Equal(t, 1, prepareCount, "TermInCommittee should not ignore Prepare message from node2")
		}

		// sending a valid prepare (From node2)
		sendPrepare(1, 1, 2, true)

		// sending an invalid prepare (From node1 - the leader)
		sendPrepare(1, 1, 1, false)
	})
}

func TestPreprepareAcceptOnlyMatchingViews(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendPreprepare := func(startView primitives.View, view primitives.View, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			block := mocks.ABlock(interfaces.GenesisBlock)

			hasPreprepare := h.hasPreprepare(1, startView, block)
			require.False(t, hasPreprepare, "No preprepare should exist in the storage")

			// current view (5) => valid
			h.receiveAndHandlePreprepare(ctx, 1, 1, view, block)
			hasPreprepare = h.hasPreprepare(1, view, block)
			if shouldAcceptMessage {
				require.True(t, hasPreprepare, "TermInCommittee should not ignore the Preprepare message")
			} else {
				require.False(t, hasPreprepare, "TermInCommittee should ignore the Preprepare message")
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

func TestPrepare2fPlus1ForACommit_fullCommtteeQuorumScenario(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)

		h := NewHarness(ctx, t, block)
		h.setNode1AsTheLeader(ctx, 1, 1, block)

		require.Equal(t, 0, h.countCommits(1, 1, block), "No commits should exist in the storage")

		h.receiveAndHandlePreprepare(ctx, 1 /* weight 2 */, 1, 1, block)
		// own weight (1) + node1 (2) = 3 < 7
		require.Equal(t, 0, h.countCommits(1, 1, block), "No commits should exist in the storage")

		h.receiveAndHandlePrepare(ctx, 2 /* weight 3 */, 1, 1, block)
		// own weight (1) + node1 (2) + node2 (3) = 6 < 7
		require.Equal(t, 0, h.countCommits(1, 1, block), "No commits should exist in the storage")

		h.receiveAndHandlePrepare(ctx, 3 /* weight 4 */, 1, 1, block)
		// all nodes - quorum
		require.Equal(t, 1, h.countCommits(1, 1, block), "There should be 1 commit in the storage")
	})
}

func TestPrepare2fPlus1ForACommit_smallerQuorumScenario(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)

		h := NewHarness(ctx, t, block)
		h.setNode1AsTheLeader(ctx, 1, 1, block)

		require.Equal(t, 0, h.countCommits(1, 1, block), "No commits should exist in the storage")

		h.receiveAndHandlePreprepare(ctx, 1 /* weight 2 */, 1, 1, block)
		// own weight (1) + node1 (2) = 3 < 7
		require.Equal(t, 0, h.countCommits(1, 1, block), "No commits should exist in the storage")

		h.receiveAndHandlePrepare(ctx, 3 /* weight 3 */, 1, 1, block)
		// own weight (1) + node1 (2) + node3 (4) = 7, quorum
		require.Equal(t, 1, h.countCommits(1, 1, block), "There should be 1 commit in the storage")

		h.receiveAndHandlePrepare(ctx, 2 /* weight 4 */, 1, 1, block)
		require.Equal(t, 1, h.countCommits(1, 1, block), "There should be 1 commit in the storage")
	})
}

func TestCommit2fPlus1_fullCommtteeQuorumScenario(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)

		committed := false
		onCommit := func(ctx context.Context, block interfaces.Block, commitMessages []*interfaces.CommitMessage) {
			committed = true
		}

		h := NewHarnessForNodeInd(ctx, 0, onCommit, t, []interfaces.Block{block})
		h.setNode1AsTheLeader(ctx, 1, 1, block)

		require.Equal(t, 0, h.countCommits(1, 1, block), "No commits should exist in the storage")

		h.receiveAndHandlePreprepare(ctx, 1, 1, 1, block)
		h.receiveAndHandlePrepare(ctx, 2, 1, 1, block)
		h.receiveAndHandlePrepare(ctx, 3, 1, 1, block)
		require.Equal(t, 1, h.countCommits(1, 1, block), "There should be 1 commit in the storage")

		// Now prepared
		// node0 (weight 1) already committed

		h.receiveAndHandleCommit(ctx, 1 /* weight 2 */, 1, 1, block, 0)
		require.False(t, committed, "Should not have committed")
		h.receiveAndHandleCommit(ctx, 2 /* weight 3 */, 1, 1, block, 0)
		require.False(t, committed, "Should not have committed")

		// Total committed weight is 6 < 7, no quorum yet

		h.receiveAndHandleCommit(ctx, 3, 1, 1, block, 0)
		require.True(t, committed, "Should have committed")
	})
}

func TestCommit2fPlus1_smallerQuorumScenario(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)

		committed := false
		onCommit := func(ctx context.Context, block interfaces.Block, commitMessages []*interfaces.CommitMessage) {
			committed = true
		}

		h := NewHarnessForNodeInd(ctx, 0, onCommit, t, []interfaces.Block{block})
		h.setNode1AsTheLeader(ctx, 1, 1, block)

		require.Equal(t, 0, h.countCommits(1, 1, block), "No commits should exist in the storage")

		h.receiveAndHandlePreprepare(ctx, 1, 1, 1, block)
		h.receiveAndHandlePrepare(ctx, 2, 1, 1, block)
		h.receiveAndHandlePrepare(ctx, 3, 1, 1, block)
		require.Equal(t, 1, h.countCommits(1, 1, block), "There should be 1 commit in the storage")

		// Now prepared
		// node0 (weight 1) already committed

		h.receiveAndHandleCommit(ctx, 1 /* weight 2 */, 1, 1, block, 0)
		require.False(t, committed, "Should not have committed")
		h.receiveAndHandleCommit(ctx, 3 /* weight 4 */, 1, 1, block, 0)
		// Total weight is 7 (1 + 2 + 4), reached quorum
		require.True(t, committed, "Should have committed")
	})
}

func TestViewChange2fPlus1ToBecomeElected_largeQuorumScenario(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)

		h := NewHarnessForNodeInd(ctx, 1, nil, t, []interfaces.Block{block})

		require.False(t, h.hasPreprepare(1, 1, block), "Should not have a preprepare message")

		h.receiveAndHandleViewChange(ctx, 0 /* weight 1 */, 1, 1)
		// node0 (1)  < 7
		require.False(t, h.hasPreprepare(1, 1, block), "Should not have a preprepare message")

		h.receiveAndHandleViewChange(ctx, 1 /* weight 2 */, 1, 1)
		// node0 (1) + node1 (2) = 3 < 7
		require.False(t, h.hasPreprepare(1, 1, block), "Should not have a preprepare message")

		h.receiveAndHandleViewChange(ctx, 3 /* weight 4 */, 1, 1)
		// node0 (1) + node1 (2) + node3 (4) = 7, quorum
		require.True(t, h.hasPreprepare(1, 1, block), "There should be a preprepare message")

		h.receiveAndHandleViewChange(ctx, 2 /* weight 3 */, 1, 1)
		require.True(t, h.hasPreprepare(1, 1, block), "There should still be a preprepare message")
	})
}

func TestViewChange2fPlus1ToBecomeElected_smallerQuorumScenario(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)

		h := NewHarnessForNodeInd(ctx, 1, nil, t, []interfaces.Block{block})

		require.False(t, h.hasPreprepare(1, 1, block), "Should not have a preprepare message")

		h.receiveAndHandleViewChange(ctx, 2 /* weight 3 */, 1, 1)
		// node2 (3) = 6 < 7
		require.False(t, h.hasPreprepare(1, 1, block), "Should not have a preprepare message")

		h.receiveAndHandleViewChange(ctx, 3 /* weight 4 */, 1, 1)
		// node2 (3) + node3(4) = 7, quorum
		require.True(t, h.hasPreprepare(1, 1, block), "There should be a preprepare message")

		h.receiveAndHandleViewChange(ctx, 0 /* weight 1 */, 1, 1)
		require.True(t, h.hasPreprepare(1, 1, block), "There should still be a preprepare message")
	})
}

func TestDisposingATermInCommitteeStopsTheElectionTrigger(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)

		h := NewHarness(ctx, t, block)

		// view was changed because of election
		h.assertView(0)
		h.triggerElection(ctx)
		h.assertView(1)

		// dispose the termInCommittee
		h.disposeTerm()

		// view was not changed
		h.assertView(1)
		h.triggerElection(ctx)
		h.assertView(1)
	})
}

func TestDisposingATermInCommitteeClearTheStorage(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)

		h := NewHarness(ctx, t, block)

		// good consensus on block
		h.receiveAndHandlePrepare(ctx, 1, 1, 0, block)
		h.receiveAndHandlePrepare(ctx, 2, 1, 0, block)

		// dispose the termInCommittee
		h.disposeTerm()

		// make sure that all the messages are cleared from the storage
		require.False(t, h.hasPreprepare(1, 0, block), "There should be no preprepare in the storage")
		require.Equal(t, 0, h.countPrepare(1, 0, block), "There should be no prepares in the storage")
		require.Equal(t, 0, h.countCommits(1, 0, block), "There should be no commit in the storage")
	})
}

func TestAValidPreparedProofIsSentOnViewChange(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)

		h := NewHarness(ctx, t, block)

		// Get prepared on block
		h.receiveAndHandlePrepare(ctx, 2, 1, 0, block)
		h.receiveAndHandlePrepare(ctx, 3, 1, 0, block)

		h.triggerElection(ctx)

		msg := h.getLastSentViewChangeMessage()
		msgContent := msg.Content()
		vcSenderId := msgContent.Sender().MemberId()
		vcHeader := msgContent.SignedHeader()
		resultView := vcHeader.View()
		resultHeight := vcHeader.BlockHeight()
		preparedProof := vcHeader.PreparedProof()
		ppSenderId := preparedProof.PreprepareSender().MemberId()
		ppBlockRef := preparedProof.PreprepareBlockRef()
		pBlockRef := preparedProof.PrepareBlockRef()

		var pSendersIds []primitives.MemberId
		pSendersIter := preparedProof.PrepareSendersIterator()
		for {
			if !pSendersIter.HasNext() {
				break
			}
			pSendersIds = append(pSendersIds, pSendersIter.NextPrepareSenders().MemberId())
		}
		require.Equal(t, 2, len(pSendersIds), "expected 2 senders of Prepare messages but got %d", len(pSendersIds))

		member2Id := h.getNodeMemberId(2)
		member3Id := h.getNodeMemberId(3)

		pSendersEqual := (member3Id.Equal(pSendersIds[0]) && member2Id.Equal(pSendersIds[1])) ||
			(member3Id.Equal(pSendersIds[1]) && member2Id.Equal(pSendersIds[0]))

		require.True(t, pSendersEqual)
		require.Equal(t, primitives.BlockHeight(1), pBlockRef.BlockHeight())
		require.Equal(t, primitives.View(0), pBlockRef.View())
		require.Equal(t, primitives.BlockHeight(1), ppBlockRef.BlockHeight())
		require.Equal(t, primitives.View(0), ppBlockRef.View())
		require.Equal(t, h.getMyNodeMemberId(), vcSenderId)
		require.Equal(t, h.getMyNodeMemberId(), ppSenderId)
		require.Equal(t, primitives.View(1), resultView)
		require.Equal(t, primitives.BlockHeight(1), resultHeight)
		require.Equal(t, block, msg.Block())
	})
}

func TestAValidViewChangeMessageWithPreparedProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		h := NewHarness(ctx, t)
		h.setNode1AsTheLeader(ctx, 10, 1, block1)

		preparedMessages := &preparedmessages.PreparedMessages{
			PreprepareMessage: builders.APreprepareMessage(h.instanceId, h.getMyKeyManager(), h.myMemberId, 1, 0, block2),
			PrepareMessages: []*interfaces.PrepareMessage{
				builders.APrepareMessage(h.instanceId, h.getMemberKeyManager(2), h.getNodeMemberId(2), 1, 0, block2),
				builders.APrepareMessage(h.instanceId, h.getMemberKeyManager(3), h.getNodeMemberId(3), 1, 0, block2),
			},
		}

		msg := builders.AViewChangeMessage(h.instanceId, h.getMyKeyManager(), h.myMemberId, 10, 4, preparedMessages)
		h.handleViewChangeMessage(ctx, msg)

		require.Exactly(t, 1, h.countViewChange(10, 4))
	})
}

func TestViewChangeMessageWithoutQuorumInThePreparedProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		// an invalid prepare messages
		h := NewHarness(ctx, t)
		h.setNode1AsTheLeader(ctx, 10, 1, block1)

		preparedMessages := &preparedmessages.PreparedMessages{
			PreprepareMessage: builders.APreprepareMessage(h.instanceId, h.getMyKeyManager(), h.myMemberId, 1, 0, block2),
			PrepareMessages: []*interfaces.PrepareMessage{
				builders.APrepareMessage(h.instanceId, h.getMemberKeyManager(1), h.getNodeMemberId(1), 1, 0, block2),
			}, // not enough
		}

		msg := builders.AViewChangeMessage(h.instanceId, h.getMyKeyManager(), h.myMemberId, 10, 4, preparedMessages)
		h.handleViewChangeMessage(ctx, msg)

		require.Exactly(t, 0, h.countViewChange(10, 4))
	})
}

func TestViewChangeMessageWithAnInvalidPreparedProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		// an invalid prepare messages
		h := NewHarness(ctx, t)
		h.setNode1AsTheLeader(ctx, 10, 1, block1)

		preparedMessages := &preparedmessages.PreparedMessages{
			PreprepareMessage: builders.APreprepareMessage(h.instanceId, h.getMyKeyManager(), h.myMemberId, 1, 0, block2),
			PrepareMessages:   nil, // BAD
		}

		msg := builders.AViewChangeMessage(h.instanceId, h.getMyKeyManager(), h.myMemberId, 10, 4, preparedMessages)
		h.handleViewChangeMessage(ctx, msg)

		require.Exactly(t, 0, h.countViewChange(10, 4))
	})
}
