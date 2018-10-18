package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/primitives"
)

type NetworkMessageFilter interface {
	OnGossipMessage(rawMessage ConsensusRawMessage)
}

type NetworkMessageFilterImpl struct {
	blockHeight  primitives.BlockHeight
	MyPublicKey  primitives.Ed25519PublicKey
	messageCache []ConsensusMessage
	comm         NetworkCommunication
	Receiver     MessageReceiver
}

func NewNetworkMessageFilter(comm NetworkCommunication, blockHeight primitives.BlockHeight, myPublicKey primitives.Ed25519PublicKey, messageReceiver MessageReceiver) NetworkMessageFilter {
	return &NetworkMessageFilterImpl{
		blockHeight:  blockHeight,
		MyPublicKey:  myPublicKey,
		messageCache: make([]ConsensusMessage, 0, 10),
		Receiver:     messageReceiver,
		comm:         comm,
	}
}

// Entry point of messages for consensus messages

func (filter *NetworkMessageFilterImpl) OnGossipMessage(rawMessage ConsensusRawMessage) {

	header := ConsensusMessageHeaderReader(rawMessage.Header())
	var message ConsensusMessage

	switch header.MessageType() {
	case LEAN_HELIX_PREPREPARE:
		content := BlockRefContentReader(rawMessage.Content())
		message = &PreprepareMessageImpl{
			MyContent: content,
			MyBlock:   rawMessage.Block(),
		}

	case LEAN_HELIX_PREPARE:
		content := BlockRefContentReader(rawMessage.Content())
		message = &PrepareMessageImpl{
			MyContent: content,
		}

	case LEAN_HELIX_COMMIT:
		content := BlockRefContentReader(rawMessage.Content())
		message = &CommitMessageImpl{
			MyContent: content,
		}
	case LEAN_HELIX_VIEW_CHANGE:
		content := ViewChangeMessageContentReader(rawMessage.Content())
		message = &ViewChangeMessageImpl{
			MyContent: content,
			MyBlock:   rawMessage.Block(),
		}

	case LEAN_HELIX_NEW_VIEW:
		content := NewViewMessageContentReader(rawMessage.Content())
		message = &NewViewMessageImpl{
			MyContent: content,
			MyBlock:   rawMessage.Block(),
		}
	}

	if !filter.acceptMessage(message) {
		return
	}

	if message.BlockHeight() > filter.blockHeight {
		filter.pushToCache(message)
		return
	}
	filter.ProcessGossipMessage(message)
}

func (filter *NetworkMessageFilterImpl) pushToCache(message ConsensusMessage) {
	filter.messageCache = append(filter.messageCache, message)
}

func (filter *NetworkMessageFilterImpl) acceptMessage(message ConsensusMessage) bool {

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

// TODO Callback

func (filter *NetworkMessageFilterImpl) ConsumeCacheMessage() {

}

func (filter *NetworkMessageFilterImpl) ProcessGossipMessage(consensusMessage ConsensusMessage) {

	switch message := consensusMessage.(type) {
	case *PreprepareMessageImpl:
		filter.Receiver.OnReceivePreprepareMessage(message)
	case *PrepareMessageImpl:
		filter.Receiver.OnReceivePrepareMessage(message)
		// Send message to MessageReceiver
	}
}

func (filter *NetworkMessageFilterImpl) SetBlockHeight(blockHeight primitives.BlockHeight, receiver MessageReceiver) {
	filter.blockHeight = blockHeight
	filter.Receiver = receiver
	filter.consumeCacheMessage()
}

func (filter *NetworkMessageFilterImpl) consumeCacheMessage() {
	unconsumed := make([]ConsensusMessage, 0, 1)
	for _, msg := range filter.messageCache {
		if msg.BlockHeight().Equal(filter.blockHeight) {
			filter.ProcessGossipMessage(msg)
		} else {
			unconsumed = append(unconsumed, msg)
		}
	}
	filter.messageCache = unconsumed
}
