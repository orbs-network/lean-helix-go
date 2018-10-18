package test

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProofsValidator(t *testing.T) {
	dummyKeyManager := builders.NewMockKeyManager(Ed25519PublicKey("Dummy PK"))
	leaderKeyManager := builders.NewMockKeyManager(Ed25519PublicKey("Leader PK"))
	node1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey("Node 1"))
	node2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey("Node 2"))
	badSignerKeyManager := builders.NewMockKeyManager(Ed25519PublicKey("Node 2"))
	badSignerKeyManager.When("Sign").Return("") // Return empty signature

	membersPKs := []Ed25519PublicKey{Ed25519PublicKey("Leader PK"), Ed25519PublicKey("Node 1"), Ed25519PublicKey("Node 2"), Ed25519PublicKey("Node 3")}
	calcLeaderPk := func(view View) Ed25519PublicKey {
		return membersPKs[view]
	}

	const f = 1
	const height = 0
	const view = 0
	const targetHeight = height
	const targetView = view + 1
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	goodPreparedProof := lh.CreatePreparedProof(leaderKeyManager, []lh.KeyManager{node1KeyManager, node2KeyManager}, height, view, blockHash)

	t.Run("TestProofsValidatorHappyPath", func(t *testing.T) {
		dummyKeyManager.When("Verify", mock.Any, mock.Any).Return(true)
		result := lh.ValidatePreparedProof(targetHeight, targetView, goodPreparedProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a well-formed proof")
	})

	t.Run("TestProofsValidatorWithNoPrePrepare", func(t *testing.T) {
		preparedProofWithoutPP := lh.CreatePreparedProof(nil, []lh.KeyManager{node1KeyManager, node2KeyManager}, height, view, blockHash)
		result := lh.ValidatePreparedProof(targetHeight, targetView, preparedProofWithoutPP, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have a preprepare message")
	})

	t.Run("TestProofsValidatorWithNoPrepares", func(t *testing.T) {
		preparedProofWithoutP := lh.CreatePreparedProof(leaderKeyManager, nil, height, view, blockHash)

		result := lh.ValidatePreparedProof(targetHeight, targetView, preparedProofWithoutP, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have prepare messages")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := lh.ValidatePreparedProof(targetHeight, targetView, nil, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a nil proof")
	})

	t.Run("TestProofsValidatorWithBadPreprepareSignature", func(t *testing.T) {
		preparedProofWithBadPPSig := lh.CreatePreparedProof(badSignerKeyManager, []lh.KeyManager{node1KeyManager, node2KeyManager}, height, view, blockHash)
		result := lh.ValidatePreparedProof(targetHeight, targetView, preparedProofWithBadPPSig, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass preprepare signature validation")
	})

	t.Run("TestProofsValidatorWithBadPrepareSignature", func(t *testing.T) {
		preparedProofWithBadPSig := lh.CreatePreparedProof(leaderKeyManager, []lh.KeyManager{badSignerKeyManager, node2KeyManager}, height, view, blockHash)
		result := lh.ValidatePreparedProof(targetHeight, targetView, preparedProofWithBadPSig, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass prepare signature validation")
	})

	t.Run("TestProofsValidatorWithNotEnoughPrepareMessages", func(t *testing.T) {
		preparedProofWithNotEnoughP := lh.CreatePreparedProof(leaderKeyManager, []lh.KeyManager{node1KeyManager}, height, view, blockHash)
		result := lh.ValidatePreparedProof(targetHeight, targetView, preparedProofWithNotEnoughP, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with not enough prepares")
	})

	t.Run("TestProofsValidatorWithMismatchedHeight", func(t *testing.T) {
		result := lh.ValidatePreparedProof(666, targetView, goodPreparedProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with mismatching height")
	})

	t.Run("TestProofsValidatorWithTheSameView", func(t *testing.T) {
		result := lh.ValidatePreparedProof(targetHeight, view, goodPreparedProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with equal targetView")
	})

	t.Run("TestProofsValidatorWithTheSmallerView", func(t *testing.T) {
		result := lh.ValidatePreparedProof(targetHeight, targetView-1, goodPreparedProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with smaller targetView")
	})

	t.Run("TestProofsValidatorWithANonMember", func(t *testing.T) {
		nonMemberKeyManager := builders.NewMockKeyManager(Ed25519PublicKey("Not in members PK"))
		preparedProof := lh.CreatePreparedProof(leaderKeyManager, []lh.KeyManager{node1KeyManager, nonMemberKeyManager}, height, view, blockHash)
		result := lh.ValidatePreparedProof(targetHeight, targetView, preparedProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a none member")
	})

	t.Run("TestProofsValidatorWithPrepareFromTheLeader", func(t *testing.T) {
		preparedProofWithPFromLeader := lh.CreatePreparedProof(node1KeyManager, []lh.KeyManager{leaderKeyManager, node2KeyManager}, height, view, blockHash)
		result := lh.ValidatePreparedProof(targetHeight, targetView, preparedProofWithPFromLeader, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a prepare from the leader")
	})

	t.Run("TestProofsValidatorWithMismatchingViewToLeader", func(t *testing.T) {
		calcLeaderPk := func(view View) Ed25519PublicKey {
			return Ed25519PublicKey("Some other node PK")
		}
		result := lh.ValidatePreparedProof(targetHeight, targetView, goodPreparedProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a mismatching view to leader")

	})

	//t.Run("TestProofsValidatorWithMismatchingContent", func(t *testing.T) {
	//	// Good proof //
	//	const term = 5
	//	const view = 0
	//	const targetTerm = term
	//	const targetView = view + 1
	//
	//	leaderMF := builders.NewMockMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
	//	node1MF := builders.NewMockMessageFactory(builders.CalculateBlockHash, node1KeyManager)
	//	node2MF := builders.NewMockMessageFactory(builders.CalculateBlockHash, node2KeyManager)
	//
	//	// TODO Maybe can use node1MsgFactory instead of creating node1MF here (same for leader and node2)
	//	// Good proof //
	//	goodPrepareProof := lh.CreatePreparedProof(leaderMF.CreatePreprepareMessage(term, view, block), []lh.PrepareMessage{
	//		node1MF.CreatePrepareMessage(term, view, block),
	//		node2MF.CreatePrepareMessage(term, view, block),
	//	})
	//	actualGood := lh.ValidatePreparedProof(targetTerm, targetView, goodPrepareProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
	//	require.True(t, actualGood, "Did not approve a valid proof")
	//
	//	// Mismatching height //
	//	badTermProof := lh.CreatePreparedProof(leaderMF.CreatePreprepareMessage(term, view, block), []lh.PrepareMessage{
	//		node1MF.CreatePrepareMessage(term, view, block),
	//		node2MF.CreatePrepareMessage(666, view, block),
	//	})
	//
	//	actualBadTerm := lh.ValidatePreparedProof(targetTerm, targetView, badTermProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
	//	require.False(t, actualBadTerm, "Did not reject mismatching height")
	//
	//	// Mismatching view //
	//	badViewProof := lh.CreatePreparedProof(leaderMF.CreatePreprepareMessage(term, view, block), []lh.PrepareMessage{
	//		node1MF.CreatePrepareMessage(term, view, block),
	//		node2MF.CreatePrepareMessage(term, 666, block),
	//	})
	//	actualBadView := lh.ValidatePreparedProof(targetTerm, targetView, badViewProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
	//	require.False(t, actualBadView, "Did not reject mismatching view")
	//
	//	// Mismatching blockHash //
	//	otherBlock := builders.CreateBlock(builders.GenesisBlock)
	//	badBlockHashProof := lh.CreatePreparedProof(leaderMF.CreatePreprepareMessage(term, view, block), []lh.PrepareMessage{
	//		node1MF.CreatePrepareMessage(term, view, block),
	//		node2MF.CreatePrepareMessage(term, view, otherBlock),
	//	})
	//
	//	actualBadBlockHash := lh.ValidatePreparedProof(targetTerm, targetView, badBlockHashProof, f, dummyKeyManager, membersPKs, calcLeaderPk)
	//	require.False(t, actualBadBlockHash, "Did not reject mismatching block hash")
	//})

	t.Run("TestProofsValidatorWithDuplicate prepare sender PK", func(t *testing.T) {
		preparedProofWithDuplicatePSenderPK := lh.CreatePreparedProof(leaderKeyManager, []lh.KeyManager{node1KeyManager, node1KeyManager}, height, view, blockHash)
		result := lh.ValidatePreparedProof(targetHeight, targetView, preparedProofWithDuplicatePSenderPK, f, dummyKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with duplicate sender PK")
	})
}
