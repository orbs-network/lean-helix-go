package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

// This is the ORBS side

type MessageFactoryImpl struct {
	KeyManager
}

func (f *MessageFactoryImpl) CreatePreprepareMessageContentBuilder(blockHeight BlockHeight, view View, block Block) *PreprepareMessageContentBuilder {
	header := &BlockRefBuilder{
		MessageType: LEAN_HELIX_PREPREPARE,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   block.BlockHash(),
	}
	sig := Ed25519Sig(f.KeyManager.Sign(header.Build().Raw()))
	me := Ed25519PublicKey(f.KeyManager.MyPublicKey())
	sender := &SenderSignatureBuilder{
		SenderPublicKey: me,
		Signature:       sig,
	}
	ppmc := &PreprepareMessageContentBuilder{
		SignedHeader: header,
		Sender:       sender,
	}
	return ppmc
}

func (f *MessageFactoryImpl) CreatePreprepareMessage(blockHeight BlockHeight, view View, block Block) PreprepareMessage {
	ppmc := f.CreatePreprepareMessageContentBuilder(blockHeight, view, block)
	ppm := &PreprepareMessageImpl{
		Content: ppmc.Build(),
		MyBlock: block,
	}
	return ppm
}

func (f *MessageFactoryImpl) CreatePrepareMessage(blockHeight BlockHeight, view View, blockHash Uint256) PrepareMessage {
	header := &BlockRefBuilder{
		MessageType: LEAN_HELIX_PREPARE,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}
	sig := Ed25519Sig(f.KeyManager.Sign(header.Build().Raw()))
	me := Ed25519PublicKey(f.KeyManager.MyPublicKey())
	sender := &SenderSignatureBuilder{
		SenderPublicKey: me,
		Signature:       sig,
	}
	pmc := PrepareMessageContentBuilder{
		SignedHeader: header,
		Sender:       sender,
	}
	pm := &PrepareMessageImpl{
		Content: pmc.Build(),
	}
	return pm
}

func (f *MessageFactoryImpl) CreateCommitMessage(blockHeight BlockHeight, view View, blockHash Uint256) CommitMessage {
	header := &BlockRefBuilder{
		MessageType: LEAN_HELIX_COMMIT,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}
	sig := Ed25519Sig(f.KeyManager.Sign(header.Build().Raw()))
	me := Ed25519PublicKey(f.KeyManager.MyPublicKey())
	sender := &SenderSignatureBuilder{
		SenderPublicKey: me,
		Signature:       sig,
	}
	cmc := CommitMessageContentBuilder{
		SignedHeader: header,
		Sender:       sender,
	}
	pm := &CommitMessageImpl{
		Content: cmc.Build(),
	}
	return pm
}

func (f *MessageFactoryImpl) CreateViewChangeMessageContentBuilder(blockHeight BlockHeight, view View, preparedMessages *PreparedMessages) *ViewChangeMessageContentBuilder {
	preparedProofBuilder := CreatePreparedProofBuilderFromPreparedMessages(preparedMessages)

	header := &ViewChangeHeaderBuilder{
		MessageType:   LEAN_HELIX_VIEW_CHANGE,
		BlockHeight:   blockHeight,
		View:          view,
		PreparedProof: preparedProofBuilder,
	}
	sig := Ed25519Sig(f.KeyManager.Sign(header.Build().Raw()))
	me := Ed25519PublicKey(f.KeyManager.MyPublicKey())
	sender := &SenderSignatureBuilder{
		SenderPublicKey: me,
		Signature:       sig,
	}
	cvmcb := &ViewChangeMessageContentBuilder{
		SignedHeader: header,
		Sender:       sender,
	}
	return cvmcb

}

func (f *MessageFactoryImpl) CreateViewChangeMessage(blockHeight BlockHeight, view View, preparedMessages *PreparedMessages) ViewChangeMessage {

	var block Block
	if preparedMessages != nil && preparedMessages.PreprepareMessage != nil {
		block = preparedMessages.PreprepareMessage.Block()
	} else {
		block = nil
	}

	vcmcb := f.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages)
	vcm := &ViewChangeMessageImpl{
		Content: vcmcb.Build(),
		MyBlock: block,
	}

	return vcm
}

func (f *MessageFactoryImpl) CreateNewViewMessageContentBuilder(blockHeight BlockHeight, view View, ppmcb *PreprepareMessageContentBuilder, confirmations []*ViewChangeMessageContentBuilder) *NewViewMessageContentBuilder {

	header := &NewViewHeaderBuilder{
		MessageType: LEAN_HELIX_NEW_VIEW,
		BlockHeight: blockHeight,
		View:        view,
		ViewChangeConfirmations: confirmations,
	}

	sig := Ed25519Sig(f.KeyManager.Sign(header.Build().Raw()))
	me := Ed25519PublicKey(f.KeyManager.MyPublicKey())
	sender := &SenderSignatureBuilder{
		SenderPublicKey: me,
		Signature:       sig,
	}

	return &NewViewMessageContentBuilder{
		SignedHeader: header,
		Sender:       sender,
		PreprepareMessageContent: ppmcb,
	}
}

func (f *MessageFactoryImpl) CreateNewViewMessage(blockHeight BlockHeight, view View, ppmcb *PreprepareMessageContentBuilder, confirmations []*ViewChangeMessageContentBuilder, block Block) NewViewMessage {
	nvmcb := f.CreateNewViewMessageContentBuilder(blockHeight, view, ppmcb, confirmations).Build()
	return &NewViewMessageImpl{
		Content: nvmcb,
		MyBlock: block,
	}
}

func NewMessageFactory(keyManager KeyManager) MessageFactory {
	return &MessageFactoryImpl{
		KeyManager: keyManager,
	}
}
