package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

func isInMembers(membersIds []primitives.MemberId, memberId primitives.MemberId) bool {
	for _, currentId := range membersIds {
		if currentId.Equal(memberId) {
			return true
		}
	}
	return false
}

func verifyBlockRefMessage(blockRef *protocol.BlockRef, sender *protocol.SenderSignature, keyManager KeyManager) bool {
	return keyManager.Verify(blockRef.Raw(), sender)
}

type CalcLeaderId = func(view primitives.View) primitives.MemberId

func ValidatePreparedProof(
	targetHeight primitives.BlockHeight,
	targetView primitives.View,
	preparedProof *protocol.PreparedProof,
	q int,
	keyManager KeyManager,
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

	if len(pSenders) < q-1 {
		return false
	}

	if !verifyBlockRefMessage(ppBlockRef, ppSender, keyManager) {
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

	set := make(map[MemberIdStr]bool)
	for _, pSender := range pSenders {
		pSenderMemberId := pSender.MemberId()
		if keyManager.Verify(pBlockRef.Raw(), pSender) == false {
			return false
		}

		if pSenderMemberId.Equal(leaderFromPPMessage) {
			return false
		}

		if isInMembers(membersIds, pSenderMemberId) == false {
			return false
		}

		if _, ok := set[MemberIdStr(pSenderMemberId)]; ok {
			return false
		}

		set[MemberIdStr(pSenderMemberId)] = true
	}

	return true
}
