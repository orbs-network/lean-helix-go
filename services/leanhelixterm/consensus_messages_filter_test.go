// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func GeneratePreprepareMessage(instanceId primitives.InstanceId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.APreprepareMessage(instanceId, keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GeneratePrepareMessage(instanceId primitives.InstanceId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.APrepareMessage(instanceId, keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GenerateCommitMessage(instanceId primitives.InstanceId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string, randomSeed uint64) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.ACommitMessage(instanceId, keyManager, senderMemberId, blockHeight, view, block, randomSeed).ToConsensusRawMessage()
}

func GenerateViewChangeMessage(instanceId primitives.InstanceId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	return builders.AViewChangeMessage(instanceId, keyManager, senderMemberId, blockHeight, view, nil).ToConsensusRawMessage()
}

func GenerateNewViewMessage(instanceId primitives.InstanceId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.ANewViewMessage(instanceId, keyManager, senderMemberId, blockHeight, view, nil, nil, block).ToConsensusRawMessage()

}

func TestProcessingAMessage(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		instanceId := primitives.InstanceId(rand.Uint64())
		messagesHandler := mocks.NewTermMessagesHandlerMock()
		keyManager := mocks.NewMockKeyManager(primitives.MemberId("My ID"))
		consensusMessagesFilter := NewConsensusMessagesFilter(messagesHandler, keyManager, 99)

		ppm := GeneratePreprepareMessage(instanceId, 10, 20, "Sender MemberId")
		pm := GeneratePrepareMessage(instanceId, 10, 20, "Sender MemberId")
		cm := GenerateCommitMessage(instanceId, 10, 20, "Sender MemberId", 99)
		vcm := GenerateViewChangeMessage(instanceId, 10, 20, "Sender MemberId")
		nvm := GenerateNewViewMessage(instanceId, 10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.HistoryPP))
		require.Equal(t, 0, len(messagesHandler.HistoryP))
		require.Equal(t, 0, len(messagesHandler.HistoryC))
		require.Equal(t, 0, len(messagesHandler.HistoryNV))
		require.Equal(t, 0, len(messagesHandler.HistoryVC))

		err1 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(ppm))
		err2 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(pm))
		err3 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(cm))
		err4 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(vcm))
		err5 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(nvm))

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)
		require.NoError(t, err4)
		require.NoError(t, err5)

		require.Equal(t, 1, len(messagesHandler.HistoryPP))
		require.Equal(t, 1, len(messagesHandler.HistoryP))
		require.Equal(t, 1, len(messagesHandler.HistoryC))
		require.Equal(t, 1, len(messagesHandler.HistoryNV))
		require.Equal(t, 1, len(messagesHandler.HistoryVC))
	})
}

func TestFilteringACommitWithBadSeed(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		instanceId := primitives.InstanceId(rand.Uint64())
		messagesHandler := mocks.NewTermMessagesHandlerMock()
		keyManager := mocks.NewMockKeyManager(primitives.MemberId("My ID"))
		consensusMessagesFilter := NewConsensusMessagesFilter(messagesHandler, keyManager, 99)

		goodCommit := GenerateCommitMessage(instanceId, 10, 20, "Sender MemberId", 99)
		badCommit := GenerateCommitMessage(instanceId, 10, 20, "Sender MemberId", 666)

		require.Equal(t, 0, len(messagesHandler.HistoryC))

		err := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(goodCommit))
		require.NoError(t, err)
		require.Equal(t, 1, len(messagesHandler.HistoryC))

		err = consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(badCommit))
		require.NoError(t, err)
		require.Equal(t, 1, len(messagesHandler.HistoryC)) // still on 1
	})
}

func TestNotSendingMessagesWhenTheHandlerWasNotSet(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		instanceId := primitives.InstanceId(rand.Uint64())
		keyManager := mocks.NewMockKeyManager(primitives.MemberId("My ID"))
		consensusMessagesFilter := NewConsensusMessagesFilter(nil, keyManager, 99)

		ppm := GeneratePreprepareMessage(instanceId, 10, 20, "Sender MemberId")
		pm := GeneratePrepareMessage(instanceId, 10, 20, "Sender MemberId")
		cm := GenerateCommitMessage(instanceId, 10, 20, "Sender MemberId", 99)
		vcm := GenerateViewChangeMessage(instanceId, 10, 20, "Sender MemberId")
		nvm := GenerateNewViewMessage(instanceId, 10, 20, "Sender MemberId")

		err1 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(ppm))
		err2 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(pm))
		err3 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(cm))
		err4 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(vcm))
		err5 := consensusMessagesFilter.HandleConsensusMessage(interfaces.ToConsensusMessage(nvm))

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)
		require.NoError(t, err4)
		require.NoError(t, err5)
		
		// expect that we don't panic
	})
}
