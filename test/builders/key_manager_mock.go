package builders

import (
	"bytes"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type mockKeyManager struct {
	myPublicKey        Ed25519PublicKey
	rejectedPublicKeys []Ed25519PublicKey
}

func NewMockKeyManager(publicKey Ed25519PublicKey, rejectedPublicKeys ...Ed25519PublicKey) *mockKeyManager {
	return &mockKeyManager{
		myPublicKey:        publicKey,
		rejectedPublicKeys: rejectedPublicKeys,
	}
}

func (km *mockKeyManager) Sign(content []byte) []byte {
	return []byte(fmt.Sprintf("SIG|%x|%x", km.myPublicKey, content))
}

func (km *mockKeyManager) Verify(content []byte, sender *lh.SenderSignature) bool {
	for _, rejectedKey := range km.rejectedPublicKeys {
		if rejectedKey.Equal(sender.SenderPublicKey()) {
			return false
		}
	}

	expected := []byte(fmt.Sprintf("SIG|%x|%x", sender.SenderPublicKey(), content))
	return bytes.Equal(expected, sender.Signature())
}

func (km *mockKeyManager) MyPublicKey() Ed25519PublicKey {
	return km.myPublicKey
}
