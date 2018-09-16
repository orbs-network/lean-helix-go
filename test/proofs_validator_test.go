package test

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProofsValidator(t *testing.T) {
	keyManager := builders.NewMockKeyManager(lh.PublicKey("Dummy PK"))
	leaderKeyManager := builders.NewMockKeyManager(lh.PublicKey("Leader PK"))
	node1KeyManager := builders.NewMockKeyManager(lh.PublicKey("Node 1"))
	node2KeyManager := builders.NewMockKeyManager(lh.PublicKey("Node 2"))

	membersPKs := []lh.PublicKey{lh.PublicKey("Leader PK"), lh.PublicKey("Node 1"), lh.PublicKey("Node 2"), lh.PublicKey("Node 3")}
	calcLeaderPk := func(view lh.ViewCounter) lh.PublicKey {
		return membersPKs[view]
	}

	const f = 1
	const term = 0
	const view = 0
	const targetTerm = term
	const targetView = view + 1
	block := builders.CreateBlock(builders.GenesisBlock)
	leaderMsgFactory := builders.NewMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
	node1MsgFactory := builders.NewMessageFactory(builders.CalculateBlockHash, node1KeyManager)
	node2MsgFactory := builders.NewMessageFactory(builders.CalculateBlockHash, node2KeyManager)

	preprepareMessage := leaderMsgFactory.CreatePreprepareMessage(term, view, block)
	prepareMessage1 := node1MsgFactory.CreatePrepareMessage(term, view, block)
	prepareMessage2 := node2MsgFactory.CreatePrepareMessage(term, view, block)
	preparedProof := lh.CreatePreparedProof(preprepareMessage, []lh.PrepareMessage{prepareMessage1, prepareMessage2})

	t.Run("TestProofsValidatorWithNoPrePrepare", func(t *testing.T) {
		preparedProof := lh.CreatePreparedProof(nil, []lh.PrepareMessage{prepareMessage1, prepareMessage2})
		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have a preprepare message")
	})

	t.Run("TestProofsValidatorWithNoPrepares", func(t *testing.T) {
		preparedProof := lh.CreatePreparedProof(preprepareMessage, nil)

		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have prepare messages")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := lh.ValidatePreparedProof(targetTerm, targetView, nil, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a nil proof")
	})

	t.Run("TestProofsValidatorWithBadPreprepareSignature", func(t *testing.T) {
		keyManager := builders.NewMockKeyManager(lh.PublicKey("Dummy PK"), lh.PublicKey("Leader PK"))
		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass preprepare signature validation")
	})

	t.Run("TestProofsValidatorWithBadPrepareSignature", func(t *testing.T) {
		keyManager := builders.NewMockKeyManager(lh.PublicKey("Dummy PK"), lh.PublicKey("Node 2"))
		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass prepare signature validation")
	})

	t.Run("TestProofsValidatorWithNotEnoughPrepareMessages", func(t *testing.T) {
		preparedProof := lh.CreatePreparedProof(preprepareMessage, []lh.PrepareMessage{prepareMessage1})
		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with not enough prepares")
	})

	t.Run("TestProofsValidatorWithTerm", func(t *testing.T) {
		result := lh.ValidatePreparedProof(666, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with mismatching term")
	})

	t.Run("TestProofsValidatorWithTheSameView", func(t *testing.T) {
		result := lh.ValidatePreparedProof(targetTerm, view, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with equal targetView")
	})

	t.Run("TestProofsValidatorWithTheSmallerView", func(t *testing.T) {
		result := lh.ValidatePreparedProof(targetTerm, targetView-1, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with smaller targetView")
	})

	t.Run("TestProofsValidatorWithANoneMember", func(t *testing.T) {
		noneMemberKeyManager := builders.NewMockKeyManager(lh.PublicKey("Not in members PK"))
		mf := builders.NewMessageFactory(builders.CalculateBlockHash, noneMemberKeyManager)
		prepareMessage1 := mf.CreatePrepareMessage(term, view, block)
		preparedProof := lh.CreatePreparedProof(preprepareMessage, []lh.PrepareMessage{prepareMessage1, prepareMessage2})
		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a none member")
	})

	t.Run("TestProofsValidatorWithPrepareFromTheLeader", func(t *testing.T) {
		mf := builders.NewMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
		prepareMessage1 := mf.CreatePrepareMessage(term, view, block)
		preparedProof := lh.CreatePreparedProof(preprepareMessage, []lh.PrepareMessage{prepareMessage1, prepareMessage2})
		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a prepare from the leader")
	})

	t.Run("TestProofsValidatorWithMismatchingViewToLeader", func(t *testing.T) {
		calcLeaderPk := func(view lh.ViewCounter) lh.PublicKey {
			return lh.PublicKey("Some other node PK")
		}
		preparedProof := lh.CreatePreparedProof(preprepareMessage, []lh.PrepareMessage{prepareMessage1, prepareMessage2})
		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a mismatching view to leader")
	})

	t.Run("TestProofsValidatorWithMismatchingContent", func(t *testing.T) {
		// Good proof //
		const term = 5
		const view = 0
		const targetTerm = term
		const targetView = view + 1

		leaderMF := builders.NewMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
		node1MF := builders.NewMessageFactory(builders.CalculateBlockHash, node1KeyManager)
		node2MF := builders.NewMessageFactory(builders.CalculateBlockHash, node2KeyManager)

		// TODO Maybe can use node1MsgFactory instead of creating node1MF here (same for leader and node2)
		// Good proof //
		goodPrepareProof := lh.CreatePreparedProof(leaderMF.CreatePreprepareMessage(term, view, block), []lh.PrepareMessage{
			node1MF.CreatePrepareMessage(term, view, block),
			node2MF.CreatePrepareMessage(term, view, block),
		})
		actualGood := lh.ValidatePreparedProof(targetTerm, targetView, goodPrepareProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, actualGood, "Did not approve a valid proof")

		// Mismatching term //
		badTermProof := lh.CreatePreparedProof(leaderMF.CreatePreprepareMessage(term, view, block), []lh.PrepareMessage{
			node1MF.CreatePrepareMessage(term, view, block),
			node2MF.CreatePrepareMessage(666, view, block),
		})

		actualBadTerm := lh.ValidatePreparedProof(targetTerm, targetView, badTermProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadTerm, "Did not reject mismatching term")

		// Mismatching view //
		badViewProof := lh.CreatePreparedProof(leaderMF.CreatePreprepareMessage(term, view, block), []lh.PrepareMessage{
			node1MF.CreatePrepareMessage(term, view, block),
			node2MF.CreatePrepareMessage(term, 666, block),
		})
		actualBadView := lh.ValidatePreparedProof(targetTerm, targetView, badViewProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadView, "Did not reject mismatching view")

		// Mismatching blockHash //
		otherBlock := builders.CreateBlock(builders.GenesisBlock)
		badBlockHashProof := lh.CreatePreparedProof(leaderMF.CreatePreprepareMessage(term, view, block), []lh.PrepareMessage{
			node1MF.CreatePrepareMessage(term, view, block),
			node2MF.CreatePrepareMessage(term, view, otherBlock),
		})

		actualBadBlockHash := lh.ValidatePreparedProof(targetTerm, targetView, badBlockHashProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadBlockHash, "Did not reject mismatching block hash")
	})

	t.Run("TestProofsValidatorWithDuplicate prepare sender PK", func(t *testing.T) {
		leaderMF := builders.NewMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
		node1MF := builders.NewMessageFactory(builders.CalculateBlockHash, node1KeyManager)

		preparedProof := lh.CreatePreparedProof(leaderMF.CreatePreprepareMessage(term, view, block), []lh.PrepareMessage{
			node1MF.CreatePrepareMessage(term, view, block),
			node1MF.CreatePrepareMessage(term, view, block),
		})
		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with duplicate sender PK")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := lh.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a valid proof")
	})
}
