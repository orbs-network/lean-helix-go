package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKeyManagerVerify(t *testing.T) {
	signerPk := primitives.Ed25519PublicKey("SignerPK")
	verifierPk := primitives.Ed25519PublicKey("VerifierPK")

	signerKeyManager := builders.NewMockKeyManager(signerPk)
	verifierKeyManager := builders.NewMockKeyManager(verifierPk)

	content := []byte{1, 2, 3}

	signature := signerKeyManager.Sign(content)

	senderSignature := &leanhelix.SenderSignatureBuilder{
		SenderPublicKey: signerPk,
		Signature:       signature,
	}

	actual := verifierKeyManager.Verify(content, senderSignature.Build())
	require.True(t, actual)

}
