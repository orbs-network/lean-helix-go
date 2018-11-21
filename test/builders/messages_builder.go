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
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block, CalculateBlockHash(block))
}

func APrepareMessage(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PrepareMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreatePrepareMessage(blockHeight, view, CalculateBlockHash(block))
}

func ACommitMessage(
	keyManager leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.CommitMessage {

	messageFactory := leanhelix.NewMessageFactory(keyManager)
	return messageFactory.CreateCommitMessage(blockHeight, view, CalculateBlockHash(block))
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

func AValidNewViewMessage(
	newLeader *Node,
	members []*Node,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.NewViewMessage {

	ppmFactory := leanhelix.NewMessageFactory(newLeader.KeyManager)
	ppmCB := ppmFactory.CreatePreprepareMessageContentBuilder(blockHeight, view, block, CalculateBlockHash(block))

	var votes []*leanhelix.ViewChangeMessageContentBuilder
	for _, voter := range members {
		messageFactory := leanhelix.NewMessageFactory(voter.KeyManager)
		vcmCB := messageFactory.CreateViewChangeMessageContentBuilder(blockHeight, view, nil)
		votes = append(votes, vcmCB)
	}

	return ANewViewMessage(newLeader.KeyManager, blockHeight, view, ppmCB, votes, block)
}
