package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

// This is the ORBS side

type MessageFactory struct {
	KeyManager
}

func (f *MessageFactory) CreatePreprepareMessageContentBuilder(
	blockHeight BlockHeight,
	view View,
	block Block) *PreprepareContentBuilder {

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
	ppContentBuilder := &PreprepareContentBuilder{
		SignedHeader: header,
		Sender:       sender,
	}
	return ppContentBuilder
}

func (f *MessageFactory) CreatePreprepareMessage(
	blockHeight BlockHeight,
	view View,
	block Block) *PreprepareMessage {

	ppmc := f.CreatePreprepareMessageContentBuilder(blockHeight, view, block)
	ppm := &PreprepareMessage{
		content: ppmc.Build(),
		block:   block,
	}
	return ppm
}

func (f *MessageFactory) CreatePrepareMessage(
	blockHeight BlockHeight,
	view View,
	blockHash Uint256) *PrepareMessage {

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
	pContentBuilder := PrepareContentBuilder{
		SignedHeader: header,
		Sender:       sender,
	}
	pm := &PrepareMessage{
		content: pContentBuilder.Build(),
	}
	return pm
}

func (f *MessageFactory) CreateCommitMessage(
	blockHeight BlockHeight,
	view View,
	blockHash Uint256) *CommitMessage {

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
	cContentBuilder := CommitContentBuilder{
		SignedHeader: header,
		Sender:       sender,
	}
	cm := &CommitMessage{
		content: cContentBuilder.Build(),
	}
	return cm
}

func (f *MessageFactory) CreateViewChangeMessageContentBuilder(
	blockHeight BlockHeight,
	view View,
	preparedMessages *PreparedMessages) *ViewChangeMessageContentBuilder {

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

func (f *MessageFactory) CreateViewChangeMessage(
	blockHeight BlockHeight,
	view View,
	preparedMessages *PreparedMessages) *ViewChangeMessage {

	var block Block
	if preparedMessages != nil && preparedMessages.PreprepareMessage != nil {
		block = preparedMessages.PreprepareMessage.Block()
	} else {
		block = nil
	}

	vcmcb := f.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages)
	vcm := &ViewChangeMessage{
		content: vcmcb.Build(),
		block:   block,
	}

	return vcm
}

func (f *MessageFactory) CreateNewViewMessageContentBuilder(
	blockHeight BlockHeight,
	view View,
	ppContentBuilder *PreprepareContentBuilder,
	confirmations []*ViewChangeMessageContentBuilder) *NewViewMessageContentBuilder {

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
		PreprepareMessageContent: ppContentBuilder,
	}
}

func (f *MessageFactory) CreateNewViewMessage(
	blockHeight BlockHeight,
	view View,
	ppContentBuilder *PreprepareContentBuilder,
	confirmations []*ViewChangeMessageContentBuilder,
	block Block) *NewViewMessage {

	nvmcb := f.CreateNewViewMessageContentBuilder(blockHeight, view, ppContentBuilder, confirmations).Build()
	return &NewViewMessage{
		content: nvmcb,
		block:   block,
	}
}

func NewMessageFactory(keyManager KeyManager) *MessageFactory {
	return &MessageFactory{
		KeyManager: keyManager,
	}
}
