package leanhelix

func isInMembers(membersPKs []PublicKey, publicKey PublicKey) bool {
	for _, memberPK := range membersPKs {
		if memberPK.Equals(publicKey) {
			return true
		}
	}
	return false
}

func verifyBlockRefMessage(blockRef BlockRef, sender SenderSignature, keyManager KeyManager) bool {
	return keyManager.VerifyBlockRef(blockRef, sender)
}

type CalcLeaderPk = func(view View) PublicKey

func ValidatePreparedProof(
	targetHeight BlockHeight,
	targetView View,
	preparedProof PreparedProof,
	f int,
	keyManager KeyManager,
	membersPKs []PublicKey,
	calcLeaderPk CalcLeaderPk) bool {
	if preparedProof == nil {
		return true
	}

	ppBlockRef := preparedProof.PPBlockRef()
	ppSender := preparedProof.PPSender()
	pBlockRef := preparedProof.PBlockRef()
	pSenders := preparedProof.PSenders()

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

	if !pBlockRef.BlockHash().Equals(ppBlockRef.BlockHash()) {
		return false
	}

	if len(pSenders) < 2*f {
		return false
	}

	// TODO Refactor names here!!!

	if !verifyBlockRefMessage(ppBlockRef, ppSender, keyManager) {
		return false
	}

	leaderFromPPMessage := ppSender.SenderPublicKey()
	leaderFromView := calcLeaderPk(ppView)
	if !leaderFromView.Equals(leaderFromPPMessage) {
		return false
	}

	if !pBlockRef.BlockHeight().Equals(ppBlockHeight) {
		return false
	}

	if !pBlockRef.View().Equals(ppView) {
		return false
	}

	set := make(map[PublicKeyStr]bool, len(pSenders))
	for _, pSender := range pSenders {

		pSenderPublicKey := pSender.SenderPublicKey()
		if keyManager.VerifyBlockRef(pBlockRef, pSender) == false {
			return false
		}

		if pSenderPublicKey.Equals(leaderFromPPMessage) {
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
