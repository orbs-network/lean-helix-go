package leanhelix

type MessageReceiver interface {
	OnReceive(message []byte) error
	OnReceiveWithBlock(message []byte, block Block) error
}

type MessageReceiverImpl struct {
}

func (rec *MessageReceiverImpl) OnReceive(message []byte) error {
	panic("implement me")
}

func (rec *MessageReceiverImpl) OnReceiveWithBlock(message []byte, block Block) error {
	panic("implement me")
}
