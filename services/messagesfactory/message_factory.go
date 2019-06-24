// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package messagesfactory

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
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
	blockHash primitives.BlockHash) *protocol.PreprepareContentBuilder {

	signedHeader := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_PREPREPARE,
		InstanceId:  f.instanceId,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	dataToSign := signedHeader.Build().Raw()
	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(f.keyManager.SignConsensusMessage(context.Background(), blockHeight, dataToSign)),
	}

	return &protocol.PreprepareContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}
}

func (f *MessageFactory) CreatePreprepareMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block interfaces.Block,
	blockHash primitives.BlockHash) *interfaces.PreprepareMessage {

	content := f.CreatePreprepareMessageContentBuilder(blockHeight, view, block, blockHash)

	return interfaces.NewPreprepareMessage(content.Build(), block)
}

func (f *MessageFactory) CreatePreprepareMessageFromContentBuilder(ppmc *protocol.PreprepareContentBuilder, block interfaces.Block) *interfaces.PreprepareMessage {
	return interfaces.NewPreprepareMessage(ppmc.Build(), block)
}

func (f *MessageFactory) CreatePrepareMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	blockHash primitives.BlockHash) *interfaces.PrepareMessage {

	signedHeader := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_PREPARE,
		InstanceId:  f.instanceId,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(f.keyManager.SignConsensusMessage(context.Background(), blockHeight, signedHeader.Build().Raw())),
	}

	contentBuilder := protocol.PrepareContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}

	return interfaces.NewPrepareMessage(contentBuilder.Build())
}

func (f *MessageFactory) CreateCommitMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	blockHash primitives.BlockHash) *interfaces.CommitMessage {

	signedHeader := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_COMMIT,
		InstanceId:  f.instanceId,
		BlockHeight: blockHeight,
		View:        view,
		BlockHash:   blockHash,
	}

	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(f.keyManager.SignConsensusMessage(context.Background(), blockHeight, signedHeader.Build().Raw())),
	}

	randomSeedBytes := randomseed.RandomSeedToBytes(f.randomSeed)
	share := f.keyManager.SignRandomSeed(context.Background(), blockHeight, randomSeedBytes)
	contentBuilder := protocol.CommitContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
		Share:        share,
	}

	return interfaces.NewCommitMessage(contentBuilder.Build())
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
	preparedMessages *preparedmessages.PreparedMessages) *protocol.ViewChangeMessageContentBuilder {

	preparedProofBuilder := CreatePreparedProofBuilderFromPreparedMessages(preparedMessages)
	signedHeader := &protocol.ViewChangeHeaderBuilder{
		MessageType:   protocol.LEAN_HELIX_VIEW_CHANGE,
		InstanceId:    f.instanceId,
		BlockHeight:   blockHeight,
		View:          view,
		PreparedProof: preparedProofBuilder,
	}

	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(f.keyManager.SignConsensusMessage(context.Background(), blockHeight, signedHeader.Build().Raw())),
	}

	return &protocol.ViewChangeMessageContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
	}
}

func (f *MessageFactory) CreateViewChangeMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	preparedMessages *preparedmessages.PreparedMessages) *interfaces.ViewChangeMessage {

	var block interfaces.Block
	if preparedMessages != nil && preparedMessages.PreprepareMessage != nil {
		block = preparedMessages.PreprepareMessage.Block()
	}

	contentBuilder := f.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages)

	return interfaces.NewViewChangeMessage(contentBuilder.Build(), block)
}

func (f *MessageFactory) CreateNewViewMessageContentBuilder(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	ppContentBuilder *protocol.PreprepareContentBuilder,
	confirmations []*protocol.ViewChangeMessageContentBuilder) *protocol.NewViewMessageContentBuilder {

	signedHeader := &protocol.NewViewHeaderBuilder{
		MessageType:             protocol.LEAN_HELIX_NEW_VIEW,
		InstanceId:              f.instanceId,
		BlockHeight:             blockHeight,
		View:                    view,
		ViewChangeConfirmations: confirmations,
	}

	sender := &protocol.SenderSignatureBuilder{
		MemberId:  f.memberId,
		Signature: primitives.Signature(f.keyManager.SignConsensusMessage(context.Background(), blockHeight, signedHeader.Build().Raw())),
	}

	return &protocol.NewViewMessageContentBuilder{
		SignedHeader: signedHeader,
		Sender:       sender,
		Message:      ppContentBuilder,
	}
}

func (f *MessageFactory) CreateNewViewMessage(
	blockHeight primitives.BlockHeight,
	view primitives.View,
	ppContentBuilder *protocol.PreprepareContentBuilder,
	confirmations []*protocol.ViewChangeMessageContentBuilder,
	block interfaces.Block) *interfaces.NewViewMessage {

	contentBuilder := f.CreateNewViewMessageContentBuilder(blockHeight, view, ppContentBuilder, confirmations)
	return interfaces.NewNewViewMessage(contentBuilder.Build(), block)
}

func NewMessageFactory(instanceId primitives.InstanceId, keyManager interfaces.KeyManager, memberId primitives.MemberId, randomSeed uint64) *MessageFactory {
	return &MessageFactory{
		instanceId: instanceId,
		keyManager: keyManager,
		memberId:   memberId,
		randomSeed: randomSeed,
	}
}
