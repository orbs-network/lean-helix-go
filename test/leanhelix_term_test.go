package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

func GeneratePreprepareMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := builders.CreateBlock(leanhelix.GenesisBlock)
	return builders.APreprepareMessage(keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GeneratePrepareMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := builders.CreateBlock(leanhelix.GenesisBlock)
	return builders.APrepareMessage(keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GenerateCommitMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := builders.CreateBlock(leanhelix.GenesisBlock)
	return builders.ACommitMessage(keyManager, senderMemberId, blockHeight, view, block).ToConsensusRawMessage()
}

func GenerateViewChangeMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	return builders.AViewChangeMessage(keyManager, senderMemberId, blockHeight, view, nil).ToConsensusRawMessage()
}

func GenerateNewViewMessage(blockHeight primitives.BlockHeight, view primitives.View, senderMemberIdStr string) *leanhelix.ConsensusRawMessage {
	senderMemberId := primitives.MemberId(senderMemberIdStr)
	keyManager := mocks.NewMockKeyManager(senderMemberId)
	block := builders.CreateBlock(leanhelix.GenesisBlock)
	return builders.ANewViewMessage(keyManager, senderMemberId, blockHeight, view, nil, nil, block).ToConsensusRawMessage()

}

//func TestProcessingAMessage(t *testing.T) {
//	WithContext(func(ctx context.Context) {
//		leanHelixTerm := NewLeanHelixTerm(ctx)
//		leanHelixTerm.
//
//		ppm := GeneratePreprepareMessage(10, 20, "Sender MemberId")
//		pm := GeneratePrepareMessage(10, 20, "Sender MemberId")
//		cm := GenerateCommitMessage(10, 20, "Sender MemberId")
//		vcm := GenerateViewChangeMessage(10, 20, "Sender MemberId")
//		nvm := GenerateNewViewMessage(10, 20, "Sender MemberId")
//
//		require.Equal(t, 0, len(messagesHandler.HistoryPP))
//		require.Equal(t, 0, len(messagesHandler.HistoryP))
//		require.Equal(t, 0, len(messagesHandler.HistoryC))
//		require.Equal(t, 0, len(messagesHandler.HistoryNV))
//		require.Equal(t, 0, len(messagesHandler.HistoryVC))
//
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(ppm))
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(pm))
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(cm))
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(vcm))
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(nvm))
//
//		require.Equal(t, 1, len(messagesHandler.HistoryPP))
//		require.Equal(t, 1, len(messagesHandler.HistoryP))
//		require.Equal(t, 1, len(messagesHandler.HistoryC))
//		require.Equal(t, 1, len(messagesHandler.HistoryNV))
//		require.Equal(t, 1, len(messagesHandler.HistoryVC))
//	})
//}

//func TestNotSendingMessagesWhenTheHandlerWasNotSet(t *testing.T) {
//	test.WithContext(func(ctx context.Context) {
//		leanHelixTerm := NewLeanHelixTerm(nil)
//
//		ppm := GeneratePreprepareMessage(10, 20, "Sender MemberId")
//		pm := GeneratePrepareMessage(10, 20, "Sender MemberId")
//		cm := GenerateCommitMessage(10, 20, "Sender MemberId")
//		vcm := GenerateViewChangeMessage(10, 20, "Sender MemberId")
//		nvm := GenerateNewViewMessage(10, 20, "Sender MemberId")
//
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(ppm))
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(pm))
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(cm))
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(vcm))
//		leanHelixTerm.HandleConsensusMessage(ctx, ToConsensusMessage(nvm))
//
//		// expect that we don't panic
//	})
//}
