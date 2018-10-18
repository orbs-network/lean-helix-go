package leanhelix

import "github.com/orbs-network/lean-helix-go/primitives"

type NetworkMessageFilter struct {
	MyPublicKey primitives.Ed25519PublicKey
	OnGossipMessage func(message ConsensusRawMessage)
	Receiver MessageReceiver
}


// Entry point of messages for consensus messages

func (filter* NetworkMessageFilter) OnGossipMessage(rawMessage ConsensusRawMessage) {

	header := ConsensusMessageHeaderReader(rawMessage.Header())
	var message MessageWithSenderAndBlockHeight

	switch header.MessageType() {
	case LEAN_HELIX_PREPREPARE:
		content := PreprepareMessageContentReader(rawMessage.Content())
		message = &PreprepareMessageImpl{
			Content: content,
			MyBlock: rawMessage.Block(),
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
			MyBlock: rawMessage.Block(),
		}

	case LEAN_HELIX_NEW_VIEW:
		content := NewViewMessageContentReader(rawMessage.Content())
		message = &NewViewMessageImpl{
			Content: content,
			MyBlock: rawMessage.Block(),
		}
	}

	// TODO add conditions from NetworkMessageFilter.ts ...
	// TODO push to messagesCache
	// TODO Callback
}


func (filter* NetworkMessageFilter) ConsumeCacheMessage() {
}

func (filter* NetworkMessageFilter) ProcessGossipMessage(message MessageTransporter) {

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



