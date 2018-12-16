package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type MessageFactory struct {
	KeyManager
}

func (f *MessageFactory) CreatePreprepareMessageContentBuilder(
	blockHeight BlockHeight,
	view View,
	block Block,
	blockHash BlockHash) *protocol.PreprepareContentBuilder {

	signedHeader := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_PREPREPARE,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	dataToSign := signedHeader.Build().Raw()
	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.KeyManager.MyPublicKey(),
		Signature: Signature(f.KeyManager.Sign(dataToSign)),
	}

	return &protocol.PreprepareContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}
}

func (f *MessageFactory) CreatePreprepareMessage(
	blockHeight BlockHeight,
	view View,
	block Block,
	blockHash BlockHash) *PreprepareMessage {

	content := f.CreatePreprepareMessageContentBuilder(blockHeight, view, block, blockHash)

	return NewPreprepareMessage(content.Build(), block)
}

func (f *MessageFactory) CreatePreprepareMessageFromContentBuilder(ppmc *protocol.PreprepareContentBuilder, block Block) *PreprepareMessage {
	return NewPreprepareMessage(ppmc.Build(), block)
}

func (f *MessageFactory) CreatePrepareMessage(
	blockHeight BlockHeight,
	view View,
	blockHash BlockHash) *PrepareMessage {

	signedHeader := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_PREPARE,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	sender := &protocol.SenderSignatureBuilder{
		MemberId:  MemberId(f.KeyManager.MyPublicKey()),
		Signature: Signature(f.KeyManager.Sign(signedHeader.Build().Raw())),
	}

	contentBuilder := protocol.PrepareContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}

	return NewPrepareMessage(contentBuilder.Build())
}

func (f *MessageFactory) CreateCommitMessage(
	blockHeight BlockHeight,
	view View,
	blockHash BlockHash) *CommitMessage {

	signedHeader := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_COMMIT,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	sender := &protocol.SenderSignatureBuilder{
		MemberId:  MemberId(f.KeyManager.MyPublicKey()),
		Signature: Signature(f.KeyManager.Sign(signedHeader.Build().Raw())),
	}

	contentBuilder := protocol.CommitContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}

	return NewCommitMessage(contentBuilder.Build())
}

func CreatePreparedProofBuilderFromPreparedMessages(preparedMessages *PreparedMessages) *protocol.PreparedProofBuilder {
	if preparedMessages == nil {
		return nil
	}

	preprepareMessage := preparedMessages.PreprepareMessage
	prepareMessages := preparedMessages.PrepareMessages

	var ppBlockRef, pBlockRef *protocol.BlockRefBuilder
	var ppSender *protocol.SenderSignatureBuilder
	var pSenders []*protocol.SenderSignatureBuilder

	if preprepareMessage == nil {
		ppBlockRef = nil
		ppSender = nil
	} else {
		ppBlockRef = &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPREPARE,
			BlockHeight: preprepareMessage.BlockHeight(),
			View:        preprepareMessage.View(),
			BlockHash:   preprepareMessage.Content().SignedHeader().BlockHash(),
		}
		ppSender = &protocol.SenderSignatureBuilder{
			MemberId:  preprepareMessage.Content().Sender().MemberId(),
			Signature: preprepareMessage.Content().Sender().Signature(),
		}
	}

	if prepareMessages == nil {
		pBlockRef = nil
		pSenders = nil
	} else {
		pBlockRef = &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPARE,
			BlockHeight: prepareMessages[0].BlockHeight(),
			View:        prepareMessages[0].View(),
			BlockHash:   prepareMessages[0].Content().SignedHeader().BlockHash(),
		}
		pSenders = make([]*protocol.SenderSignatureBuilder, 0, len(prepareMessages))
		for _, pm := range prepareMessages {
			pSenders = append(pSenders, &protocol.SenderSignatureBuilder{
				MemberId:  pm.Content().Sender().MemberId(),
				Signature: pm.Content().Sender().Signature(),
			})
		}
	}

	return &protocol.PreparedProofBuilder{
		PreprepareBlockRef: ppBlockRef,
		PreprepareSender:   ppSender,
		PrepareBlockRef:    pBlockRef,
		PrepareSenders:     pSenders,
	}
}

func (f *MessageFactory) CreateViewChangeMessageContentBuilder(
	blockHeight BlockHeight,
	view View,
	preparedMessages *PreparedMessages) *protocol.ViewChangeMessageContentBuilder {

	preparedProofBuilder := CreatePreparedProofBuilderFromPreparedMessages(preparedMessages)
	signedHeader := &protocol.ViewChangeHeaderBuilder{
		MessageType:   protocol.LEAN_HELIX_VIEW_CHANGE,
		BlockHeight:   blockHeight,
		View:          view,
		PreparedProof: preparedProofBuilder,
	}

	sender := &protocol.SenderSignatureBuilder{
		MemberId:  MemberId(f.KeyManager.MyPublicKey()),
		Signature: Signature(f.KeyManager.Sign(signedHeader.Build().Raw())),
	}

	return &protocol.ViewChangeMessageContentBuilder{
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
	ppContentBuilder *protocol.PreprepareContentBuilder,
	confirmations []*protocol.ViewChangeMessageContentBuilder) *protocol.NewViewMessageContentBuilder {

	signedHeader := &protocol.NewViewHeaderBuilder{
		MessageType: protocol.LEAN_HELIX_NEW_VIEW,
		BlockHeight: blockHeight,
		View:        view,
		ViewChangeConfirmations: confirmations,
	}

	sender := &protocol.SenderSignatureBuilder{
		MemberId:  MemberId(f.KeyManager.MyPublicKey()),
		Signature: Signature(f.KeyManager.Sign(signedHeader.Build().Raw())),
	}

	return &protocol.NewViewMessageContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
		Message:      ppContentBuilder,
	}
}

func (f *MessageFactory) CreateNewViewMessage(
	blockHeight BlockHeight,
	view View,
	ppContentBuilder *protocol.PreprepareContentBuilder,
	confirmations []*protocol.ViewChangeMessageContentBuilder,
	block Block) *NewViewMessage {

	contentBuilder := f.CreateNewViewMessageContentBuilder(blockHeight, view, ppContentBuilder, confirmations)
	return NewNewViewMessage(contentBuilder.Build(), block)
}

func NewMessageFactory(keyManager KeyManager) *MessageFactory {
	return &MessageFactory{
		KeyManager: keyManager,
	}
}
