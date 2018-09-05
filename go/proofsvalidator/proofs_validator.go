package proofsvalidator

import lh "github.com/orbs-network/lean-helix-go/go/leanhelix"

func isInMembers(membersPKs *[]lh.PublicKey, publicKey *lh.PublicKey) bool {
	for _, memberPK := range *membersPKs {
		if memberPK == *publicKey {
			return true
		}
	}
	return false
}

func verifyBlockRefMessage(msg *lh.BlockRefMessage, keyManager lh.KeyManager) bool {
	content := msg.Content
	publicKey := msg.SignaturePair.SignerPublicKey
	signature := msg.SignaturePair.ContentSignature
	return keyManager.VerifyBlockMessageContent(content, signature, publicKey)
}

type CalcLeaderPk = func(view lh.ViewCounter) lh.PublicKey

func ValidatePreparedProof(
	targetTerm lh.BlockHeight,
	targetView lh.ViewCounter,
	preparedProof *lh.PreparedProof,
	f int,
	keyManager lh.KeyManager,
	membersPKs *[]lh.PublicKey,
	calcLeaderPk CalcLeaderPk) bool {
	if preparedProof == nil {
		return true
	}

	preprepareBlockRefMessage := preparedProof.PreprepareBlockRefMessage
	if preprepareBlockRefMessage == nil {
		return false
	}

	prepareBlockRefMessages := preparedProof.PrepareBlockRefMessages
	if prepareBlockRefMessages == nil {
		return false
	}

	term := preprepareBlockRefMessage.Content.Term
	if term != targetTerm {
		return false
	}

	view := preprepareBlockRefMessage.Content.View
	if view >= targetView {
		return false
	}

	if len(prepareBlockRefMessages) < 2*f {
		return false
	}

	// TODO Refactor names here!!!
	if verifyBlockRefMessage(preprepareBlockRefMessage.BlockRefMessage, keyManager) == false {
		return false
	}

	leaderPk := preprepareBlockRefMessage.SignaturePair.SignerPublicKey
	if calcLeaderPk(view) != leaderPk {
		return false
	}

	seen := make(map[lh.PublicKey]bool, len(prepareBlockRefMessages))
	for _, msg := range prepareBlockRefMessages {
		content := msg.Content
		signature := msg.SignaturePair.ContentSignature
		publicKey := msg.SignaturePair.SignerPublicKey

		if keyManager.VerifyBlockMessageContent(content, signature, publicKey) == false {
			return false
		}

		if publicKey == leaderPk {
			return false
		}

		if isInMembers(membersPKs, &publicKey) == false {
			return false
		}

		if content.Term != term {
			return false
		}

		if content.View != view {
			return false
		}

		if content.BlockHash != preprepareBlockRefMessage.Content.BlockHash {
			return false
		}

		if _, ok := seen[publicKey]; ok {
			return false
		}

		seen[publicKey] = true
	}

	return true
}
