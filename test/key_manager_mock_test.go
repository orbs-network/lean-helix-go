package test

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKeyManagerVerify(t *testing.T) {
	signerId := primitives.MemberId("SignerId")
	verifierId := primitives.MemberId("VerifierId")

	signerKeyManager := builders.NewMockKeyManager(signerId)
	verifierKeyManager := builders.NewMockKeyManager(verifierId)

	content := []byte{1, 2, 3}

	signature := signerKeyManager.SignConsensusMessage(content)

	actual := verifierKeyManager.VerifyConsensusMessage(content, signature, signerId)
	require.True(t, actual)

}
