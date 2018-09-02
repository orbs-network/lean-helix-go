package proofsvalidator

import lh "github.com/orbs-network/lean-helix-go/go/leanhelix"

func ValidatePreparedProof(
	targetTerm lh.BlockHeight,
	targetView lh.ViewCounter,
	keyManager lh.KeyManager,
	preparedProof *lh.PreparedProof) bool {
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
