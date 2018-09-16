package builders

import lh "github.com/orbs-network/lean-helix-go"

type MockSenderSignature struct {
	senderPublicKey lh.PublicKey
	signature       lh.Signature
}

func NewMockSenderSignature(senderPublicKey lh.PublicKey, signature lh.Signature) lh.SenderSignature {
	return &MockSenderSignature{
		senderPublicKey: senderPublicKey,
		signature:       signature,
	}
}

func (s *MockSenderSignature) SenderPublicKey() lh.PublicKey {
	return s.senderPublicKey
}

func (s *MockSenderSignature) Signature() lh.Signature {
	return s.signature
}
