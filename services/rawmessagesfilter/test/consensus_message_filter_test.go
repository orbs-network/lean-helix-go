package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/rawmessagesfilter"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func testLogger() interfaces.Logger {
	return logger.NewSilentLogger()
}

func GeneratePreprepareMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.APreprepareMessage(keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GeneratePrepareMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.APrepareMessage(keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GenerateCommitMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.ACommitMessage(keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GenerateViewChangeMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	return builders.AViewChangeMessage(keyManager, senderMemberId, blockHeight, view, nil).ToConsensusRawMessage()
}

func GenerateNewViewMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.ANewViewMessage(keyManager, senderMemberId, blockHeight, view, nil, nil, block).ToConsensusRawMessage()

}

func TestGettingAMessage(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := rawmessagesfilter.NewConsensusMessageFilter(primitives.MemberId("My MemberId"), testLogger())
		messagesHandler := NewTermMessagesHandlerMock()
		filter.SetBlockHeight(ctx, 10, messagesHandler)

		ppm := GeneratePreprepareMessage(10, 20, "Sender MemberId")
		pm := GeneratePrepareMessage(10, 20, "Sender MemberId")
		cm := GenerateCommitMessage(10, 20, "Sender MemberId")
		vcm := GenerateViewChangeMessage(10, 20, "Sender MemberId")
		nvm := GenerateNewViewMessage(10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.history))

		filter.HandleConsensusRawMessage(ctx, ppm)
		filter.HandleConsensusRawMessage(ctx, pm)
		filter.HandleConsensusRawMessage(ctx, cm)
		filter.HandleConsensusRawMessage(ctx, vcm)
		filter.HandleConsensusRawMessage(ctx, nvm)

		require.Equal(t, 5, len(messagesHandler.history))
	})
}

func TestFilterMessagesFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := rawmessagesfilter.NewConsensusMessageFilter(primitives.MemberId("My MemberId"), testLogger())
		messagesHandler := NewTermMessagesHandlerMock()
		filter.SetBlockHeight(ctx, 10, messagesHandler)

		messageFromThePast := GeneratePreprepareMessage(9, 20, "Sender MemberId")
		messageFromThePresent := GeneratePreprepareMessage(10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.history))

		filter.HandleConsensusRawMessage(ctx, messageFromThePast)
		filter.HandleConsensusRawMessage(ctx, messageFromThePresent)

		require.Equal(t, 1, len(messagesHandler.history))
	})
}

func TestCacheMessagesFromTheFuture(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := rawmessagesfilter.NewConsensusMessageFilter(primitives.MemberId("My MemberId"), testLogger())
		messagesHandler := NewTermMessagesHandlerMock()
		filter.SetBlockHeight(ctx, 10, messagesHandler)

		messageFromTheFuture := GeneratePreprepareMessage(11, 20, "Sender MemberId")
		messageFromThePresent := GeneratePreprepareMessage(10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.history))

		filter.HandleConsensusRawMessage(ctx, messageFromTheFuture)
		filter.HandleConsensusRawMessage(ctx, messageFromThePresent)

		require.Equal(t, 1, len(messagesHandler.history))
	})
}

func TestFilterMessagesWithMyMemberId(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := rawmessagesfilter.NewConsensusMessageFilter(primitives.MemberId("My MemberId"), testLogger())
		messagesHandler := NewTermMessagesHandlerMock()
		filter.SetBlockHeight(ctx, 10, messagesHandler)

		badMessage := GeneratePreprepareMessage(11, 20, "My MemberId")
		goodMessage := GeneratePreprepareMessage(10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.history))

		filter.HandleConsensusRawMessage(ctx, badMessage)
		filter.HandleConsensusRawMessage(ctx, goodMessage)

		require.Equal(t, 1, len(messagesHandler.history))
	})
}
