package builders

import (
	"bytes"
	"fmt"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type MockKeyManager struct {
	myPublicKey             primitives.MemberId
	rejectedPublicKeys      []primitives.MemberId
	FailFutureVerifications bool
}

func NewMockKeyManager(publicKey primitives.MemberId, rejectedPublicKeys ...primitives.MemberId) *MockKeyManager {
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

func (km *MockKeyManager) Verify(content []byte, sender *protocol.SenderSignature) bool {
	if km.FailFutureVerifications {
		return false
	}

	for _, rejectedKey := range km.rejectedPublicKeys {
		if rejectedKey.Equal(sender.MemberId()) {
			return false
		}
	}

	str := fmt.Sprintf("SIG|%s|%x", sender.MemberId().KeyForMap(), content)
	expected := []byte(str)
	return bytes.Equal(expected, sender.Signature())
}

func (km *MockKeyManager) MyPublicKey() primitives.MemberId {
	return km.myPublicKey
}
