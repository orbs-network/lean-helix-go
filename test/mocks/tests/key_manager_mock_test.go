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

	goodSenderSignature := &protocol.SenderSignatureBuilder{
		MemberId:  signerId,
		Signature: signerKeyManager.SignConsensusMessage(1, content),
	}

	badSenderSignature := &protocol.SenderSignatureBuilder{
		MemberId:  signerId,
		Signature: signerKeyManager.SignConsensusMessage(1, []byte{6, 6, 6}),
	}

	require.NoError(t, verifierKeyManager.VerifyConsensusMessage(1, content, goodSenderSignature.Build()))
	require.Error(t, verifierKeyManager.VerifyConsensusMessage(1, content, badSenderSignature.Build()))
}

func TestKeyManagerRandomSeedVerify(t *testing.T) {
	signerId := primitives.MemberId("SignerId")
	verifierId := primitives.MemberId("VerifierId")

	signerKeyManager := mocks.NewMockKeyManager(signerId)
	verifierKeyManager := mocks.NewMockKeyManager(verifierId)

	content := []byte{1, 2, 3}

	goodSenderSignature := &protocol.SenderSignatureBuilder{
		MemberId:  signerId,
		Signature: primitives.Signature(signerKeyManager.SignRandomSeed(1, content)),
	}

	badSenderSignature := &protocol.SenderSignatureBuilder{
		MemberId:  signerId,
		Signature: primitives.Signature(signerKeyManager.SignRandomSeed(1, []byte{6, 6, 6})),
	}

	require.Nil(t, verifierKeyManager.VerifyRandomSeed(1, content, goodSenderSignature.Build()))
	require.Error(t, verifierKeyManager.VerifyRandomSeed(1, content, badSenderSignature.Build()))
}

func TestKeyManagerAggregateVerification(t *testing.T) {
	signerId := primitives.MemberId("SignerId")
	verifierId := primitives.MemberId("VerifierId")

	signerKeyManager := mocks.NewMockKeyManager(signerId)
	verifierKeyManager := mocks.NewMockKeyManager(verifierId)

	content := []byte{1, 2, 3}

	goodSenderSignature := &protocol.SenderSignatureBuilder{
		MemberId:  signerId,
		Signature: primitives.Signature(signerKeyManager.AggregateRandomSeed(1, nil)),
	}

	badSenderSignature := &protocol.SenderSignatureBuilder{
		MemberId:  signerId,
		Signature: primitives.Signature(signerKeyManager.AggregateRandomSeed(2, nil)),
	}

	require.Nil(t, verifierKeyManager.VerifyRandomSeed(1, content, goodSenderSignature.Build()))
	require.Error(t, verifierKeyManager.VerifyRandomSeed(1, content, badSenderSignature.Build()))
}
