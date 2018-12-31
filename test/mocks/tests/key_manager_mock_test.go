package tests

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKeyManagerVerify(t *testing.T) {
	signerId := primitives.MemberId("SignerId")
	verifierId := primitives.MemberId("VerifierId")

	signerKeyManager := mocks.NewMockKeyManager(signerId)
	verifierKeyManager := mocks.NewMockKeyManager(verifierId)

	content := []byte{1, 2, 3}

	signature := signerKeyManager.SignConsensusMessage(1, content)

	senderSignature := &protocol.SenderSignatureBuilder{
		MemberId:  signerId,
		Signature: signature,
	}

	actual := verifierKeyManager.VerifyConsensusMessage(1, content, senderSignature.Build())
	require.True(t, actual)

}
