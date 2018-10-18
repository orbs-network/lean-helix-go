package leanhelix

type MessageReceiver interface {
	OnReceivePreprepareMessage(ppm PreprepareMessage) error
	OnReceivePrepareMessage(pm PrepareMessage) error
	OnReceiveCommitMessage(cm CommitMessage) error
	OnReceiveViewChangeMessage(vcm ViewChangeMessage) error
	OnReceiveNewViewMessage(nvm NewViewMessage) error
}

type MessageReceiverImpl struct {
}

func (rec *MessageReceiverImpl) OnReceivePreprepareMessage(ppm PreprepareMessage) error {
	panic("Where is TS impl?")
}

func (rec *MessageReceiverImpl) OnReceivePrepareMessage(pm PrepareMessage) error {
	panic("implement me")
}

func (rec *MessageReceiverImpl) OnReceiveCommitMessage(cm CommitMessage) error {
	panic("implement me")
}

func (rec *MessageReceiverImpl) OnReceiveViewChangeMessage(vcm ViewChangeMessage) error {
	panic("implement me")
}

func (rec *MessageReceiverImpl) OnReceiveNewViewMessage(nvm NewViewMessage) error {
	panic("implement me")
}
