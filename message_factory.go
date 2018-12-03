package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type MessageFactory struct {
	KeyManager
}

func (f *MessageFactory) CreatePreprepareMessageContentBuilder(
	blockHeight BlockHeight,
	view View,
	block Block,
	blockHash Uint256) *PreprepareContentBuilder {

	signedHeader := &BlockRefBuilder{
		MessageType: LEAN_HELIX_PREPREPARE,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	dataToSign := signedHeader.Build().Raw()
	sender := &SenderSignatureBuilder{
		SenderPublicKey: Ed25519PublicKey(f.KeyManager.MyPublicKey()),
		Signature:       Ed25519Sig(f.KeyManager.Sign(dataToSign)),
	}

	return &PreprepareContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}
}

func (f *MessageFactory) CreatePreprepareMessage(
	blockHeight BlockHeight,
	view View,
	block Block,
	blockHash Uint256) *PreprepareMessage {

	content := f.CreatePreprepareMessageContentBuilder(blockHeight, view, block, blockHash)

	return NewPreprepareMessage(content.Build(), block)
}

func (f *MessageFactory) CreatePreprepareMessageFromContentBuilder(ppmc *PreprepareContentBuilder, block Block) *PreprepareMessage {
	return NewPreprepareMessage(ppmc.Build(), block)
}

func (f *MessageFactory) CreatePrepareMessage(
	blockHeight BlockHeight,
	view View,
	blockHash Uint256) *PrepareMessage {

	signedHeader := &BlockRefBuilder{
		MessageType: LEAN_HELIX_PREPARE,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	sender := &SenderSignatureBuilder{
		SenderPublicKey: Ed25519PublicKey(f.KeyManager.MyPublicKey()),
		Signature:       Ed25519Sig(f.KeyManager.Sign(signedHeader.Build().Raw())),
	}

	contentBuilder := PrepareContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}

	return NewPrepareMessage(contentBuilder.Build())
}

func (f *MessageFactory) CreateCommitMessage(
	blockHeight BlockHeight,
	view View,
	blockHash Uint256) *CommitMessage {

	signedHeader := &BlockRefBuilder{
		MessageType: LEAN_HELIX_COMMIT,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	sender := &SenderSignatureBuilder{
		SenderPublicKey: Ed25519PublicKey(f.KeyManager.MyPublicKey()),
		Signature:       Ed25519Sig(f.KeyManager.Sign(signedHeader.Build().Raw())),
	}

	contentBuilder := CommitContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}

	return NewCommitMessage(contentBuilder.Build())
}

func CreatePreparedProofBuilderFromPreparedMessages(preparedMessages *PreparedMessages) *PreparedProofBuilder {
	if preparedMessages == nil {
		return nil
	}

	preprepareMessage := preparedMessages.PreprepareMessage
	prepareMessages := preparedMessages.PrepareMessages

	var ppBlockRef, pBlockRef *BlockRefBuilder
	var ppSender *SenderSignatureBuilder
	var pSenders []*SenderSignatureBuilder

	if preprepareMessage == nil {
		ppBlockRef = nil
		ppSender = nil
	} else {
		ppBlockRef = &BlockRefBuilder{
			MessageType: LEAN_HELIX_PREPREPARE,
			BlockHeight: preprepareMessage.BlockHeight(),
			View:        preprepareMessage.View(),
			BlockHash:   preprepareMessage.Content().SignedHeader().BlockHash(),
		}
		ppSender = &SenderSignatureBuilder{
			SenderPublicKey: preprepareMessage.Content().Sender().SenderPublicKey(),
			Signature:       preprepareMessage.Content().Sender().Signature(),
		}
	}

	if prepareMessages == nil {
		pBlockRef = nil
		pSenders = nil
	} else {
		pBlockRef = &BlockRefBuilder{
			MessageType: LEAN_HELIX_PREPARE,
			BlockHeight: prepareMessages[0].BlockHeight(),
			View:        prepareMessages[0].View(),
			BlockHash:   prepareMessages[0].Content().SignedHeader().BlockHash(),
		}
		pSenders = make([]*SenderSignatureBuilder, 0, len(prepareMessages))
		for _, pm := range prepareMessages {
			pSenders = append(pSenders, &SenderSignatureBuilder{
				SenderPublicKey: pm.Content().Sender().SenderPublicKey(),
				Signature:       pm.Content().Sender().Signature(),
			})
		}
	}

	return &PreparedProofBuilder{
		PreprepareBlockRef: ppBlockRef,
		PreprepareSender:   ppSender,
		PrepareBlockRef:    pBlockRef,
		PrepareSenders:     pSenders,
	}
}

func (f *MessageFactory) CreateViewChangeMessageContentBuilder(
	blockHeight BlockHeight,
	view View,
	preparedMessages *PreparedMessages) *ViewChangeMessageContentBuilder {

	preparedProofBuilder := CreatePreparedProofBuilderFromPreparedMessages(preparedMessages)
	signedHeader := &ViewChangeHeaderBuilder{
		MessageType:   LEAN_HELIX_VIEW_CHANGE,
		BlockHeight:   blockHeight,
		View:          view,
		PreparedProof: preparedProofBuilder,
	}

	sender := &SenderSignatureBuilder{
		SenderPublicKey: Ed25519PublicKey(f.KeyManager.MyPublicKey()),
		Signature:       Ed25519Sig(f.KeyManager.Sign(signedHeader.Build().Raw())),
	}

	return &ViewChangeMessageContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}
}

func (f *MessageFactory) CreateViewChangeMessage(
	blockHeight BlockHeight,
	view View,
	preparedMessages *PreparedMessages) *ViewChangeMessage {

	var block Block
	if preparedMessages != nil && preparedMessages.PreprepareMessage != nil {
		block = preparedMessages.PreprepareMessage.Block()
	}

	contentBuilder := f.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages)

	return NewViewChangeMessage(contentBuilder.Build(), block)
}

func (f *MessageFactory) CreateNewViewMessageContentBuilder(
	blockHeight BlockHeight,
	view View,
	ppContentBuilder *PreprepareContentBuilder,
	confirmations []*ViewChangeMessageContentBuilder) *NewViewMessageContentBuilder {

	signedHeader := &NewViewHeaderBuilder{
		MessageType: LEAN_HELIX_NEW_VIEW,
		BlockHeight: blockHeight,
		View:        view,
		ViewChangeConfirmations: confirmations,
	}

	sender := &SenderSignatureBuilder{
		SenderPublicKey: Ed25519PublicKey(f.KeyManager.MyPublicKey()),
		Signature:       Ed25519Sig(f.KeyManager.Sign(signedHeader.Build().Raw())),
	}

	return &NewViewMessageContentBuilder{
		SignedHeader: signedHeader,
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

	contentBuilder := f.CreateNewViewMessageContentBuilder(blockHeight, view, ppContentBuilder, confirmations)
	return NewNewViewMessage(contentBuilder.Build(), block)
}

func NewMessageFactory(keyManager KeyManager) *MessageFactory {
	return &MessageFactory{
		KeyManager: keyManager,
	}
}
