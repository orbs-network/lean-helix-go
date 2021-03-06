// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package proofsvalidator

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

func IsInMembers(members []interfaces.CommitteeMember, memberId primitives.MemberId) bool {
	for _, currentMember := range members {
		if currentMember.Id.Equal(memberId) {
			return true
		}
	}
	return false
}

func VerifyBlockRefMessage(blockRef *protocol.BlockRef, sender *protocol.SenderSignature, keyManager interfaces.KeyManager) error {
	return keyManager.VerifyConsensusMessage(blockRef.BlockHeight(), blockRef.Raw(), sender)
}

type CalcLeaderId = func(view primitives.View) primitives.MemberId

func ValidatePreparedProof(
	targetHeight primitives.BlockHeight,
	targetView primitives.View,
	preparedProof *protocol.PreparedProof,
	keyManager interfaces.KeyManager,
	committeeMembers []interfaces.CommitteeMember,
	calcLeaderId CalcLeaderId) bool {
	if preparedProof == nil || len(preparedProof.Raw()) == 0 {
		return true
	}

	ppBlockRef := preparedProof.PreprepareBlockRef()
	ppSender := preparedProof.PreprepareSender()
	pBlockRef := preparedProof.PrepareBlockRef()
	pSendersIter := preparedProof.PrepareSendersIterator()
	pSenders := make([]*protocol.SenderSignature, 0, 1)

	for {
		if !pSendersIter.HasNext() {
			break
		}
		pSenders = append(pSenders, pSendersIter.NextPrepareSenders())
	}

	if ppSender == nil || pSenders == nil || ppBlockRef == nil || pBlockRef == nil {
		return false
	}

	ppBlockHeight := ppBlockRef.BlockHeight()

	if ppBlockHeight != targetHeight {
		return false
	}

	ppView := ppBlockRef.View()
	if ppView >= targetView {
		return false
	}

	senderIds := make([]primitives.MemberId, len(pSenders))
	for i, pSender := range pSenders {
		senderIds[i] = pSender.MemberId()
	}
	senderIds = append(senderIds, ppSender.MemberId())

	isQuorum, _, _ := quorum.IsQuorum(senderIds, committeeMembers)
	if !isQuorum {
		return false
	}

	if err := VerifyBlockRefMessage(ppBlockRef, ppSender, keyManager); err != nil {
		return false
	}

	leaderFromPPMessage := ppSender.MemberId()
	leaderFromView := calcLeaderId(ppView)
	if !leaderFromView.Equal(leaderFromPPMessage) {
		return false
	}

	if !pBlockRef.BlockHash().Equal(ppBlockRef.BlockHash()) {
		return false
	}

	if !pBlockRef.BlockHeight().Equal(ppBlockHeight) {
		return false
	}

	if !pBlockRef.View().Equal(ppView) {
		return false
	}

	set := make(map[storage.MemberIdStr]bool)
	for _, pSender := range pSenders {
		pSenderMemberId := pSender.MemberId()
		if err := keyManager.VerifyConsensusMessage(pBlockRef.BlockHeight(), pBlockRef.Raw(), pSender); err != nil {
			return false
		}

		if pSenderMemberId.Equal(leaderFromPPMessage) {
			return false
		}

		if IsInMembers(committeeMembers, pSenderMemberId) == false {
			return false
		}

		if _, ok := set[storage.MemberIdStr(pSenderMemberId)]; ok {
			return false
		}

		set[storage.MemberIdStr(pSenderMemberId)] = true
	}

	return true
}
