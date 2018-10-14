package leanhelix

import . "github.com/orbs-network/lean-helix-go/primitives"

// This is the ORBS side

type MessageFactoryImpl struct {
	BlockUtils
	KeyManager
}

func (f *MessageFactoryImpl) CreatePreprepareMessage(blockHeight BlockHeight, view View, block Block) PreprepareMessage {

	header := &BlockRefBuilder{
		MessageType: LEAN_HELIX_PREPREPARE,
		BlockHeight: BlockHeight(blockHeight),
		View:        View(view),
		BlockHash:   Uint256(f.BlockUtils.CalculateBlockHash(block)),
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

func (f *MessageFactoryImpl) CreatePrepareMessage(blockRef BlockRef, sender SenderSignature) PrepareMessage {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateCommitMessage(blockRef BlockRef, sender SenderSignature) CommitMessage {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateViewChangeMessage(vcHeader ViewChangeHeader, sender SenderSignature, block Block) ViewChangeMessage {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateNewViewMessage(preprepareMessage PreprepareMessage, nvHeader NewViewHeader, sender SenderSignature) NewViewMessage {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateSenderSignature(sender []byte, signature []byte) SenderSignature {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateBlockRef(messageType int, blockHeight int, view int, blockHash []byte) BlockRef {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateNewViewHeader(messageType int, blockHeight int, view int, confirmations []ViewChangeConfirmation) NewViewHeader {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateViewChangeConfirmation(vcHeader ViewChangeHeader, sender SenderSignature) ViewChangeConfirmation {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreateViewChangeHeader(blockHeight int, view int, proof PreparedProof) ViewChangeHeader {
	panic("implement me")
}

func (f *MessageFactoryImpl) CreatePreparedProof(ppBlockRef BlockRef, pBlockRef BlockRef, ppSender SenderSignature, pSenders []SenderSignature) PreparedProof {
	panic("implement me")
}
