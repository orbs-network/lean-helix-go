package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
)

func APreprepareMessage(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PreprepareMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block)
}

func APrepareMessage(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PrepareMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreatePrepareMessage(blockHeight, view, block.BlockHash())
}

func ACommitMessage(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.CommitMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreateCommitMessage(blockHeight, view, block.BlockHash())
}

func AViewChangeMessage(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	preparedMessages *leanhelix.PreparedMessages) *leanhelix.ViewChangeMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreateViewChangeMessage(blockHeight, view, preparedMessages)
}

func ANewViewMessage(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	ppContentBuilder *leanhelix.PreprepareContentBuilder,
	confirmations []*leanhelix.ViewChangeMessageContentBuilder,
	block leanhelix.Block) *leanhelix.NewViewMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreateNewViewMessage(blockHeight, view, ppContentBuilder, confirmations, block)
}
