package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

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

func TestProcessingAMessage(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		messagesHandler := mocks.NewTermMessagesHandlerMock()
		consensusMessagesFilter := NewConsensusMessagesFilter(messagesHandler)

		ppm := GeneratePreprepareMessage(10, 20, "Sender MemberId")
		pm := GeneratePrepareMessage(10, 20, "Sender MemberId")
		cm := GenerateCommitMessage(10, 20, "Sender MemberId")
		vcm := GenerateViewChangeMessage(10, 20, "Sender MemberId")
		nvm := GenerateNewViewMessage(10, 20, "Sender MemberId")

		require.Equal(t, 0, len(messagesHandler.HistoryPP))
		require.Equal(t, 0, len(messagesHandler.HistoryP))
		require.Equal(t, 0, len(messagesHandler.HistoryC))
		require.Equal(t, 0, len(messagesHandler.HistoryNV))
		require.Equal(t, 0, len(messagesHandler.HistoryVC))

		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(ppm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(pm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(cm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(vcm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(nvm))

		require.Equal(t, 1, len(messagesHandler.HistoryPP))
		require.Equal(t, 1, len(messagesHandler.HistoryP))
		require.Equal(t, 1, len(messagesHandler.HistoryC))
		require.Equal(t, 1, len(messagesHandler.HistoryNV))
		require.Equal(t, 1, len(messagesHandler.HistoryVC))
	})
}

func TestNotSendingMessagesWhenTheHandlerWasNotSet(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		consensusMessagesFilter := NewConsensusMessagesFilter(nil)

		ppm := GeneratePreprepareMessage(10, 20, "Sender MemberId")
		pm := GeneratePrepareMessage(10, 20, "Sender MemberId")
		cm := GenerateCommitMessage(10, 20, "Sender MemberId")
		vcm := GenerateViewChangeMessage(10, 20, "Sender MemberId")
		nvm := GenerateNewViewMessage(10, 20, "Sender MemberId")

		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(ppm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(pm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(cm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(vcm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(nvm))

		// expect that we don't panic
	})
}
