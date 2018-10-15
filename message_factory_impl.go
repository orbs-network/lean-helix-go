package leanhelix

import . "github.com/orbs-network/lean-helix-go/primitives"

// This is the ORBS side

type MessageFactoryImpl struct {
	CalculateBlockHash func(Block) Uint256
	KeyManager
}

func (f *MessageFactoryImpl) CreatePreprepareMessage(blockHeight BlockHeight, view View, block Block) PreprepareMessage {

	header := &BlockRefBuilder{
		MessageType: LEAN_HELIX_PREPREPARE,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   Uint256(f.CalculateBlockHash(block)),
	}

	sig := Ed25519Sig(f.KeyManager.Sign(header.Build().Raw()))
	me := Ed25519PublicKey(f.KeyManager.MyPublicKey())
	sender := &SenderSignatureBuilder{
		SenderPublicKey: me,
		Signature:       sig,
	}

	ppmc := PreprepareMessageContentBuilder{
		SignedHeader: header,
		Sender:       sender,
	}

	ppm := &preprepareMessage{
		Content: ppmc.Build(),
		block:   block,
	}

	return ppm
}

func (f *MessageFactoryImpl) CreatePrepareMessage(blockHeight BlockHeight, view View, blockHash Uint256) PrepareMessage {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateCommitMessage(blockHeight BlockHeight, view View, blockHash Uint256) CommitMessage {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateViewChangeMessage(blockHeight BlockHeight, view View, preparedMessages *PreparedMessages) ViewChangeMessage {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateNewViewMessage(blockHeight BlockHeight, view View, ppm PreprepareMessage, confirmations []ViewChangeConfirmation) NewViewMessage {
	panic("implement me")
}

func NewMessageFactory(calculateBlockHash func(Block) Uint256, keyManager KeyManager) MessageFactory {
	return &MessageFactoryImpl{
		CalculateBlockHash: calculateBlockHash,
		KeyManager:         keyManager,
	}
}
