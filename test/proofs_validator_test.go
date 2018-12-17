package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestProofsValidator(t *testing.T) {
	myPK := primitives.MemberId("My MemberId")
	leaderPK := primitives.MemberId("Leader PK")
	node1PK := primitives.MemberId("Node 1")
	node2PK := primitives.MemberId("Node 2")
	node3PK := primitives.MemberId("Node 3")
	myKeyManager := builders.NewMockKeyManager(myPK)
	leaderKeyManager := builders.NewMockKeyManager(leaderPK)
	node1KeyManager := builders.NewMockKeyManager(node1PK)
	node2KeyManager := builders.NewMockKeyManager(node2PK)
	node3KeyManager := builders.NewMockKeyManager(node3PK)

	membersPKs := []primitives.MemberId{leaderPK, node1PK, node2PK, node3PK}

	nodeCount := 4
	f := int(math.Floor(float64(nodeCount) / 3))
	q := nodeCount - f
	const blockHeight = 0
	const view = 0
	const targetBlockHeight = blockHeight
	const targetView = view + 1
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	goodPrepareProof := builders.CreatePreparedProof(leaderKeyManager, []leanhelix.KeyManager{node1KeyManager, node2KeyManager, node3KeyManager}, blockHeight, view, blockHash)

	calcLeaderPk := func(view primitives.View) primitives.MemberId {
		return membersPKs[view]
	}

	t.Run("TestProofsValidatorHappyPath", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a well-formed proof")
	})

	t.Run("TestProofsValidatorWithNoPrePrepare", func(t *testing.T) {
		preparedProofWithoutPP := builders.CreatePreparedProof(nil, []leanhelix.KeyManager{node1KeyManager, node2KeyManager, node3KeyManager}, blockHeight, view, blockHash)
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithoutPP, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have a preprepare message")
	})

	t.Run("TestProofsValidatorWithNoPrepares", func(t *testing.T) {
		preparedProofWithoutP := builders.CreatePreparedProof(leaderKeyManager, nil, blockHeight, view, blockHash)
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithoutP, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have prepare messages")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, nil, q, myKeyManager, membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a nil proof")
	})

	t.Run("TestProofsValidatorWithNotEnoughPrepareMessages", func(t *testing.T) {
		preparedProofWithNotEnoughP := builders.CreatePreparedProof(leaderKeyManager, []leanhelix.KeyManager{node1KeyManager}, blockHeight, view, blockHash)
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithNotEnoughP, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with not enough prepares")
	})

	t.Run("TestProofsValidatorWithBadPreprepareSignature", func(t *testing.T) {
		rejectingKeyManager := builders.NewMockKeyManager(myPK, leaderPK)
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, rejectingKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass preprepare signature validation")
	})

	t.Run("TestProofsValidatorWithBadPrepareSignature", func(t *testing.T) {
		rejectingKeyManager := builders.NewMockKeyManager(myPK, node2PK)
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, rejectingKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass prepare signature validation")
	})

	t.Run("TestProofsValidatorWithMismatchedHeight", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(666, targetView, goodPrepareProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with mismatching blockHeight")
	})

	t.Run("TestProofsValidatorWithTheSameView", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, view, goodPrepareProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with equal targetView")
	})

	t.Run("TestProofsValidatorWithTheSmallerView", func(t *testing.T) {
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView-1, goodPrepareProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with smaller targetView")
	})

	t.Run("TestProofsValidatorWithANonMember", func(t *testing.T) {
		nonMemberKeyManager := builders.NewMockKeyManager(primitives.MemberId("Not in members PK"))
		preparedProof := builders.CreatePreparedProof(leaderKeyManager, []leanhelix.KeyManager{node1KeyManager, node2KeyManager, nonMemberKeyManager}, blockHeight, view, blockHash)
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, preparedProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a none member")
	})

	t.Run("TestProofsValidatorWithPrepareFromTheLeader", func(t *testing.T) {
		preparedProofWithPFromLeader := builders.CreatePreparedProof(node1KeyManager, []leanhelix.KeyManager{leaderKeyManager, node1KeyManager, node2KeyManager}, blockHeight, view, blockHash)
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithPFromLeader, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a prepare from the leader")
	})

	t.Run("TestProofsValidatorWithMismatchingViewToLeader", func(t *testing.T) {
		calcLeaderPk := func(view primitives.View) primitives.MemberId {
			return primitives.MemberId("Some other node PK")
		}
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a mismatching view to leader")

	})

	t.Run("TestProofsValidatorWithMismatchingContent", func(t *testing.T) {
		// Good proof //
		const blockHeight = 5
		const view = 0
		const targetBlockHeight = blockHeight
		const targetView = view + 1

		goodPrepareProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(leaderKeyManager, blockHeight, view, block),
			[]*leanhelix.PrepareMessage{
				builders.APrepareMessage(node1KeyManager, blockHeight, view, block),
				builders.APrepareMessage(node2KeyManager, blockHeight, view, block),
				builders.APrepareMessage(node3KeyManager, blockHeight, view, block),
			})

		actualGood := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.True(t, actualGood, "Did not approve a valid proof")

		// Mismatching blockHeight //
		badHeightProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(leaderKeyManager, blockHeight, view, block),
			[]*leanhelix.PrepareMessage{
				builders.APrepareMessage(node1KeyManager, blockHeight, view, block),
				builders.APrepareMessage(node2KeyManager, 666, view, block),
				builders.APrepareMessage(node3KeyManager, blockHeight, view, block),
			})

		actualBadHeight := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, badHeightProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, actualBadHeight, "Did not reject mismatching blockHeight")

		// Mismatching view //
		badViewProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(leaderKeyManager, blockHeight, view, block),
			[]*leanhelix.PrepareMessage{
				builders.APrepareMessage(node1KeyManager, blockHeight, view, block),
				builders.APrepareMessage(node2KeyManager, blockHeight, 666, block),
				builders.APrepareMessage(node3KeyManager, blockHeight, view, block),
			})

		actualBadView := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, badViewProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, actualBadView, "Did not reject mismatching view")

		// Mismatching blockHash //
		otherBlock := builders.CreateBlock(builders.GenesisBlock)
		badBlockHashProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(leaderKeyManager, blockHeight, view, block),
			[]*leanhelix.PrepareMessage{
				builders.APrepareMessage(node1KeyManager, blockHeight, view, block),
				builders.APrepareMessage(node2KeyManager, blockHeight, view, otherBlock),
				builders.APrepareMessage(node3KeyManager, blockHeight, view, block),
			})

		actualBadBlockHash := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, badBlockHashProof, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, actualBadBlockHash, "Did not reject mismatching block hash")
	})

	t.Run("TestProofsValidatorWithDuplicate prepare sender PK", func(t *testing.T) {
		preparedProofWithDuplicatePSenderPK := builders.CreatePreparedProof(leaderKeyManager, []leanhelix.KeyManager{node1KeyManager, node1KeyManager, node2KeyManager}, blockHeight, view, blockHash)
		result := leanhelix.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithDuplicatePSenderPK, q, myKeyManager, membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with duplicate sender PK")
	})
}
