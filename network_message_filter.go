package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type NetworkMessageFilter struct {
	blockHeight  primitives.BlockHeight
	MyPublicKey  primitives.Ed25519PublicKey
	messageCache []ConsensusMessage
	comm         NetworkCommunication
	Receiver     MessageReceiver
}

func NewNetworkMessageFilter(comm NetworkCommunication, blockHeight primitives.BlockHeight, myPublicKey primitives.Ed25519PublicKey, messageReceiver MessageReceiver) *NetworkMessageFilter {
	return &NetworkMessageFilter{
		blockHeight:  blockHeight,
		MyPublicKey:  myPublicKey,
		messageCache: make([]ConsensusMessage, 0, 10),
		Receiver:     messageReceiver,
		comm:         comm,
	}
}

// Entry point of messages for consensus messages

func (filter *NetworkMessageFilter) OnGossipMessage(ctx context.Context, rawMessage ConsensusRawMessage) {

	var message ConsensusMessage
	switch rawMessage.MessageType() {
	case LEAN_HELIX_PREPREPARE:
		content := PreprepareContentReader(rawMessage.Content())
		message = &PreprepareMessage{
			content: content,
			block:   rawMessage.Block(),
		}

	case LEAN_HELIX_PREPARE:
		content := PrepareContentReader(rawMessage.Content())
		message = &PrepareMessage{
			content: content,
		}

	case LEAN_HELIX_COMMIT:
		content := CommitContentReader(rawMessage.Content())
		message = &CommitMessage{
			content: content,
		}
	case LEAN_HELIX_VIEW_CHANGE:
		content := ViewChangeMessageContentReader(rawMessage.Content())
		message = &ViewChangeMessage{
			content: content,
			block:   rawMessage.Block(),
		}

	case LEAN_HELIX_NEW_VIEW:
		content := NewViewMessageContentReader(rawMessage.Content())
		message = &NewViewMessage{
			content: content,
			block:   rawMessage.Block(),
		}
	}

	if !filter.acceptMessage(message) {
		return
	}

	if message.BlockHeight() > filter.blockHeight {
		filter.pushToCache(message)
		return
	}
	filter.ProcessGossipMessage(ctx, message)
}

func (filter *NetworkMessageFilter) pushToCache(message ConsensusMessage) {
	filter.messageCache = append(filter.messageCache, message)
}

func (filter *NetworkMessageFilter) acceptMessage(message ConsensusMessage) bool {

	senderPublicKey := message.SenderPublicKey()
	if senderPublicKey.Equal(filter.MyPublicKey) {
		return false
	}

	if !filter.comm.IsMember(senderPublicKey) {
		return false
	}

	if message.BlockHeight() < filter.blockHeight {
		return false
	}
	return true
}

func (filter *NetworkMessageFilter) ProcessGossipMessage(ctx context.Context, consensusMessage ConsensusMessage) {

	switch message := consensusMessage.(type) {
	case *PreprepareMessage:
		filter.Receiver.OnReceivePreprepare(ctx, message)
	case *PrepareMessage:
		filter.Receiver.OnReceivePrepare(ctx, message)
		// Send message to MessageReceiver
	}
}

func (filter *NetworkMessageFilter) SetBlockHeight(ctx context.Context, blockHeight primitives.BlockHeight, receiver MessageReceiver) {
	filter.blockHeight = blockHeight
	filter.Receiver = receiver
	filter.consumeCacheMessage(ctx)
}

func (filter *NetworkMessageFilter) consumeCacheMessage(ctx context.Context) {
	unconsumed := make([]ConsensusMessage, 0, 1)
	for _, msg := range filter.messageCache {
		if msg.BlockHeight().Equal(filter.blockHeight) {
			filter.ProcessGossipMessage(ctx, msg)
		} else {
			unconsumed = append(unconsumed, msg)
		}
	}
	filter.messageCache = unconsumed
}
