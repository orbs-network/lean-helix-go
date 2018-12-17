package consensusmessagefilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func testLogger() leanhelix.Logger {
	return leanhelix.NewSilentLogger()
}

func GeneratePreprepareMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := builders.NewMockKeyManager(senderMemberId)
	block := builders.CreateBlock(builders.GenesisBlock)
	return builders.APreprepareMessage(keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GeneratePrepareMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := builders.NewMockKeyManager(senderMemberId)
	block := builders.CreateBlock(builders.GenesisBlock)
	return builders.APrepareMessage(keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GenerateCommitMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := builders.NewMockKeyManager(senderMemberId)
	block := builders.CreateBlock(builders.GenesisBlock)
	return builders.ACommitMessage(keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GenerateViewChangeMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := builders.NewMockKeyManager(senderMemberId)
	return builders.AViewChangeMessage(keyManager, senderMemberId, blockHeight, view, nil).ToConsensusRawMessage()
}

func GenerateNewViewMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := builders.NewMockKeyManager(senderMemberId)
	block := builders.CreateBlock(builders.GenesisBlock)
	return builders.ANewViewMessage(keyManager, senderMemberId, blockHeight, view, nil, nil, block).ToConsensusRawMessage()

}

func TestGettingAMessage(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.MemberId("My MemberId"), testLogger())
		termMessagesHandler := NewTermMessagesHandlerMock()
		filter.SetBlockHeight(ctx, 10, termMessagesHandler)

		ppm := GeneratePreprepareMessage(10, 20, "Sender MemberId")
		pm := GeneratePrepareMessage(10, 20, "Sender MemberId")
		cm := GenerateCommitMessage(10, 20, "Sender MemberId")
		vcm := GenerateViewChangeMessage(10, 20, "Sender MemberId")
		nvm := GenerateNewViewMessage(10, 20, "Sender MemberId")

		require.Equal(t, 0, len(termMessagesHandler.historyPP))
		require.Equal(t, 0, len(termMessagesHandler.historyP))
		require.Equal(t, 0, len(termMessagesHandler.historyC))
		require.Equal(t, 0, len(termMessagesHandler.historyNV))
		require.Equal(t, 0, len(termMessagesHandler.historyVC))

		filter.GossipMessageReceived(ctx, ppm)
		filter.GossipMessageReceived(ctx, pm)
		filter.GossipMessageReceived(ctx, cm)
		filter.GossipMessageReceived(ctx, vcm)
		filter.GossipMessageReceived(ctx, nvm)

		require.Equal(t, 1, len(termMessagesHandler.historyPP))
		require.Equal(t, 1, len(termMessagesHandler.historyP))
		require.Equal(t, 1, len(termMessagesHandler.historyC))
		require.Equal(t, 1, len(termMessagesHandler.historyNV))
		require.Equal(t, 1, len(termMessagesHandler.historyVC))
	})
}

func TestFilterMessagesFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.MemberId("My MemberId"), testLogger())
		termMessagesHandler := NewTermMessagesHandlerMock()
		filter.SetBlockHeight(ctx, 10, termMessagesHandler)

		messageFromThePast := GeneratePreprepareMessage(9, 20, "Sender MemberId")
		messageFromThePresent := GeneratePreprepareMessage(10, 20, "Sender MemberId")

		require.Equal(t, 0, len(termMessagesHandler.historyPP))

		filter.GossipMessageReceived(ctx, messageFromThePast)
		filter.GossipMessageReceived(ctx, messageFromThePresent)

		require.Equal(t, 1, len(termMessagesHandler.historyPP))
	})
}

func TestCacheMessagesFromTheFuture(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.MemberId("My MemberId"), testLogger())
		termMessagesHandler := NewTermMessagesHandlerMock()
		filter.SetBlockHeight(ctx, 10, termMessagesHandler)

		messageFromTheFuture := GeneratePreprepareMessage(11, 20, "Sender MemberId")
		messageFromThePresent := GeneratePreprepareMessage(10, 20, "Sender MemberId")

		require.Equal(t, 0, len(termMessagesHandler.historyPP))

		filter.GossipMessageReceived(ctx, messageFromTheFuture)
		filter.GossipMessageReceived(ctx, messageFromThePresent)

		require.Equal(t, 1, len(termMessagesHandler.historyPP))
	})
}

func TestFilterMessagesWithMyMemberId(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.MemberId("My MemberId"), testLogger())
		termMessagesHandler := NewTermMessagesHandlerMock()
		filter.SetBlockHeight(ctx, 10, termMessagesHandler)

		badMessage := GeneratePreprepareMessage(11, 20, "My MemberId")
		goodMessage := GeneratePreprepareMessage(10, 20, "Sender MemberId")

		require.Equal(t, 0, len(termMessagesHandler.historyPP))

		filter.GossipMessageReceived(ctx, badMessage)
		filter.GossipMessageReceived(ctx, goodMessage)

		require.Equal(t, 1, len(termMessagesHandler.historyPP))
	})
}
