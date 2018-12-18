package builders

import (
	"bytes"
	"fmt"
	"github.com/orbs-network/lean-helix-go"
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

func (km *MockKeyManager) SignConsensusMessage(blockHeight primitives.BlockHeight, content []byte) []byte {
	str := fmt.Sprintf("SIG|%s|%s|%x", blockHeight, km.myMemberId.KeyForMap(), content)
	return []byte(str)
}

func (km *MockKeyManager) VerifyConsensusMessage(blockHeight primitives.BlockHeight, content []byte, signature primitives.Signature, memberId primitives.MemberId) bool {
	if km.FailFutureVerifications {
		return false
	}

	for _, rejectedKey := range km.rejectedMemberIds {
		if rejectedKey.Equal(memberId) {
			return false
		}
	}

	str := fmt.Sprintf("SIG|%s|%s|%x", blockHeight, memberId.KeyForMap(), content)
	expected := []byte(str)
	return bytes.Equal(expected, signature)
}

func (km *MockKeyManager) SignRandomSeed(blockHeight primitives.BlockHeight, content []byte) []byte {
	return nil
}

func (km *MockKeyManager) VerifyRandomSeed(blockHeight primitives.BlockHeight, content []byte, signature primitives.Signature, memberId primitives.MemberId) bool {
	return false
}

func (km *MockKeyManager) AggregateRandomSeed(blockHeight primitives.BlockHeight, randomSeedShares []*leanhelix.RandomSeedShare) primitives.Signature {
	return nil
}
