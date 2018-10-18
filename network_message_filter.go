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
	messageCache []MessageWithSenderAndBlockHeight
	netComm      NetworkCommunication
	Receiver     MessageReceiver
}

func NewNetworkMessageFilter(blockHeight primitives.BlockHeight, myPublicKey primitives.Ed25519PublicKey, messageReceiver MessageReceiver, netComm NetworkCommunication) NetworkMessageFilter {
	return &NetworkMessageFilterImpl{
		blockHeight:  blockHeight,
		MyPublicKey:  myPublicKey,
		messageCache: make([]MessageWithSenderAndBlockHeight, 0, 10),
		Receiver:     messageReceiver,
		netComm:      netComm,
	}
}

// Entry point of messages for consensus messages

func (filter *NetworkMessageFilterImpl) OnGossipMessage(rawMessage ConsensusRawMessage) {

	header := ConsensusMessageHeaderReader(rawMessage.Header())
	var message ConsensusMessage

	switch header.MessageType() {
	case LEAN_HELIX_PREPREPARE:
		content := PreprepareMessageContentReader(rawMessage.Content())
		message = &PreprepareMessageImpl{
			Content: content,
			block: rawMessage.Block(),
		}

	case LEAN_HELIX_PREPARE:
		content := PrepareMessageContentReader(rawMessage.Content())
		message = &PrepareMessageImpl{
			Content: content,
		}

	case LEAN_HELIX_COMMIT:
		content := CommitMessageContentReader(rawMessage.Content())
		message = &CommitMessageImpl{
			Content: content,
		}
	case LEAN_HELIX_VIEW_CHANGE:
		content := ViewChangeMessageContentReader(rawMessage.Content())
		message = &ViewChangeMessageImpl{
			Content: content,
			block: rawMessage.Block(),
		}

	case LEAN_HELIX_NEW_VIEW:
		content := NewViewMessageContentReader(rawMessage.Content())
		message = &NewViewMessageImpl{
			Content: content,
			block: rawMessage.Block(),
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

func (filter *NetworkMessageFilterImpl) acceptMessage(message MessageWithSenderAndBlockHeight) bool {

	senderPublicKey := message.SenderPublicKey()
	if senderPublicKey.Equal(filter.MyPublicKey) {
		return false
	}

	if !filter.netComm.IsMember(senderPublicKey) {
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

func (filter *NetworkMessageFilterImpl) ProcessGossipMessage(message MessageTransporter) {
func (filter* NetworkMessageFilter) ProcessGossipMessage(message interface{}) {

	switch message.(type) {
	case PreprepareMessageImpl {
		filter.Receiver.OnReceivePreprepareMessage(message)

	}

		// Send message to MessageReceiver
	}
}

//public setBlockHeight(blockHeight: number, messagesHandler: MessagesHandler) {
//this.blockHeight = blockHeight;
//this.PBFTMessagesHandler = messagesHandler;
//this.consumeCacheMessages();
//}
