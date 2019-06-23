// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"bytes"
	"fmt"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/pkg/errors"
)

type VerifyRandomSeedCallParams struct {
	BlockHeight primitives.BlockHeight
	Content     []byte
	Sender      *protocol.SenderSignature
}

type MockKeyManager struct {
	myMemberId              primitives.MemberId
	rejectedMemberIds       []primitives.MemberId
	FailFutureVerifications bool
	VerifyRandomSeedHistory []*VerifyRandomSeedCallParams
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

func (km *MockKeyManager) VerifyConsensusMessage(blockHeight primitives.BlockHeight, content []byte, sender *protocol.SenderSignature) error {
	if km.FailFutureVerifications {
		return errors.New("FailFutureVerifications=true")
	}

	for _, rejectedKey := range km.rejectedMemberIds {
		if rejectedKey.Equal(sender.MemberId()) {
			return errors.New("memberId equals rejectedKey")
		}
	}

	str := fmt.Sprintf("SIG|%s|%s|%x", blockHeight, sender.MemberId().KeyForMap(), content)
	expected := []byte(str)
	if !bytes.Equal(expected, sender.Signature()) {
		return errors.New("expected is different from sender.Signature")
	}
	return nil
}

func (km *MockKeyManager) SignRandomSeed(blockHeight primitives.BlockHeight, content []byte) primitives.RandomSeedSignature {
	str := fmt.Sprintf("RND_SIG|%s|%s|%x", blockHeight, km.myMemberId.KeyForMap(), content)
	return []byte(str)
}

func (km *MockKeyManager) VerifyRandomSeed(blockHeight primitives.BlockHeight, content []byte, sender *protocol.SenderSignature) error {
	km.VerifyRandomSeedHistory = append(km.VerifyRandomSeedHistory, &VerifyRandomSeedCallParams{blockHeight, content, sender})

	str := fmt.Sprintf("RND_SIG|%s|%s|%x", blockHeight, sender.MemberId().KeyForMap(), content)
	expected := []byte(str)

	aggStr := fmt.Sprintf("AGG_RND_SIG|%s", blockHeight)
	aggExpected := []byte(aggStr)
	if !bytes.Equal(expected, sender.Signature()) && !bytes.Equal(aggExpected, sender.Signature()) {
		return errors.Errorf("Mismatch in expected and actual signatures")
	}
	return nil
}

func (km *MockKeyManager) AggregateRandomSeed(blockHeight primitives.BlockHeight, randomSeedShares []*protocol.SenderSignature) primitives.RandomSeedSignature {
	str := fmt.Sprintf("AGG_RND_SIG|%s", blockHeight)
	return []byte(str)
}
