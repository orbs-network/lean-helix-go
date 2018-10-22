package builders

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

// TODO Keys should not be strings - convert to our primitives

const PRIVATE_KEY_PREFIX = "PRIVATE_KEY"

type mockKeyManager struct {
	mock.Mock
	myPublicKey        Ed25519PublicKey
	RejectedPublicKeys []Ed25519PublicKey
}

func NewMockKeyManager(publicKey Ed25519PublicKey, rejectedPublicKeys ...Ed25519PublicKey) *mockKeyManager {
	return &mockKeyManager{
		myPublicKey:        publicKey,
		RejectedPublicKeys: rejectedPublicKeys,
	}
}

var MOCK_SIG_PREFIX = []byte("SIG|")

func (km *mockKeyManager) Sign(content []byte) []byte {
	return append(MOCK_SIG_PREFIX, content...)
}

func (km *mockKeyManager) Verify(content []byte, sender *lh.SenderSignature) bool {
	ret := km.Called(content, sender)
	return ret.Bool(0)
}

func (km *mockKeyManager) MyPublicKey() Ed25519PublicKey {
	return km.myPublicKey
}
