package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type MockSenderSignature struct {
	senderPublicKey primitives.Ed25519PublicKey
	signature       primitives.Ed25519Sig
}

func NewMockSenderSignature(senderPublicKey primitives.Ed25519PublicKey, signature primitives.Ed25519Sig) lh.SenderSignature {
	return &MockSenderSignature{
		senderPublicKey: senderPublicKey,
		signature:       signature,
	}
}

func (s *MockSenderSignature) SenderPublicKey() primitives.Ed25519PublicKey {
	return s.senderPublicKey
}

func (s *MockSenderSignature) Signature() primitives.Ed25519Sig {
	return s.signature
}
