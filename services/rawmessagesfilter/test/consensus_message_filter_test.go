// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/rawmessagesfilter"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func testLogger(state *state.State) L.LHLogger {
	return L.NewLhLogger(mocks.NewMockConfigSimple(), state)
}

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

func GenerateCommitMessage(instanceId primitives.InstanceId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.ACommitMessage(instanceId, keyManager, senderMemberId, blockHeight, view, block, 0).ToConsensusRawMessage()
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

func TestGettingAMessage(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		instanceId := primitives.InstanceId(rand.Uint64())
		state := mocks.NewMockState().WithHeightView(10, 20)
		filter := rawmessagesfilter.NewConsensusMessageFilter(instanceId, primitives.MemberId("My MemberId"), testLogger(state.State), state.State)
		messagesHandler := NewTermMessagesHandlerMock()
		filter.ConsumeCacheMessages(messagesHandler)

		ppm := GeneratePreprepareMessage(instanceId, 10, 20, "Sender MemberId")
		pm := GeneratePrepareMessage(instanceId, 10, 20, "Sender MemberId")
		cm := GenerateCommitMessage(instanceId, 10, 20, "Sender MemberId")
		vcm := GenerateViewChangeMessage(instanceId, 10, 20, "Sender MemberId")
		nvm := GenerateNewViewMessage(instanceId, 10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.history))

		filter.HandleConsensusRawMessage(ppm)
		filter.HandleConsensusRawMessage(pm)
		filter.HandleConsensusRawMessage(cm)
		filter.HandleConsensusRawMessage(vcm)
		filter.HandleConsensusRawMessage(nvm)

		require.Equal(t, 5, len(messagesHandler.history))
	})
}

func TestFilterMessagesFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		instanceId := primitives.InstanceId(rand.Uint64())
		state := mocks.NewMockState().WithHeightView(10, 0)
		filter := rawmessagesfilter.NewConsensusMessageFilter(instanceId, primitives.MemberId("My MemberId"), testLogger(state.State), state.State)
		messagesHandler := NewTermMessagesHandlerMock()
		filter.ConsumeCacheMessages(messagesHandler)

		messageFromThePast := GeneratePreprepareMessage(instanceId, 9, 20, "Sender MemberId")
		messageFromThePresent := GeneratePreprepareMessage(instanceId, 10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.history))

		filter.HandleConsensusRawMessage(messageFromThePast)
		filter.HandleConsensusRawMessage(messageFromThePresent)

		require.Equal(t, 1, len(messagesHandler.history))
	})
}

func TestFilterMessagesWithBadInstanceId(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		state := mocks.NewMockState().WithHeightView(10, 0)
		filter := rawmessagesfilter.NewConsensusMessageFilter(777, primitives.MemberId("My MemberId"), testLogger(state.State), state.State)
		messagesHandler := NewTermMessagesHandlerMock()
		filter.ConsumeCacheMessages(messagesHandler)

		messageWithGoodInstanceId := GeneratePreprepareMessage(777, 10, 20, "Sender MemberId")
		messageWithBadInstanceId := GeneratePreprepareMessage(666, 10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.history))

		filter.HandleConsensusRawMessage(messageWithGoodInstanceId)
		filter.HandleConsensusRawMessage(messageWithBadInstanceId)

		require.Equal(t, 1, len(messagesHandler.history))
	})
}

func TestCacheMessagesFromTheFuture(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		instanceId := primitives.InstanceId(rand.Uint64())
		state := mocks.NewMockState().WithHeightView(10, 0)
		filter := rawmessagesfilter.NewConsensusMessageFilter(instanceId, primitives.MemberId("My MemberId"), testLogger(state.State), state.State)
		messagesHandler := NewTermMessagesHandlerMock()
		filter.ConsumeCacheMessages(messagesHandler)

		messageFromTheFuture := GeneratePreprepareMessage(instanceId, 11, 20, "Sender MemberId")
		messageFromThePresent := GeneratePreprepareMessage(instanceId, 10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.history))

		filter.HandleConsensusRawMessage(messageFromTheFuture)
		filter.HandleConsensusRawMessage(messageFromThePresent)

		require.Equal(t, 1, len(messagesHandler.history))
	})
}

func TestFilterMessagesWithMyMemberId(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		instanceId := primitives.InstanceId(rand.Uint64())
		state := mocks.NewMockState().WithHeightView(10, 0)
		filter := rawmessagesfilter.NewConsensusMessageFilter(instanceId, primitives.MemberId("My MemberId"), testLogger(state.State), state.State)
		messagesHandler := NewTermMessagesHandlerMock()
		filter.ConsumeCacheMessages(messagesHandler)

		badMessage := GeneratePreprepareMessage(instanceId, 11, 20, "My MemberId")
		goodMessage := GeneratePreprepareMessage(instanceId, 10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.history))

		filter.HandleConsensusRawMessage(badMessage)
		filter.HandleConsensusRawMessage(goodMessage)

		require.Equal(t, 1, len(messagesHandler.history))
	})
}
