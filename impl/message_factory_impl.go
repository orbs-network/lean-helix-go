package impl

import "github.com/orbs-network/lean-helix-go"

// This is the ORBS side

type MessageFactoryImpl struct {
}

func (factory *MessageFactoryImpl) CreatePreprepareMessage(blockHeight leanhelix.BlockHeight, view leanhelix.View, block leanhelix.Block) leanhelix.PreprepareMessage {

	panic("implement me")
}

func (factory *MessageFactoryImpl) CreatePrepareMessage(blockHeight leanhelix.BlockHeight, view leanhelix.View, blockHash leanhelix.BlockHash) leanhelix.PrepareMessage {
	panic("implement me")
}

func (factory *MessageFactoryImpl) CreateCommitMessage(blockHeight leanhelix.BlockHeight, view leanhelix.View, blockHash leanhelix.BlockHash) leanhelix.CommitMessage {
	panic("implement me")
}

func (factory *MessageFactoryImpl) CreateViewChangeMessage(blockHeight leanhelix.BlockHeight, view leanhelix.View, preparedMessages []leanhelix.PreprepareMessage) leanhelix.ViewChangeMessage {
	panic("implement me")
}

func (factory *MessageFactoryImpl) CreateNewViewMessage(blockHeight leanhelix.BlockHeight, view leanhelix.View, preprepareMessage leanhelix.PreprepareMessage, viewChangeConfirmations []leanhelix.ViewChangeConfirmation) leanhelix.NewViewMessage {
	panic("implement me")
}
