package leanhelix

type MessageReceiver interface {
	OnReceive(message []byte) error
	OnReceiveWithBlock(message []byte, block Block) error
}

type MessageReceiverImpl struct {
}

func (rec *MessageReceiverImpl) OnReceive(rawMessage ConsensusRawMessage) error {

	message := toMessageTransporter(rawMessage)

	panic("implement me")
}
func toMessageTransporter(rawMessage ConsensusRawMessage) MessageTransporter {
	header := ConsensusMessageHeaderReader(rawMessage.Header())

	var message MessageTransporter

	switch header.MessageType() {
	case LEAN_HELIX_PREPREPARE:
		message = PreprepareMessageContentReader(rawMessage.Content())
	case LEAN_HELIX_PREPARE:
		message = PrepareMessageContentReader(rawMessage.Content())
	case LEAN_HELIX_COMMIT:
		message = CommitMessageContentReader(rawMessage.Content())
	case LEAN_HELIX_VIEW_CHANGE:
		message = ViewChangeMessageContentReader(rawMessage.Content())
	case LEAN_HELIX_NEW_VIEW:
		message = NewViewMessageContentReader(rawMessage.Content())

	}
	return message

}
