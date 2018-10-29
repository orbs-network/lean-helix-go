package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type NetworkMessageFilter struct {
	blockHeight  primitives.BlockHeight
	MyPublicKey  primitives.Ed25519PublicKey
	messageCache []ConsensusMessage
	comm         NetworkCommunication
	receiver     MessageReceiver
}

func NewNetworkMessageFilter(comm NetworkCommunication, myPublicKey primitives.Ed25519PublicKey) *NetworkMessageFilter {
	res := &NetworkMessageFilter{
		blockHeight:  0,
		MyPublicKey:  myPublicKey,
		messageCache: make([]ConsensusMessage, 0, 10),
		comm:         comm,
	}

	res.comm.RegisterOnMessage(res.OnGossipMessage)
	return res
}

// Entry point of messages for consensus messages
func (filter *NetworkMessageFilter) OnGossipMessage(ctx context.Context, rawMessage ConsensusRawMessage) {

	message := rawMessage.ToConsensusMessage()

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

	if !filter.AllowedToReceiveMessageFrom(senderPublicKey) {
		return false
	}

	if message.BlockHeight() < filter.blockHeight {
		return false
	}
	return true
}

func (filter *NetworkMessageFilter) AllowedToReceiveMessageFrom(senderPublicKey primitives.Ed25519PublicKey) bool {
	if filter.isSameNodeAs(senderPublicKey) {
		return false
	}

	return filter.comm.IsMember(senderPublicKey)
}

func (filter *NetworkMessageFilter) isSameNodeAs(sender primitives.Ed25519PublicKey) bool {
	myPublicKeyStr := filter.MyPublicKey.String()
	senderPublicKeyStr := sender.String()
	return myPublicKeyStr == senderPublicKeyStr
}

func (filter *NetworkMessageFilter) ProcessGossipMessage(ctx context.Context, consensusMessage ConsensusMessage) error {

	if filter.receiver == nil {
		panic("no receiver")
	}
	switch message := consensusMessage.(type) {
	case *PreprepareMessage:
		filter.receiver.OnReceivePreprepare(ctx, message) // filter.receiver is actually the LeanHelixTerm
	case *PrepareMessage:
		filter.receiver.OnReceivePrepare(ctx, message)
	case *CommitMessage:
		filter.receiver.OnReceiveCommit(ctx, message)
	case *ViewChangeMessage:
		filter.receiver.OnReceiveViewChange(ctx, message)
	case *NewViewMessage:
		filter.receiver.OnReceiveNewView(ctx, message)
	default:
		return fmt.Errorf("unknown message type: %T", consensusMessage)
	}
	return nil
}

func (filter *NetworkMessageFilter) SetBlockHeight(ctx context.Context, blockHeight primitives.BlockHeight, messageReceiver MessageReceiver) {
	filter.blockHeight = blockHeight
	filter.receiver = messageReceiver
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
