package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

func APreprepareMessage(
	keyManager leanhelix.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PreprepareMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager, senderMemberId)
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block, CalculateBlockHash(block))
}

func APrepareMessage(
	keyManager leanhelix.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PrepareMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager, senderMemberId)
	return messageFactory.CreatePrepareMessage(blockHeight, view, CalculateBlockHash(block))
}

func ACommitMessage(
	keyManager leanhelix.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.CommitMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager, senderMemberId)
	return messageFactory.CreateCommitMessage(blockHeight, view, CalculateBlockHash(block))
}

func AViewChangeMessage(
	keyManager leanhelix.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	preparedMessages *leanhelix.PreparedMessages) *leanhelix.ViewChangeMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager, senderMemberId)
	return messageFactory.CreateViewChangeMessage(blockHeight, view, preparedMessages)
}

func ANewViewMessage(
	keyManager leanhelix.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	ppContentBuilder *protocol.PreprepareContentBuilder,
	confirmations []*protocol.ViewChangeMessageContentBuilder,
	block leanhelix.Block) *leanhelix.NewViewMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager, senderMemberId)
	return messageFactory.CreateNewViewMessage(blockHeight, view, ppContentBuilder, confirmations, block)
}
