package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
)

func APrepreparedMessages(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PreprepareMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block)
}

func APreparedMessages(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PrepareMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreatePrepareMessage(blockHeight, view, block.BlockHash())
}

func ACommitMessages(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.CommitMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreateCommitMessage(blockHeight, view, block.BlockHash())
}

func AViewChangeMessages(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	preparedMessages *leanhelix.PreparedMessages) *leanhelix.ViewChangeMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreateViewChangeMessage(blockHeight, view, preparedMessages)
}
