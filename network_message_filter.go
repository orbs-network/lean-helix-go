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
	Receiver     MessageReceiver
}

func NewNetworkMessageFilter(comm NetworkCommunication, blockHeight primitives.BlockHeight, myPublicKey primitives.Ed25519PublicKey, messageReceiver MessageReceiver) *NetworkMessageFilter {
	res := &NetworkMessageFilter{
		blockHeight:  blockHeight,
		MyPublicKey:  myPublicKey,
		messageCache: make([]ConsensusMessage, 0, 10),
		Receiver:     messageReceiver,
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

// TODO unit-test this
func (filter *NetworkMessageFilter) acceptMessage(message ConsensusMessage) bool {

	fmt.Printf("acceptMessage: type: %T\n", message)
	senderPublicKey := message.SenderPublicKey()
	if filter.isSameNodeAs(senderPublicKey) {
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

func (filter *NetworkMessageFilter) isSameNodeAs(sender primitives.Ed25519PublicKey) bool {
	myPublicKeyStr := filter.MyPublicKey.String()
	senderPublicKeyStr := sender.String()
	return myPublicKeyStr == senderPublicKeyStr
}

func (filter *NetworkMessageFilter) ProcessGossipMessage(ctx context.Context, consensusMessage ConsensusMessage) error {

	if filter.Receiver == nil {
		panic("no receiver")
	}
	switch message := consensusMessage.(type) {
	case *PreprepareMessage:
		filter.Receiver.OnReceivePreprepare(ctx, message) // filter.Receiver is actually the LeanHelixTerm
	case *PrepareMessage:
		filter.Receiver.OnReceivePrepare(ctx, message)
	case *CommitMessage:
		filter.Receiver.OnReceiveCommit(ctx, message)
	case *ViewChangeMessage:
		filter.Receiver.OnReceiveViewChange(ctx, message)
	case *NewViewMessage:
		filter.Receiver.OnReceiveNewView(ctx, message)
	default:
		return fmt.Errorf("unknown message type: %T", consensusMessage)
	}
	return nil
}

func (filter *NetworkMessageFilter) SetBlockHeight(ctx context.Context, blockHeight primitives.BlockHeight) {
	filter.blockHeight = blockHeight
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
