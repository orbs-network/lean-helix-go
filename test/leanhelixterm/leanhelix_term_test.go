package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
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

func TestNewViewNotAcceptedIfDidNotPassValidation(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(startView primitives.View, view primitives.View, failValidations bool, shouldAcceptMessage bool) {
			h := NewHarness(ctx, t)
			h.electionTillView(ctx, startView)

			block := builders.CreateBlock(builders.GenesisBlock)

			h.checkView(startView)
			if failValidations {
				h.failValidations()
			}
			h.receiveNewView(ctx, 2, 1, view, block)
			if shouldAcceptMessage {
				h.checkView(view)
			} else {
				h.checkView(startView)
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

func TestNewViewIsSentWithTheHighestBlockFromTheViewChangeProofs(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		blockOnView3 := builders.CreateBlock(builders.GenesisBlock)
		preparedMessagesOnView3 := builders.CreatePreparedMessages(
			h.getMemberKeyManager(3),
			[]leanhelix.KeyManager{h.getMemberKeyManager(0), h.getMemberKeyManager(1), h.getMemberKeyManager(2)},
			1,
			3,
			blockOnView3)

		blockOnView4 := builders.CreateBlock(builders.GenesisBlock)
		preparedMessagesOnView4 := builders.CreatePreparedMessages(
			h.getMemberKeyManager(0),
			[]leanhelix.KeyManager{h.getMemberKeyManager(1), h.getMemberKeyManager(2), h.getMemberKeyManager(3)},
			1,
			4,
			blockOnView4)

		// voting node1 as the new leader (view 5)
		votes := builders.NewVotesBuilder().
			WithVoter(h.getMemberKeyManager(0), 1, 5, preparedMessagesOnView3).
			WithVoter(h.getMemberKeyManager(2), 1, 5, preparedMessagesOnView4).
			WithVoter(h.getMemberKeyManager(3), 1, 5, nil).
			Build()

		h.checkView(0)

		nvm := builders.
			NewNewViewBuilder().
			LeadBy(h.getMemberKeyManager(1)).
			WithViewChangeVotes(votes).
			OnBlock(blockOnView4).
			OnBlockHeight(1).
			OnView(5).
			Build()

		h.HandleLeanHelixNewView(ctx, nvm)

		h.checkView(5)
		require.True(t, h.hasPreprepare(1, 5, blockOnView4))
	})
}

func TestNewViewWithOlderBlockIsRejected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		blockOnView3 := builders.CreateBlock(builders.GenesisBlock)
		preparedMessagesOnView3 := builders.CreatePreparedMessages(
			h.getMemberKeyManager(3),
			[]leanhelix.KeyManager{h.getMemberKeyManager(0), h.getMemberKeyManager(1), h.getMemberKeyManager(2)},
			1,
			3,
			blockOnView3)

		blockOnView4 := builders.CreateBlock(builders.GenesisBlock)
		preparedMessagesOnView4 := builders.CreatePreparedMessages(
			h.getMemberKeyManager(0),
			[]leanhelix.KeyManager{h.getMemberKeyManager(1), h.getMemberKeyManager(2), h.getMemberKeyManager(3)},
			1,
			4,
			blockOnView4)

		// voting node1 as the new leader (view 5)
		votes := builders.NewVotesBuilder().
			WithVoter(h.getMemberKeyManager(0), 1, 5, preparedMessagesOnView3).
			WithVoter(h.getMemberKeyManager(2), 1, 5, preparedMessagesOnView4).
			WithVoter(h.getMemberKeyManager(3), 1, 5, nil).
			Build()

		h.checkView(0)

		nvm := builders.
			NewNewViewBuilder().
			LeadBy(h.getMemberKeyManager(1)).
			WithViewChangeVotes(votes).
			OnBlock(blockOnView3).
			OnBlockHeight(1).
			OnView(5).
			Build()

		h.HandleLeanHelixNewView(ctx, nvm)

		h.checkView(0)
		require.False(t, h.hasPreprepare(1, 5, blockOnView3))
		require.False(t, h.hasPreprepare(1, 5, blockOnView4))
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

func TestNewViewNotAcceptedWithWrongPPDetails(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		sendNewView := func(
			block leanhelix.Block,
			blockHeight primitives.BlockHeight,
			view primitives.View,
			preprepareBlock leanhelix.Block,
			preprepareBlockHeight primitives.BlockHeight,
			preprepareView primitives.View,
			shouldAcceptMessage bool,
		) {
			h := NewHarness(ctx, t)

			h.checkView(0)

			membersKeyManagers := []leanhelix.KeyManager{h.getMemberKeyManager(0), h.getMemberKeyManager(2), h.getMemberKeyManager(3)}
			votes := builders.ASimpleViewChangeVotes(membersKeyManagers, blockHeight, view)

			newLeaderKeyManager := h.getMemberKeyManager(1)
			nvm := builders.NewNewViewBuilder().
				LeadBy(newLeaderKeyManager).
				OnBlock(block).
				OnBlockHeight(blockHeight).
				OnView(view).
				WithCustomPreprepare(newLeaderKeyManager, preprepareBlockHeight, preprepareView, preprepareBlock).
				WithViewChangeVotes(votes).
				Build()

			h.HandleLeanHelixNewView(ctx, nvm)

			if shouldAcceptMessage {
				h.checkView(1)
			} else {
				h.checkView(0)
			}
		}

		block := builders.CreateBlock(builders.GenesisBlock)
		mismatchingBlock := builders.CreateBlock(builders.GenesisBlock)

		// good new view
		sendNewView(block, 10, 1, block, 10, 1, true)

		// mismatching preprepare block
		sendNewView(block, 10, 1, mismatchingBlock, 10, 1, false)

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
			block := builders.CreateBlock(builders.GenesisBlock)

			h.checkView(0)

			votesBuilder := builders.NewVotesBuilder()
			votesBuilder.WithVoter(h.getMemberKeyManager(0), vcsBlockHeight[0], vcsView[0], nil)
			votesBuilder.WithVoter(h.getMemberKeyManager(2), vcsBlockHeight[1], vcsView[1], nil)
			votesBuilder.WithVoter(h.getMemberKeyManager(3), vcsBlockHeight[2], vcsView[2], nil)
			votes := votesBuilder.Build()

			newLeaderKeyManager := h.getMemberKeyManager(1)
			nvm := builders.NewNewViewBuilder().
				LeadBy(newLeaderKeyManager).
				OnBlock(block).
				OnBlockHeight(blockHeight).
				OnView(view).
				WithViewChangeVotes(votes).
				Build()

			h.HandleLeanHelixNewView(ctx, nvm)

			if shouldAcceptMessage {
				h.checkView(1)
			} else {
				h.checkView(0)
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
			block := builders.CreateBlock(builders.GenesisBlock)

			h.checkView(0)

			leaderKeyManager := h.getMemberKeyManager(leaderNodeIdx)

			votesBuilder := builders.NewVotesBuilder()
			for _, memberIdx := range members {
				votesBuilder.WithVoter(h.net.Nodes[memberIdx].KeyManager, 10, 1, nil)
			}

			nvm := builders.
				NewNewViewBuilder().
				LeadBy(leaderKeyManager).
				WithViewChangeVotes(votesBuilder.Build()).
				OnBlock(block).
				OnBlockHeight(10).
				OnView(1).
				Build()
			h.HandleLeanHelixNewView(ctx, nvm)

			if shouldAcceptMessage {
				h.checkView(1)
			} else {
				h.checkView(0)
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
		sendPreprepare := func(startView primitives.View, block leanhelix.Block, blockHash primitives.BlockHash, shouldAcceptMessage bool) {
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
		block := builders.CreateBlock(builders.GenesisBlock)

		h := NewHarness(ctx, t, block)
		h.setNode1AsTheLeader(ctx, 1, 1, block)

		require.Equal(t, 0, h.countCommits(1, 1, block), "No commits should exist in the storage")
		h.receivePreprepare(ctx, 1, 1, 1, block)

		require.Equal(t, 0, h.countCommits(1, 1, block), "No commits should exist in the storage")
		h.receivePrepare(ctx, 2, 1, 1, block)

		require.Equal(t, 1, h.countCommits(1, 1, block), "There should be 1 commit in the storage")
		h.receivePrepare(ctx, 3, 1, 1, block)

		require.Equal(t, 1, h.countCommits(1, 1, block), "There should be 1 commit in the storage")
	})
}

func TestDisposingALeanHelixTermClearTheStorage(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := builders.CreateBlock(builders.GenesisBlock)

		h := NewHarness(ctx, t, block)

		// good consensus on block
		h.receivePrepare(ctx, 1, 1, 0, block)
		h.receivePrepare(ctx, 2, 1, 0, block)

		// make sure we have all the messages in the storage
		require.True(t, h.hasPreprepare(1, 0, block), "There should be a preprepare in the storage")
		require.Equal(t, 2, h.countPrepare(1, 0, block), "There should be 3 prepares in the storage")
		require.Equal(t, 1, h.countCommits(1, 0, block), "There should be 1 commit in the storage")

		// dispose the term
		h.disposeTerm()

		// make sure that all the messages are cleared from the storage
		require.False(t, h.hasPreprepare(1, 0, block), "There should be no preprepare in the storage")
		require.Equal(t, 0, h.countPrepare(1, 0, block), "There should be no prepares in the storage")
		require.Equal(t, 0, h.countCommits(1, 0, block), "There should be no commit in the storage")
	})
}

func TestAValidPreparedProofIsSentOnViewChange(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := builders.CreateBlock(builders.GenesisBlock)

		h := NewHarness(ctx, t, block)

		// Get prepared on block
		h.receivePrepare(ctx, 1, 1, 0, block)
		h.receivePrepare(ctx, 2, 1, 0, block)

		h.triggerElection(ctx)

		msg := h.getLastSentViewChangeMessage()
		msgContent := msg.Content()
		vcSenderPK := msgContent.Sender().MemberId()
		vcHeader := msgContent.SignedHeader()
		resultView := vcHeader.View()
		resultHeight := vcHeader.BlockHeight()
		preparedProof := vcHeader.PreparedProof()
		ppSenderPK := preparedProof.PreprepareSender().MemberId()
		ppBlockRef := preparedProof.PreprepareBlockRef()
		pBlockRef := preparedProof.PrepareBlockRef()

		var pSendersPKs []primitives.MemberId
		pSendersIter := preparedProof.PrepareSendersIterator()
		for {
			if !pSendersIter.HasNext() {
				break
			}
			pSendersPKs = append(pSendersPKs, pSendersIter.NextPrepareSenders().MemberId())
		}

		member1PK := h.getMemberPk(1)
		member2PK := h.getMemberPk(2)
		pSendersEqual := (member1PK.Equal(pSendersPKs[0]) && member2PK.Equal(pSendersPKs[1])) ||
			(member1PK.Equal(pSendersPKs[1]) && member2PK.Equal(pSendersPKs[0]))

		require.True(t, pSendersEqual)
		require.Equal(t, primitives.BlockHeight(1), pBlockRef.BlockHeight())
		require.Equal(t, primitives.View(0), pBlockRef.View())
		require.Equal(t, primitives.BlockHeight(1), ppBlockRef.BlockHeight())
		require.Equal(t, primitives.View(0), ppBlockRef.View())
		require.Equal(t, h.getMyNodePk(), vcSenderPK)
		require.Equal(t, h.getMyNodePk(), ppSenderPK)
		require.Equal(t, primitives.View(1), resultView)
		require.Equal(t, primitives.BlockHeight(1), resultHeight)
		require.Equal(t, block, msg.Block())
	})
}

func TestAValidViewChangeMessageWithPreparedProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		h := NewHarness(ctx, t)
		h.setNode1AsTheLeader(ctx, 10, 1, block1)

		preparedMessages := &leanhelix.PreparedMessages{
			PreprepareMessage: builders.APreprepareMessage(h.getMyKeyManager(), 1, 0, block2),
			PrepareMessages: []*leanhelix.PrepareMessage{
				builders.APrepareMessage(h.getMemberKeyManager(1), 1, 0, block2),
				builders.APrepareMessage(h.getMemberKeyManager(2), 1, 0, block2),
			},
		}

		msg := builders.AViewChangeMessage(h.getMyKeyManager(), 10, 4, preparedMessages)
		h.receiveViewChangeMessage(ctx, msg)

		require.Exactly(t, 1, h.countViewChange(10, 4))
	})
}

func TestViewChangeMessageWithoutQuorumInThePreparedProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		// an invalid prepare messages
		h := NewHarness(ctx, t)
		h.setNode1AsTheLeader(ctx, 10, 1, block1)

		preparedMessages := &leanhelix.PreparedMessages{
			PreprepareMessage: builders.APreprepareMessage(h.getMyKeyManager(), 1, 0, block2),
			PrepareMessages: []*leanhelix.PrepareMessage{
				builders.APrepareMessage(h.getMemberKeyManager(1), 1, 0, block2),
			}, // <- not enough
		}

		msg := builders.AViewChangeMessage(h.getMyKeyManager(), 10, 4, preparedMessages)
		h.receiveViewChangeMessage(ctx, msg)

		require.Exactly(t, 0, h.countViewChange(10, 4))
	})
}

func TestViewChangeMessageWithAnInvalidPreparedProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(builders.GenesisBlock)
		block2 := builders.CreateBlock(block1)

		// an invalid prepare messages
		h := NewHarness(ctx, t)
		h.setNode1AsTheLeader(ctx, 10, 1, block1)

		preparedMessages := &leanhelix.PreparedMessages{
			PreprepareMessage: builders.APreprepareMessage(h.getMyKeyManager(), 1, 0, block2),
			PrepareMessages:   nil, // <- BAD
		}

		msg := builders.AViewChangeMessage(h.getMyKeyManager(), 10, 4, preparedMessages)
		h.receiveViewChangeMessage(ctx, msg)

		require.Exactly(t, 0, h.countViewChange(10, 4))
	})
}
