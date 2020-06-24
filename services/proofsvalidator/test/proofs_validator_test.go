// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/testhelpers"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestProofsValidator(t *testing.T) {
	instanceId := primitives.InstanceId(rand.Uint64())
	myMemberId := primitives.MemberId("My MemberId")
	leaderIdW1 := primitives.MemberId("Leader ID (W1)")
	nodeIdW2 := primitives.MemberId("Node 1 (W2)")
	nodeIdW3 := primitives.MemberId("Node 2 (W3)")
	nodeIdW4 := primitives.MemberId("Node 3 (W4)")
	myKeyManager := mocks.NewMockKeyManager(myMemberId)
	leaderW1KeyManager := mocks.NewMockKeyManager(leaderIdW1)
	nodeW2KeyManager := mocks.NewMockKeyManager(nodeIdW2)
	nodeW3KeyManager := mocks.NewMockKeyManager(nodeIdW3)
	nodeW4KeyManager := mocks.NewMockKeyManager(nodeIdW4)

	membersIds := []primitives.MemberId{leaderIdW1, nodeIdW2, nodeIdW3, nodeIdW4}
	weights := []primitives.MemberWeight{1, 2, 3, 4}
	committeeMembers := testhelpers.GenMembersWithWeights(membersIds, weights)

	require.Equal(t, uint(7), quorum.CalcQuorumWeight(quorum.GetWeights(committeeMembers)))

	const blockHeight = 0
	const view = 0
	const targetBlockHeight = blockHeight
	const targetView = view + 1
	block := mocks.ABlock(interfaces.GenesisBlock)
	blockHash := mocks.CalculateBlockHash(block)

	leaderMessageSigner := &builders.MessageSigner{KeyManager: leaderW1KeyManager, MemberId: leaderIdW1}
	nodesMessageSigners := []*builders.MessageSigner{
		{KeyManager: nodeW2KeyManager, MemberId: nodeIdW2},
		{KeyManager: nodeW3KeyManager, MemberId: nodeIdW3},
		{KeyManager: nodeW4KeyManager, MemberId: nodeIdW4},
	}
	goodPrepareProof := builders.CreatePreparedProof(instanceId, leaderMessageSigner, nodesMessageSigners, blockHeight, view, blockHash)

	calcLeaderId := func(view primitives.View) primitives.MemberId {
		return membersIds[view]
	}

	t.Run("TestProofsValidatorHappyPath", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, myKeyManager, committeeMembers, calcLeaderId)
		require.True(t, result, "Did not approve a well-formed proof")
	})

	t.Run("TestProofsValidatorWithNoPrePrepare", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: nodeW2KeyManager, MemberId: nodeIdW2},
			{KeyManager: nodeW3KeyManager, MemberId: nodeIdW3},
			{KeyManager: nodeW4KeyManager, MemberId: nodeIdW4},
		}
		preparedProofWithoutPP := builders.CreatePreparedProof(instanceId, nil, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithoutPP, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof that did not have a preprepare message")
	})

	t.Run("TestProofsValidatorWithNoPrepares", func(t *testing.T) {
		preparedProofWithoutP := builders.CreatePreparedProof(instanceId, leaderMessageSigner, nil, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithoutP, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof that did not have prepare messages")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, nil, myKeyManager, committeeMembers, calcLeaderId)
		require.True(t, result, "Did not approve a nil proof")
	})

	t.Run("TestProofsValidatorWithPrepareMessagesThatHaventReachedQuorum", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: nodeW2KeyManager, MemberId: nodeIdW2},
			{KeyManager: nodeW3KeyManager, MemberId: nodeIdW3},
		}

		preparedProofWithNotEnoughP := builders.CreatePreparedProof(instanceId, leaderMessageSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithNotEnoughP, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with not enough prepares")
	})

	t.Run("TestProofsValidatorWithBadPreprepareSignature", func(t *testing.T) {
		rejectingKeyManager := mocks.NewMockKeyManager(myMemberId, leaderIdW1)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, rejectingKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof that did not pass preprepare signature validation")
	})

	t.Run("TestProofsValidatorWithBadPrepareSignature", func(t *testing.T) {
		rejectingKeyManager := mocks.NewMockKeyManager(myMemberId, nodeIdW3)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, rejectingKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof that did not pass prepare signature validation")
	})

	t.Run("TestProofsValidatorWithMismatchedHeight", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(666, targetView, goodPrepareProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with mismatching blockHeight")
	})

	t.Run("TestProofsValidatorWithTheSameView", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, view, goodPrepareProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with equal targetView")
	})

	t.Run("TestProofsValidatorWithTheSmallerView", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView-1, goodPrepareProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with smaller targetView")
	})

	t.Run("TestProofsValidatorWithANonMember", func(t *testing.T) {
		nonMemberId := primitives.MemberId("Not in members Ids")
		nonMemberKeyManager := mocks.NewMockKeyManager(nonMemberId)
		pSigners := []*builders.MessageSigner{
			{KeyManager: nodeW2KeyManager, MemberId: nodeIdW2},
			{KeyManager: nodeW3KeyManager, MemberId: nodeIdW3},
			{KeyManager: nonMemberKeyManager, MemberId: nonMemberId},
		}
		preparedProof := builders.CreatePreparedProof(instanceId, leaderMessageSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with a none member")
	})

	t.Run("TestProofsValidatorWithPrepareFromTheLeader", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: leaderW1KeyManager, MemberId: leaderIdW1},
			{KeyManager: nodeW2KeyManager, MemberId: nodeIdW2},
			{KeyManager: nodeW3KeyManager, MemberId: nodeIdW3},
		}
		ppSigner := &builders.MessageSigner{KeyManager: nodeW2KeyManager, MemberId: nodeIdW2}

		preparedProofWithPFromLeader := builders.CreatePreparedProof(instanceId, ppSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithPFromLeader, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with a prepare from the leader")
	})

	t.Run("TestProofsValidatorWithMismatchingViewToLeader", func(t *testing.T) {
		calcLeaderId := func(view primitives.View) primitives.MemberId {
			return primitives.MemberId("Some other node Id")
		}
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with a mismatching view to leader")

	})

	t.Run("TestProofsValidatorWithMismatchingContent", func(t *testing.T) {
		// Good proof //
		instanceId := primitives.InstanceId(rand.Uint64())
		const blockHeight = 5
		const view = 0
		const targetBlockHeight = blockHeight
		const targetView = view + 1

		goodPrepareProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(instanceId, nodeW2KeyManager, nodeIdW2, blockHeight, view, block),
				builders.APrepareMessage(instanceId, nodeW3KeyManager, nodeIdW3, blockHeight, view, block),
				builders.APrepareMessage(instanceId, nodeW4KeyManager, nodeIdW4, blockHeight, view, block),
			})

		actualGood := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, myKeyManager, committeeMembers, calcLeaderId)
		require.True(t, actualGood, "Did not approve a valid proof")

		// Mismatching blockHeight //
		badHeightProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(instanceId, nodeW2KeyManager, nodeIdW2, blockHeight, view, block),
				builders.APrepareMessage(instanceId, nodeW3KeyManager, nodeIdW3, 666, view, block),
				builders.APrepareMessage(instanceId, nodeW4KeyManager, nodeIdW4, blockHeight, view, block),
			})

		actualBadHeight := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, badHeightProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, actualBadHeight, "Did not reject mismatching blockHeight")

		// Mismatching view //
		badViewProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(instanceId, nodeW2KeyManager, nodeIdW2, blockHeight, view, block),
				builders.APrepareMessage(instanceId, nodeW3KeyManager, nodeIdW3, blockHeight, 666, block),
				builders.APrepareMessage(instanceId, nodeW4KeyManager, nodeIdW4, blockHeight, view, block),
			})

		actualBadView := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, badViewProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, actualBadView, "Did not reject mismatching view")

		// Mismatching blockHash //
		otherBlock := mocks.ABlock(interfaces.GenesisBlock)
		badBlockHashProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(instanceId, nodeW2KeyManager, nodeIdW2, blockHeight, view, block),
				builders.APrepareMessage(instanceId, nodeW3KeyManager, nodeIdW3, blockHeight, view, otherBlock),
				builders.APrepareMessage(instanceId, nodeW4KeyManager, nodeIdW4, blockHeight, view, block),
			})

		actualBadBlockHash := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, badBlockHashProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, actualBadBlockHash, "Did not reject mismatching block hash")
	})

	t.Run("TestProofsValidatorWithDuplicate prepare sender Id", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: nodeW2KeyManager, MemberId: nodeIdW2},
			{KeyManager: nodeW3KeyManager, MemberId: nodeIdW3},
			{KeyManager: nodeW2KeyManager, MemberId: nodeIdW2},
		}
		preparedProofWithDuplicatePSenderId := builders.CreatePreparedProof(instanceId, leaderMessageSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithDuplicatePSenderId, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with duplicate sender Id")
	})
}
