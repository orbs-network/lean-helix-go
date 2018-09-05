package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/inmemoryblockchain"
	"github.com/orbs-network/lean-helix-go/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProofsValidator(t *testing.T) {
	keyManager := leanhelix.NewMockKeyManager("Dummy PK")
	leaderKeyManager := leanhelix.NewMockKeyManager("Leader PK")
	node1KeyManager := leanhelix.NewMockKeyManager("Node 1")
	node2KeyManager := leanhelix.NewMockKeyManager("Node 2")

	membersPKs := []types.PublicKey{"Leader PK", "Node 1", "Node 2", "Node 3"}
	calcLeaderPk := func(view types.ViewCounter) types.PublicKey {
		return membersPKs[view]
	}

	const f = 1
	const term = 0
	const view = 0
	const targetTerm = term
	const targetView = view + 1
	block := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	leaderMsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, leaderKeyManager)
	node1MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, node1KeyManager)
	node2MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, node2KeyManager)

	preprepareMessage := leaderMsgFactory.CreatePreprepareMessage(term, view, block)
	prepareMessage1 := node1MsgFactory.CreatePrepareMessage(term, view, block)
	prepareMessage2 := node2MsgFactory.CreatePrepareMessage(term, view, block)
	preparedProof := &leanhelix.PreparedProof{
		PreprepareBlockRefMessage: preprepareMessage,
		PrepareBlockRefMessages:   []*leanhelix.PrepareMessage{prepareMessage1, prepareMessage2},
	}

	t.Run("TestProofsValidatorWithNoPrePrepare", func(t *testing.T) {
		preparedProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: nil,
			PrepareBlockRefMessages:   []*leanhelix.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have a preprepare message")
	})

	t.Run("TestProofsValidatorWithNoPrepares", func(t *testing.T) {
		preparedProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage,
			PrepareBlockRefMessages:   nil,
		}
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have prepare messages")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, nil, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a nil proof")
	})

	t.Run("TestProofsValidatorWithBadPreprepareSignature", func(t *testing.T) {
		keyManager := leanhelix.NewMockKeyManager("Dummy PK", "Leader PK")
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass preprepare signature validation")
	})

	t.Run("TestProofsValidatorWithBadPrepareSignature", func(t *testing.T) {
		keyManager := leanhelix.NewMockKeyManager("Dummy PK", "Node 2")
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass prepare signature validation")
	})

	t.Run("TestProofsValidatorWithNotEnoughPrepareMessages", func(t *testing.T) {
		preparedProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage,
			PrepareBlockRefMessages:   []*leanhelix.PrepareMessage{prepareMessage1},
		}
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with not enough prepares")
	})

	t.Run("TestProofsValidatorWithTerm", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(666, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with mismatching term")
	})

	t.Run("TestProofsValidatorWithTheSameView", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(targetTerm, view, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with equal targetView")
	})

	t.Run("TestProofsValidatorWithTheSmallerView", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView-1, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with smaller targetView")
	})

	t.Run("TestProofsValidatorWithANoneMember", func(t *testing.T) {
		noneMemberKeyManager := leanhelix.NewMockKeyManager("Not in members PK")
		mf := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, noneMemberKeyManager)
		prepareMessage1 := mf.CreatePrepareMessage(term, view, block)
		preparedProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage,
			PrepareBlockRefMessages:   []*leanhelix.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a none member")
	})

	t.Run("TestProofsValidatorWithPrepareFromTheLeader", func(t *testing.T) {
		mf := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, leaderKeyManager)
		prepareMessage1 := mf.CreatePrepareMessage(term, view, block)
		preparedProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage,
			PrepareBlockRefMessages:   []*leanhelix.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a prepare from the leader")
	})

	t.Run("TestProofsValidatorWithMismatchingViewToLeader", func(t *testing.T) {
		calcLeaderPk := func(view types.ViewCounter) types.PublicKey {
			return "Some other node PK"
		}
		preparedProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage,
			PrepareBlockRefMessages:   []*leanhelix.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a mismatching view to leader")
	})

	t.Run("TestProofsValidatorWithMismatchingContent", func(t *testing.T) {
		// Good proof //
		const term = 5
		const view = 0
		const targetTerm = term
		const targetView = view + 1

		leaderMF := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, leaderKeyManager)
		node1MF := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, node1KeyManager)
		node2MF := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, node2KeyManager)

		// TODO Maybe can use node1MsgFactory instead of creating node1MF here (same for leader and node2)
		// Good proof //
		goodPrepareProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*leanhelix.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node2MF.CreatePrepareMessage(term, view, block),
			},
		}
		actualGood := leanhelix.ValidatePreparedProof(targetTerm, targetView, goodPrepareProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, actualGood, "Did not approve a valid proof")

		// Mismatching term //
		badTermProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*leanhelix.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node2MF.CreatePrepareMessage(666, view, block),
			},
		}
		actualBadTerm := leanhelix.ValidatePreparedProof(targetTerm, targetView, badTermProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadTerm, "Did not reject mismatching term")

		// Mismatching view //
		badViewProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*leanhelix.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node2MF.CreatePrepareMessage(term, 666, block),
			},
		}
		actualBadView := leanhelix.ValidatePreparedProof(targetTerm, targetView, badViewProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadView, "Did not reject mismatching view")

		// Mismatching blockHash //
		otherBlock := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
		badBlockHashProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*leanhelix.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node2MF.CreatePrepareMessage(term, view, otherBlock),
			},
		}
		actualBadBlockHash := leanhelix.ValidatePreparedProof(targetTerm, targetView, badBlockHashProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadBlockHash, "Did not reject mismatching block hash")
	})

	t.Run("TestProofsValidatorWithDuplicate prepare sender PK", func(t *testing.T) {
		leaderMF := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, leaderKeyManager)
		node1MF := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, node1KeyManager)

		preparedProof := &leanhelix.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*leanhelix.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node1MF.CreatePrepareMessage(term, view, block),
			},
		}

		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with duplicate sender PK")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a valid proof")
	})
}
