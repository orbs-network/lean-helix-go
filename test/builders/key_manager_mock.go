package builders

import (
	"bytes"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type MockKeyManager struct {
	myPublicKey             Ed25519PublicKey
	rejectedPublicKeys      []Ed25519PublicKey
	FailFutureVerifications bool
}

func NewMockKeyManager(publicKey Ed25519PublicKey, rejectedPublicKeys ...Ed25519PublicKey) *MockKeyManager {
	return &MockKeyManager{
		myPublicKey:             publicKey,
		rejectedPublicKeys:      rejectedPublicKeys,
		FailFutureVerifications: false,
	}
}

func (km *MockKeyManager) Sign(content []byte) []byte {
	str := fmt.Sprintf("SIG|%s|%x", km.myPublicKey.KeyForMap(), content)
	return []byte(str)
}

func (km *MockKeyManager) Verify(content []byte, sender *lh.SenderSignature) bool {
	if km.FailFutureVerifications {
		return false
	}

	for _, rejectedKey := range km.rejectedPublicKeys {
		if rejectedKey.Equal(sender.SenderPublicKey()) {
			return false
		}
	}

	str := fmt.Sprintf("SIG|%s|%x", sender.SenderPublicKey().KeyForMap(), content)
	expected := []byte(str)
	return bytes.Equal(expected, sender.Signature())
}

func (km *MockKeyManager) MyPublicKey() Ed25519PublicKey {
	return km.myPublicKey
}
