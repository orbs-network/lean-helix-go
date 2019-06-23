// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package tests

import (
	"context"
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
		Signature: signerKeyManager.SignConsensusMessage(context.Background(), 1, content),
	}

	badSenderSignature := &protocol.SenderSignatureBuilder{
		MemberId:  signerId,
		Signature: signerKeyManager.SignConsensusMessage(context.Background(), 1, []byte{6, 6, 6}),
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
