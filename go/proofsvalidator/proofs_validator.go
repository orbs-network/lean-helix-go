package proofsvalidator

import lh "github.com/orbs-network/lean-helix-go/go/leanhelix"

func ValidatePreparedProof(
	targetTerm lh.BlockHeight,
	targetView lh.ViewCounter,
	preparedProof *lh.PreparedProof,
	f int,
	keyManager lh.KeyManager) bool {
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

	term := preprepareBlockRefMessage.Term
	if term != targetTerm {
		return false
	}

	if len(prepareBlockRefMessages) < 2*f {
		return false
	}

	expectedPrePrepareMessageContent := preprepareBlockRefMessage.BlockMessageContent
	signaturePair := preprepareBlockRefMessage.SignaturePair
	leaderPk := signaturePair.SignerPublicKey
	contentSignature := signaturePair.ContentSignature
	if keyManager.VerifyBlockMessageContent(expectedPrePrepareMessageContent, contentSignature, leaderPk) == false {
		return false
	}

	for _, msg := range prepareBlockRefMessages {
		content := msg.BlockMessageContent
		signature := msg.SignaturePair.ContentSignature
		publicKey := msg.SignaturePair.SignerPublicKey
		if keyManager.VerifyBlockMessageContent(content, signature, publicKey) == false {
			return false
		}
	}

	return true
}
