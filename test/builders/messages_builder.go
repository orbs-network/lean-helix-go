package builders

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

func APreprepareMessage(
	instanceId primitives.InstanceId,
	keyManager interfaces.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block interfaces.Block) *interfaces.PreprepareMessage {

	messageFactory := messagesfactory.NewMessageFactory(instanceId, keyManager, senderMemberId, 0)
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block, mocks.CalculateBlockHash(block))
}

func APrepareMessage(
	instanceId primitives.InstanceId,
	keyManager interfaces.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block interfaces.Block) *interfaces.PrepareMessage {

	messageFactory := messagesfactory.NewMessageFactory(instanceId, keyManager, senderMemberId, 0)
	return messageFactory.CreatePrepareMessage(blockHeight, view, mocks.CalculateBlockHash(block))
}

func ACommitMessage(
	instanceId primitives.InstanceId,
	keyManager interfaces.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block interfaces.Block,
	randomSeed uint64) *interfaces.CommitMessage {

	messageFactory := messagesfactory.NewMessageFactory(instanceId, keyManager, senderMemberId, randomSeed)
	return messageFactory.CreateCommitMessage(blockHeight, view, mocks.CalculateBlockHash(block))
}

func AViewChangeMessage(
	instanceId primitives.InstanceId,
	keyManager interfaces.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	preparedMessages *preparedmessages.PreparedMessages) *interfaces.ViewChangeMessage {

	messageFactory := messagesfactory.NewMessageFactory(instanceId, keyManager, senderMemberId, 0)
	return messageFactory.CreateViewChangeMessage(blockHeight, view, preparedMessages)
}

func ANewViewMessage(
	instanceId primitives.InstanceId,
	keyManager interfaces.KeyManager,
	senderMemberId primitives.MemberId,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	ppContentBuilder *protocol.PreprepareContentBuilder,
	confirmations []*protocol.ViewChangeMessageContentBuilder,
	block interfaces.Block) *interfaces.NewViewMessage {

	messageFactory := messagesfactory.NewMessageFactory(instanceId, keyManager, senderMemberId, 0)
	return messageFactory.CreateNewViewMessage(blockHeight, view, ppContentBuilder, confirmations, block)
}
