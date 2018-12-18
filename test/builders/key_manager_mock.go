package builders

import (
	"bytes"
	"fmt"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
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

func (km *MockKeyManager) SignConsensusMessage(content []byte) []byte {
	str := fmt.Sprintf("SIG|%s|%x", km.myMemberId.KeyForMap(), content)
	return []byte(str)
}

func (km *MockKeyManager) VerifyConsensusMessage(content []byte, signature primitives.Signature, memberId primitives.MemberId) bool {
	if km.FailFutureVerifications {
		return false
	}

	for _, rejectedKey := range km.rejectedMemberIds {
		if rejectedKey.Equal(memberId) {
			return false
		}
	}

	str := fmt.Sprintf("SIG|%s|%x", memberId.KeyForMap(), content)
	expected := []byte(str)
	return bytes.Equal(expected, signature)
}
