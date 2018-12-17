package builders

import (
	"bytes"
	"fmt"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type MockKeyManager struct {
	myMemberId              primitives.MemberId
	rejectedMemberIds       []primitives.MemberId
	FailFutureVerifications bool
}

func NewMockKeyManager(memberId primitives.MemberId, rejectedMemberIds ...primitives.MemberId) *MockKeyManager {
	return &MockKeyManager{
		myMemberId:              memberId,
		rejectedMemberIds:       rejectedMemberIds,
		FailFutureVerifications: false,
	}
}

func (km *MockKeyManager) Sign(content []byte) []byte {
	str := fmt.Sprintf("SIG|%s|%x", km.myMemberId.KeyForMap(), content)
	return []byte(str)
}

func (km *MockKeyManager) Verify(content []byte, sender *protocol.SenderSignature) bool {
	if km.FailFutureVerifications {
		return false
	}

	for _, rejectedKey := range km.rejectedMemberIds {
		if rejectedKey.Equal(sender.MemberId()) {
			return false
		}
	}

	str := fmt.Sprintf("SIG|%s|%x", sender.MemberId().KeyForMap(), content)
	expected := []byte(str)
	return bytes.Equal(expected, sender.Signature())
}
