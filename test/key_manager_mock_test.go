package test

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKeyManagerVerify(t *testing.T) {
	signerPk := primitives.MemberId("SignerPK")
	verifierPk := primitives.MemberId("VerifierPK")

	signerKeyManager := builders.NewMockKeyManager(signerPk)
	verifierKeyManager := builders.NewMockKeyManager(verifierPk)

	content := []byte{1, 2, 3}

	signature := signerKeyManager.Sign(content)

	senderSignature := &protocol.SenderSignatureBuilder{
		MemberId:  signerPk,
		Signature: signature,
	}

	actual := verifierKeyManager.Verify(content, senderSignature.Build())
	require.True(t, actual)

}
