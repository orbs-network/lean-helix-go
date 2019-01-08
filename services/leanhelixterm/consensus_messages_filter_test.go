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

func GeneratePreprepareMessage(networkId primitives.NetworkId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.APreprepareMessage(networkId, keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GeneratePrepareMessage(networkId primitives.NetworkId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.APrepareMessage(networkId, keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GenerateCommitMessage(networkId primitives.NetworkId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string, randomSeed uint64) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.ACommitMessage(networkId, keyManager, senderMemberId, blockHeight, view, block, randomSeed).ToConsensusRawMessage()
}

func GenerateViewChangeMessage(networkId primitives.NetworkId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	return builders.AViewChangeMessage(networkId, keyManager, senderMemberId, blockHeight, view, nil).ToConsensusRawMessage()
}

func GenerateNewViewMessage(networkId primitives.NetworkId, blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *interfaces.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := mocks.ABlock(interfaces.GenesisBlock)
	return builders.ANewViewMessage(networkId, keyManager, senderMemberId, blockHeight, view, nil, nil, block).ToConsensusRawMessage()

}

func TestProcessingAMessage(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		networkId := primitives.NetworkId(rand.Uint64())
		messagesHandler := mocks.NewTermMessagesHandlerMock()
		keyManager := mocks.NewMockKeyManager(primitives.MemberId("My ID"))
		consensusMessagesFilter := NewConsensusMessagesFilter(messagesHandler, keyManager, 99)

		ppm := GeneratePreprepareMessage(networkId, 10, 20, "Sender MemberId")
		pm := GeneratePrepareMessage(networkId, 10, 20, "Sender MemberId")
		cm := GenerateCommitMessage(networkId, 10, 20, "Sender MemberId", 99)
		vcm := GenerateViewChangeMessage(networkId, 10, 20, "Sender MemberId")
		nvm := GenerateNewViewMessage(networkId, 10, 20, "Sender MemberId")

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

func TestFilteringACommitWithBadSeed(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		networkId := primitives.NetworkId(rand.Uint64())
		messagesHandler := mocks.NewTermMessagesHandlerMock()
		keyManager := mocks.NewMockKeyManager(primitives.MemberId("My ID"))
		consensusMessagesFilter := NewConsensusMessagesFilter(messagesHandler, keyManager, 99)

		goodCommit := GenerateCommitMessage(networkId, 10, 20, "Sender MemberId", 99)
		badCommit := GenerateCommitMessage(networkId, 10, 20, "Sender MemberId", 666)

		require.Equal(t, 0, len(messagesHandler.HistoryC))

		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(goodCommit))
		require.Equal(t, 1, len(messagesHandler.HistoryC))

		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(badCommit))
		require.Equal(t, 1, len(messagesHandler.HistoryC)) // still on 1
	})
}

func TestNotSendingMessagesWhenTheHandlerWasNotSet(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		networkId := primitives.NetworkId(rand.Uint64())
		keyManager := mocks.NewMockKeyManager(primitives.MemberId("My ID"))
		consensusMessagesFilter := NewConsensusMessagesFilter(nil, keyManager, 99)

		ppm := GeneratePreprepareMessage(networkId, 10, 20, "Sender MemberId")
		pm := GeneratePrepareMessage(networkId, 10, 20, "Sender MemberId")
		cm := GenerateCommitMessage(networkId, 10, 20, "Sender MemberId", 99)
		vcm := GenerateViewChangeMessage(networkId, 10, 20, "Sender MemberId")
		nvm := GenerateNewViewMessage(networkId, 10, 20, "Sender MemberId")

		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(ppm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(pm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(cm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(vcm))
		consensusMessagesFilter.HandleConsensusMessage(ctx, interfaces.ToConsensusMessage(nvm))

		// expect that we don't panic
	})
}
