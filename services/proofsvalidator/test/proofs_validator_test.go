package test

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestProofsValidator(t *testing.T) {
	myMemberId := primitives.MemberId("My MemberId")
	leaderId := primitives.MemberId("Leader ID")
	node1Id := primitives.MemberId("Node 1")
	node2Id := primitives.MemberId("Node 2")
	node3Id := primitives.MemberId("Node 3")
	myKeyManager := mocks.NewMockKeyManager(myMemberId)
	leaderKeyManager := mocks.NewMockKeyManager(leaderId)
	node1KeyManager := mocks.NewMockKeyManager(node1Id)
	node2KeyManager := mocks.NewMockKeyManager(node2Id)
	node3KeyManager := mocks.NewMockKeyManager(node3Id)

	membersIds := []primitives.MemberId{leaderId, node1Id, node2Id, node3Id}

	nodeCount := 4
	f := int(math.Floor(float64(nodeCount) / 3))
	q := nodeCount - f
	const blockHeight = 0
	const view = 0
	const targetBlockHeight = blockHeight
	const targetView = view + 1
	block := mocks.ABlock(interfaces.GenesisBlock)
	blockHash := mocks.CalculateBlockHash(block)

	leaderMessageSigner := &builders.MessageSigner{KeyManager: leaderKeyManager, MemberId: leaderId}
	nodesMessageSigners := []*builders.MessageSigner{
		{KeyManager: node1KeyManager, MemberId: node1Id},
		{KeyManager: node2KeyManager, MemberId: node2Id},
		{KeyManager: node3KeyManager, MemberId: node3Id},
	}
	goodPrepareProof := builders.CreatePreparedProof(leaderMessageSigner, nodesMessageSigners, blockHeight, view, blockHash)

	calcLeaderId := func(view primitives.View) primitives.MemberId {
		return membersIds[view]
	}

	t.Run("TestProofsValidatorHappyPath", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, myKeyManager, membersIds, calcLeaderId)
		require.True(t, result, "Did not approve a well-formed proof")
	})

	t.Run("TestProofsValidatorWithNoPrePrepare", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: node1KeyManager, MemberId: node1Id},
			{KeyManager: node2KeyManager, MemberId: node2Id},
			{KeyManager: node3KeyManager, MemberId: node3Id},
		}
		preparedProofWithoutPP := builders.CreatePreparedProof(nil, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithoutPP, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof that did not have a preprepare message")
	})

	t.Run("TestProofsValidatorWithNoPrepares", func(t *testing.T) {
		preparedProofWithoutP := builders.CreatePreparedProof(leaderMessageSigner, nil, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithoutP, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof that did not have prepare messages")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, nil, q, myKeyManager, membersIds, calcLeaderId)
		require.True(t, result, "Did not approve a nil proof")
	})

	t.Run("TestProofsValidatorWithNotEnoughPrepareMessages", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: node1KeyManager, MemberId: node1Id},
		}
		preparedProofWithNotEnoughP := builders.CreatePreparedProof(leaderMessageSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithNotEnoughP, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof with not enough prepares")
	})

	t.Run("TestProofsValidatorWithBadPreprepareSignature", func(t *testing.T) {
		rejectingKeyManager := mocks.NewMockKeyManager(myMemberId, leaderId)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, rejectingKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof that did not pass preprepare signature validation")
	})

	t.Run("TestProofsValidatorWithBadPrepareSignature", func(t *testing.T) {
		rejectingKeyManager := mocks.NewMockKeyManager(myMemberId, node2Id)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, rejectingKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof that did not pass prepare signature validation")
	})

	t.Run("TestProofsValidatorWithMismatchedHeight", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(666, targetView, goodPrepareProof, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof with mismatching blockHeight")
	})

	t.Run("TestProofsValidatorWithTheSameView", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, view, goodPrepareProof, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof with equal targetView")
	})

	t.Run("TestProofsValidatorWithTheSmallerView", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView-1, goodPrepareProof, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof with smaller targetView")
	})

	t.Run("TestProofsValidatorWithANonMember", func(t *testing.T) {
		nonMemberId := primitives.MemberId("Not in members Ids")
		nonMemberKeyManager := mocks.NewMockKeyManager(nonMemberId)
		pSigners := []*builders.MessageSigner{
			{KeyManager: node1KeyManager, MemberId: node1Id},
			{KeyManager: node2KeyManager, MemberId: node2Id},
			{KeyManager: nonMemberKeyManager, MemberId: nonMemberId},
		}
		preparedProof := builders.CreatePreparedProof(leaderMessageSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProof, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof with a none member")
	})

	t.Run("TestProofsValidatorWithPrepareFromTheLeader", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: leaderKeyManager, MemberId: leaderId},
			{KeyManager: node1KeyManager, MemberId: node1Id},
			{KeyManager: node2KeyManager, MemberId: node2Id},
		}
		ppSigner := &builders.MessageSigner{KeyManager: node1KeyManager, MemberId: node1Id}

		preparedProofWithPFromLeader := builders.CreatePreparedProof(ppSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithPFromLeader, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof with a prepare from the leader")
	})

	t.Run("TestProofsValidatorWithMismatchingViewToLeader", func(t *testing.T) {
		calcLeaderId := func(view primitives.View) primitives.MemberId {
			return primitives.MemberId("Some other node Id")
		}
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof with a mismatching view to leader")

	})

	t.Run("TestProofsValidatorWithMismatchingContent", func(t *testing.T) {
		// Good proof //
		const blockHeight = 5
		const view = 0
		const targetBlockHeight = blockHeight
		const targetView = view + 1

		goodPrepareProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(node1KeyManager, node1Id, blockHeight, view, block),
				builders.APrepareMessage(node2KeyManager, node2Id, blockHeight, view, block),
				builders.APrepareMessage(node3KeyManager, node3Id, blockHeight, view, block),
			})

		actualGood := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, q, myKeyManager, membersIds, calcLeaderId)
		require.True(t, actualGood, "Did not approve a valid proof")

		// Mismatching blockHeight //
		badHeightProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(node1KeyManager, node1Id, blockHeight, view, block),
				builders.APrepareMessage(node2KeyManager, node2Id, 666, view, block),
				builders.APrepareMessage(node3KeyManager, node3Id, blockHeight, view, block),
			})

		actualBadHeight := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, badHeightProof, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, actualBadHeight, "Did not reject mismatching blockHeight")

		// Mismatching view //
		badViewProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(node1KeyManager, node1Id, blockHeight, view, block),
				builders.APrepareMessage(node2KeyManager, node2Id, blockHeight, 666, block),
				builders.APrepareMessage(node3KeyManager, node3Id, blockHeight, view, block),
			})

		actualBadView := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, badViewProof, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, actualBadView, "Did not reject mismatching view")

		// Mismatching blockHash //
		otherBlock := mocks.ABlock(interfaces.GenesisBlock)
		badBlockHashProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(node1KeyManager, node1Id, blockHeight, view, block),
				builders.APrepareMessage(node2KeyManager, node2Id, blockHeight, view, otherBlock),
				builders.APrepareMessage(node3KeyManager, node3Id, blockHeight, view, block),
			})

		actualBadBlockHash := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, badBlockHashProof, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, actualBadBlockHash, "Did not reject mismatching block hash")
	})

	t.Run("TestProofsValidatorWithDuplicate prepare sender Id", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: node1KeyManager, MemberId: node1Id},
			{KeyManager: node2KeyManager, MemberId: node2Id},
			{KeyManager: node1KeyManager, MemberId: node1Id},
		}
		preparedProofWithDuplicatePSenderId := builders.CreatePreparedProof(leaderMessageSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithDuplicatePSenderId, q, myKeyManager, membersIds, calcLeaderId)
		require.False(t, result, "Did not reject a proof with duplicate sender Id")
	})
}