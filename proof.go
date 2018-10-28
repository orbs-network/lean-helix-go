package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

func isInMembers(membersPKs []Ed25519PublicKey, publicKey Ed25519PublicKey) bool {
	for _, memberPK := range membersPKs {
		if memberPK.Equal(publicKey) {
			return true
		}
	}
	return false
}

func verifyBlockRefMessage(blockRef *BlockRef, sender *SenderSignature, keyManager KeyManager) bool {
	return keyManager.Verify(blockRef.Raw(), sender)
}

type CalcLeaderPk = func(view View) Ed25519PublicKey

func CreatePreparedProofBuilder(ppKeyManager KeyManager, pKeyManagers []KeyManager, height BlockHeight, view View, blockHash Uint256) *PreparedProofBuilder {

	var ppBlockRef *BlockRefBuilder
	var pBlockRef *BlockRefBuilder
	var ppSender *SenderSignatureBuilder
	var pSenders []*SenderSignatureBuilder

	if len(pKeyManagers) == 0 {
		pBlockRef = nil
		pSenders = nil
	} else {
		pBlockRef = &BlockRefBuilder{
			MessageType: LEAN_HELIX_PREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders = make([]*SenderSignatureBuilder, len(pKeyManagers))
		for i, mgr := range pKeyManagers {
			pSenders[i] = &SenderSignatureBuilder{
				SenderPublicKey: mgr.MyPublicKey(),
				Signature:       mgr.Sign(pBlockRef.Build().Raw()),
			}
		}
	}
	if ppKeyManager == nil {
		ppBlockRef = nil
		ppSender = nil
	} else {
		ppBlockRef = &BlockRefBuilder{
			MessageType: LEAN_HELIX_PREPREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		ppSender = &SenderSignatureBuilder{
			SenderPublicKey: ppKeyManager.MyPublicKey(),
			Signature:       ppKeyManager.Sign(ppBlockRef.Build().Raw()),
		}
	}
	preparedProof := &PreparedProofBuilder{
		PreprepareBlockRef: ppBlockRef,
		PreprepareSender:   ppSender,
		PrepareBlockRef:    pBlockRef,
		PrepareSenders:     pSenders,
	}
	return preparedProof
}

func CreatePreparedProof(ppKeyManager KeyManager, pKeyManagers []KeyManager, height BlockHeight, view View, blockHash Uint256) *PreparedProof {
	return CreatePreparedProofBuilder(ppKeyManager, pKeyManagers, height, view, blockHash).Build()
}

func ValidatePreparedProof(
	targetHeight BlockHeight,
	targetView View,
	preparedProof *PreparedProof,
	q int,
	keyManager KeyManager,
	membersPKs []Ed25519PublicKey,
	calcLeaderPk CalcLeaderPk) bool {
	if preparedProof == nil {
		return true
	}

	ppBlockRef := preparedProof.PreprepareBlockRef()
	ppSender := preparedProof.PreprepareSender()
	pBlockRef := preparedProof.PrepareBlockRef()
	pSendersIter := preparedProof.PrepareSendersIterator()
	pSenders := make([]*SenderSignature, 0, 1)

	for {
		if !pSendersIter.HasNext() {
			break
		}
		pSenders = append(pSenders, pSendersIter.NextPrepareSenders())
	}

	if ppBlockRef == nil {
		return false
	}

	if pSenders == nil {
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

	if !pBlockRef.BlockHash().Equal(ppBlockRef.BlockHash()) {
		return false
	}

	if len(pSenders) < q-1 {
		return false
	}

	// TODO Refactor names here!!!

	if !verifyBlockRefMessage(ppBlockRef, ppSender, keyManager) {
		return false
	}

	leaderFromPPMessage := ppSender.SenderPublicKey()
	leaderFromView := calcLeaderPk(ppView)
	if !leaderFromView.Equal(leaderFromPPMessage) {
		return false
	}

	if !pBlockRef.BlockHeight().Equal(ppBlockHeight) {
		return false
	}

	if !pBlockRef.View().Equal(ppView) {
		return false
	}

	set := make(map[PublicKeyStr]bool)
	for _, pSender := range pSenders {

		pSenderPublicKey := pSender.SenderPublicKey()
		if keyManager.Verify(pBlockRef.Raw(), pSender) == false {
			return false
		}

		if pSenderPublicKey.Equal(leaderFromPPMessage) {
			return false
		}

		if isInMembers(membersPKs, pSenderPublicKey) == false {
			return false
		}

		if _, ok := set[PublicKeyStr(pSenderPublicKey)]; ok {
			return false
		}

		set[PublicKeyStr(pSenderPublicKey)] = true
	}

	return true
}
