// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package proofsvalidator

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

func IsInMembers(membersIds []primitives.MemberId, memberId primitives.MemberId) bool {
	for _, currentId := range membersIds {
		if currentId.Equal(memberId) {
			return true
		}
	}
	return false
}

func VerifyBlockRefMessage(blockRef *protocol.BlockRef, sender *protocol.SenderSignature, keyManager interfaces.KeyManager) error {
	return keyManager.VerifyConsensusMessage(blockRef.BlockHeight(), blockRef.Raw(), sender)
}

type CalcLeaderId = func(view primitives.View) primitives.MemberId

// TODO does this verify no repetition of sender addresses?
func ValidatePreparedProof(
	targetHeight primitives.BlockHeight,
	targetView primitives.View,
	preparedProof *protocol.PreparedProof,
	isQuorum func([]primitives.MemberId) bool,
	keyManager interfaces.KeyManager,
	membersIds []primitives.MemberId,
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

	pSendersMemberIds := make([]primitives.MemberId, len(pSenders)+1)
	pSendersMemberIds[0] = ppSender.MemberId()
	for i := 1; i <= len(pSenders); i++ {
		pSendersMemberIds[i] = pSenders[i-1].MemberId()
	}

	//if len(pSenders) < q-1 {
	if !isQuorum(pSendersMemberIds) { // todo -1?
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

		if IsInMembers(membersIds, pSenderMemberId) == false {
			return false
		}

		if _, ok := set[storage.MemberIdStr(pSenderMemberId)]; ok {
			return false
		}

		set[storage.MemberIdStr(pSenderMemberId)] = true
	}

	return true
}
