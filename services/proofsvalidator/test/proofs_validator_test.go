// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func genMembers(ids []primitives.MemberId) []interfaces.CommitteeMember {
	members := make([]interfaces.CommitteeMember, len(ids))
	for i := 0; i < len(ids); i++ {
		members[i] = interfaces.CommitteeMember{
			Id:     ids[i],
			Weight: 1,
		}
	}
	return members
}

func TestProofsValidator(t *testing.T) {
	instanceId := primitives.InstanceId(rand.Uint64())
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
	committeeMembers := genMembers(membersIds)

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
			{KeyManager: node1KeyManager, MemberId: node1Id},
			{KeyManager: node2KeyManager, MemberId: node2Id},
			{KeyManager: node3KeyManager, MemberId: node3Id},
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

	t.Run("TestProofsValidatorWithNotEnoughPrepareMessages", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: node1KeyManager, MemberId: node1Id},
		}
		preparedProofWithNotEnoughP := builders.CreatePreparedProof(instanceId, leaderMessageSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithNotEnoughP, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with not enough prepares")
	})

	t.Run("TestProofsValidatorWithBadPreprepareSignature", func(t *testing.T) {
		rejectingKeyManager := mocks.NewMockKeyManager(myMemberId, leaderId)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, rejectingKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof that did not pass preprepare signature validation")
	})

	t.Run("TestProofsValidatorWithBadPrepareSignature", func(t *testing.T) {
		rejectingKeyManager := mocks.NewMockKeyManager(myMemberId, node2Id)
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
			{KeyManager: node1KeyManager, MemberId: node1Id},
			{KeyManager: node2KeyManager, MemberId: node2Id},
			{KeyManager: nonMemberKeyManager, MemberId: nonMemberId},
		}
		preparedProof := builders.CreatePreparedProof(instanceId, leaderMessageSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with a none member")
	})

	t.Run("TestProofsValidatorWithPrepareFromTheLeader", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: leaderKeyManager, MemberId: leaderId},
			{KeyManager: node1KeyManager, MemberId: node1Id},
			{KeyManager: node2KeyManager, MemberId: node2Id},
		}
		ppSigner := &builders.MessageSigner{KeyManager: node1KeyManager, MemberId: node1Id}

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
			builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(instanceId, node1KeyManager, node1Id, blockHeight, view, block),
				builders.APrepareMessage(instanceId, node2KeyManager, node2Id, blockHeight, view, block),
				builders.APrepareMessage(instanceId, node3KeyManager, node3Id, blockHeight, view, block),
			})

		actualGood := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, goodPrepareProof, myKeyManager, committeeMembers, calcLeaderId)
		require.True(t, actualGood, "Did not approve a valid proof")

		// Mismatching blockHeight //
		badHeightProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(instanceId, node1KeyManager, node1Id, blockHeight, view, block),
				builders.APrepareMessage(instanceId, node2KeyManager, node2Id, 666, view, block),
				builders.APrepareMessage(instanceId, node3KeyManager, node3Id, blockHeight, view, block),
			})

		actualBadHeight := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, badHeightProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, actualBadHeight, "Did not reject mismatching blockHeight")

		// Mismatching view //
		badViewProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(instanceId, node1KeyManager, node1Id, blockHeight, view, block),
				builders.APrepareMessage(instanceId, node2KeyManager, node2Id, blockHeight, 666, block),
				builders.APrepareMessage(instanceId, node3KeyManager, node3Id, blockHeight, view, block),
			})

		actualBadView := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, badViewProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, actualBadView, "Did not reject mismatching view")

		// Mismatching blockHash //
		otherBlock := mocks.ABlock(interfaces.GenesisBlock)
		badBlockHashProof := builders.APreparedProofByMessages(
			builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, view, block),
			[]*interfaces.PrepareMessage{
				builders.APrepareMessage(instanceId, node1KeyManager, node1Id, blockHeight, view, block),
				builders.APrepareMessage(instanceId, node2KeyManager, node2Id, blockHeight, view, otherBlock),
				builders.APrepareMessage(instanceId, node3KeyManager, node3Id, blockHeight, view, block),
			})

		actualBadBlockHash := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, badBlockHashProof, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, actualBadBlockHash, "Did not reject mismatching block hash")
	})

	t.Run("TestProofsValidatorWithDuplicate prepare sender Id", func(t *testing.T) {
		pSigners := []*builders.MessageSigner{
			{KeyManager: node1KeyManager, MemberId: node1Id},
			{KeyManager: node2KeyManager, MemberId: node2Id},
			{KeyManager: node1KeyManager, MemberId: node1Id},
		}
		preparedProofWithDuplicatePSenderId := builders.CreatePreparedProof(instanceId, leaderMessageSigner, pSigners, blockHeight, view, blockHash)
		result := proofsvalidator.ValidatePreparedProof(targetBlockHeight, targetView, preparedProofWithDuplicatePSenderId, myKeyManager, committeeMembers, calcLeaderId)
		require.False(t, result, "Did not reject a proof with duplicate sender Id")
	})
}
