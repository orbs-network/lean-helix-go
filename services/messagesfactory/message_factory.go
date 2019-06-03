package messagesfactory

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/pkg/errors"
)

type MessageFactory struct {
	instanceId primitives.InstanceId
	keyManager interfaces.KeyManager
	memberId   primitives.MemberId
	randomSeed uint64
}

func (f *MessageFactory) CreatePreprepareMessageContentBuilder(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block interfaces.Block,
	blockHash primitives.BlockHash) (*protocol.PreprepareContentBuilder, error) {

	signedHeader := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_PREPREPARE,
		InstanceId:  f.instanceId,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	sig, err := f.keyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw())
	if err != nil {
		return nil, errors.Wrap(err, "could not create preprepare message")
	}
	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(sig),
	}

	return &protocol.PreprepareContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}, nil
}

func (f *MessageFactory) CreatePreprepareMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block interfaces.Block,
	blockHash primitives.BlockHash) (*interfaces.PreprepareMessage, error) {

	content, err := f.CreatePreprepareMessageContentBuilder(blockHeight, view, block, blockHash)
	if err != nil {
		return nil, err
	}

	return interfaces.NewPreprepareMessage(content.Build(), block), nil
}

func (f *MessageFactory) CreatePreprepareMessageFromContentBuilder(ppmc *protocol.PreprepareContentBuilder, block interfaces.Block) *interfaces.PreprepareMessage {
	return interfaces.NewPreprepareMessage(ppmc.Build(), block)
}

func (f *MessageFactory) CreatePrepareMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	blockHash primitives.BlockHash) (*interfaces.PrepareMessage, error) {

	signedHeader := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_PREPARE,
		InstanceId:  f.instanceId,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	sig, err := f.keyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw())
	if err != nil {
		return nil, errors.Wrap(err, "could not create prepare message")
	}
	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(sig),
	}

	contentBuilder := protocol.PrepareContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}

	return interfaces.NewPrepareMessage(contentBuilder.Build()), nil
}

func (f *MessageFactory) CreateCommitMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	blockHash primitives.BlockHash) (*interfaces.CommitMessage, error) {

	signedHeader := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_COMMIT,
		InstanceId:  f.instanceId,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	sig, err := f.keyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw())
	if err != nil {
		return nil, errors.Wrap(err, "could not create commit message")
	}

	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(sig),
	}

	randomSeedBytes := randomseed.RandomSeedToBytes(f.randomSeed)
	share := f.keyManager.SignRandomSeed(blockHeight, randomSeedBytes)
	contentBuilder := protocol.CommitContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
		Share:        share,
	}

	return interfaces.NewCommitMessage(contentBuilder.Build()), nil
}

func CreatePreparedProofBuilderFromPreparedMessages(preparedMessages *preparedmessages.PreparedMessages) *protocol.PreparedProofBuilder {
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
			InstanceId:  preprepareMessage.InstanceId(),
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
			InstanceId:  prepareMessages[0].InstanceId(),
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
	blockHeight primitives.BlockHeight,
	view primitives.View,
	preparedMessages *preparedmessages.PreparedMessages) (*protocol.ViewChangeMessageContentBuilder, error) {

	preparedProofBuilder := CreatePreparedProofBuilderFromPreparedMessages(preparedMessages)
	signedHeader := &protocol.ViewChangeHeaderBuilder{
		MessageType:   protocol.LEAN_HELIX_VIEW_CHANGE,
		InstanceId:    f.instanceId,
		BlockHeight:   blockHeight,
		View:          view,
		PreparedProof: preparedProofBuilder,
	}

	sig, err := f.keyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw())
	if err != nil {
		return nil, errors.Wrap(err, "could not create view change message")
	}
	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(sig),
	}

	return &protocol.ViewChangeMessageContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}, nil
}

func (f *MessageFactory) CreateViewChangeMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	preparedMessages *preparedmessages.PreparedMessages) (*interfaces.ViewChangeMessage, error) {

	var block interfaces.Block
	if preparedMessages != nil && preparedMessages.PreprepareMessage != nil {
		block = preparedMessages.PreprepareMessage.Block()
	}

	contentBuilder, err := f.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages)
	if err != nil {
		return nil, err
	}

	return interfaces.NewViewChangeMessage(contentBuilder.Build(), block), err
}

func (f *MessageFactory) CreateNewViewMessageContentBuilder(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	ppContentBuilder *protocol.PreprepareContentBuilder,
	confirmations []*protocol.ViewChangeMessageContentBuilder) (*protocol.NewViewMessageContentBuilder, error) {

	signedHeader := &protocol.NewViewHeaderBuilder{
		MessageType:             protocol.LEAN_HELIX_NEW_VIEW,
		InstanceId:              f.instanceId,
		BlockHeight:             blockHeight,
		View:                    view,
		ViewChangeConfirmations: confirmations,
	}

	sig, err := f.keyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw())
	if err != nil {
		return nil, errors.Wrap(err, "could not create new view")
	}
	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(sig),
	}

	return &protocol.NewViewMessageContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
		Message:      ppContentBuilder,
	}, nil
}

func (f *MessageFactory) CreateNewViewMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	ppContentBuilder *protocol.PreprepareContentBuilder,
	confirmations []*protocol.ViewChangeMessageContentBuilder,
	block interfaces.Block) (*interfaces.NewViewMessage, error) {

	contentBuilder, err := f.CreateNewViewMessageContentBuilder(blockHeight, view, ppContentBuilder, confirmations)
	if err != nil {
		return nil, err
	}
	return interfaces.NewNewViewMessage(contentBuilder.Build(), block), nil
}

func NewMessageFactory(instanceId primitives.InstanceId, keyManager interfaces.KeyManager, memberId primitives.MemberId, randomSeed uint64) *MessageFactory {
	return &MessageFactory{
		instanceId: instanceId,
		keyManager: keyManager,
		memberId:   memberId,
		randomSeed: randomSeed,
	}
}
