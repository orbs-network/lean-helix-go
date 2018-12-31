package mocks

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

func (km *MockKeyManager) SignConsensusMessage(blockHeight primitives.BlockHeight, content []byte) primitives.Signature {
	str := fmt.Sprintf("SIG|%s|%s|%x", blockHeight, km.myMemberId.KeyForMap(), content)
	return []byte(str)
}

func (km *MockKeyManager) VerifyConsensusMessage(blockHeight primitives.BlockHeight, content []byte, sender *protocol.SenderSignature) bool {
	if km.FailFutureVerifications {
		return false
	}

	for _, rejectedKey := range km.rejectedMemberIds {
		if rejectedKey.Equal(sender.MemberId()) {
			return false
		}
	}

	str := fmt.Sprintf("SIG|%s|%s|%x", blockHeight, sender.MemberId().KeyForMap(), content)
	expected := []byte(str)
	return bytes.Equal(expected, sender.Signature())
}

func (km *MockKeyManager) SignRandomSeed(blockHeight primitives.BlockHeight, content []byte) primitives.RandomSeedSignature {
	str := fmt.Sprintf("RND_SIG|%s|%s|%x", blockHeight, km.myMemberId.KeyForMap(), content)
	return []byte(str)
}

func (km *MockKeyManager) VerifyRandomSeed(blockHeight primitives.BlockHeight, content []byte, sender *protocol.SenderSignature) bool {
	str := fmt.Sprintf("RND_SIG|%s|%s|%x", blockHeight, sender.MemberId().KeyForMap(), content)
	expected := []byte(str)

	aggStr := fmt.Sprintf("AGG_RND_SIG|%s", blockHeight)
	aggExpected := []byte(aggStr)
	return bytes.Equal(expected, sender.Signature()) || bytes.Equal(aggExpected, sender.Signature())
}

func (km *MockKeyManager) AggregateRandomSeed(blockHeight primitives.BlockHeight, randomSeedShares []*protocol.SenderSignature) primitives.RandomSeedSignature {
	str := fmt.Sprintf("AGG_RND_SIG|%s", blockHeight)
	return []byte(str)
}
