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

type CalcLeaderPk = func(view ViewCounter) PublicKey

func ValidatePreparedProof(
	targetTerm BlockHeight,
	targetView ViewCounter,
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

	term := ppm.Term()
	if term != targetTerm {
		return false
	}

	view := ppm.View()
	if view >= targetView {
		return false
	}

	if len(prepareBlockRefMessages) < 2*f {
		return false
	}

	// TODO Refactor names here!!!

	if verifyBlockRefMessage(ppm, ppm.Sender(), keyManager) == false {
		return false
	}

	leaderPk := ppm.Sender().SenderPublicKey()
	if !calcLeaderPk(view).Equals(leaderPk) {
		return false
	}

	seen := make(map[PublicKeyStr]bool, len(prepareBlockRefMessages))
	for _, msg := range prepareBlockRefMessages {

		if keyManager.VerifyBlockRef(msg, msg.Sender()) == false {
			return false
		}

		publicKey := msg.Sender().SenderPublicKey()

		if publicKey.Equals(leaderPk) {
			return false
		}

		if isInMembers(membersPKs, &publicKey) == false {
			return false
		}

		if msg.Term() != term {
			return false
		}

		if msg.View() != view {
			return false
		}

		if !msg.BlockHash().Equals(ppm.BlockHash()) {
			return false
		}

		if _, ok := seen[PublicKeyStr(publicKey)]; ok {
			return false
		}

		seen[PublicKeyStr(publicKey)] = true
	}

	return true
}
