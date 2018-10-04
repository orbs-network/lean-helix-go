package leanhelix

func isInMembers(membersPKs *[]PublicKey, publicKey *PublicKey) bool {
	for _, memberPK := range *membersPKs {
		if memberPK.Equals(*publicKey) {
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
	membersPKs *[]PublicKey,
	calcLeaderPk CalcLeaderPk) bool {
	if preparedProof == nil {
		return true
	}

	ppm := preparedProof.PreprepareMessage()
	if ppm == nil {
		return false
	}

	prepareBlockRefMessages := preparedProof.PrepareMessages()
	if prepareBlockRefMessages == nil {
		return false
	}

	term := ppm.SignedHeader().BlockHeight()
	if term != targetHeight {
		return false
	}

	view := ppm.SignedHeader().View()
	if view >= targetView {
		return false
	}

	if len(prepareBlockRefMessages) < 2*f {
		return false
	}

	// TODO Refactor names here!!!

	if verifyBlockRefMessage(ppm.SignedHeader(), ppm.Sender(), keyManager) == false {
		return false
	}

	leaderPk := ppm.Sender().SenderPublicKey()
	if !calcLeaderPk(view).Equals(leaderPk) {
		return false
	}

	seen := make(map[PublicKeyStr]bool, len(prepareBlockRefMessages))
	for _, msg := range prepareBlockRefMessages {

		if keyManager.VerifyBlockRef(msg.SignedHeader(), msg.Sender()) == false {
			return false
		}

		publicKey := msg.Sender().SenderPublicKey()

		if publicKey.Equals(leaderPk) {
			return false
		}

		if isInMembers(membersPKs, &publicKey) == false {
			return false
		}

		if msg.SignedHeader().BlockHeight() != term {
			return false
		}

		if msg.SignedHeader().View() != view {
			return false
		}

		if !msg.SignedHeader().BlockHash().Equals(ppm.SignedHeader().BlockHash()) {
			return false
		}

		if _, ok := seen[PublicKeyStr(publicKey)]; ok {
			return false
		}

		seen[PublicKeyStr(publicKey)] = true
	}

	return true
}
