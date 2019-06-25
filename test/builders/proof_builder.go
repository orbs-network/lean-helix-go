// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package builders

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type MessageSigner struct {
	KeyManager interfaces.KeyManager
	MemberId   primitives.MemberId
}

func CreatePreparedProof(
	instanceId primitives.InstanceId,
	ppSigner *MessageSigner,
	pSigners []*MessageSigner,
	height primitives.BlockHeight,
	view primitives.View,
	blockHash primitives.BlockHash) *protocol.PreparedProof {

	var ppBlockRef *protocol.BlockRefBuilder
	var pBlockRef *protocol.BlockRefBuilder
	var ppSender *protocol.SenderSignatureBuilder
	var pSenders []*protocol.SenderSignatureBuilder

	if len(pSigners) == 0 {
		pBlockRef = nil
		pSenders = nil
	} else {
		pBlockRef = &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPARE,
			InstanceId:  instanceId,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders = make([]*protocol.SenderSignatureBuilder, len(pSigners))
		for i, mgr := range pSigners {
			pSenders[i] = &protocol.SenderSignatureBuilder{
				MemberId:  mgr.MemberId,
				Signature: mgr.KeyManager.SignConsensusMessage(context.Background(), pBlockRef.BlockHeight, pBlockRef.Build().Raw()),
			}
		}
	}
	if ppSigner == nil {
		ppBlockRef = nil
		ppSender = nil
	} else {
		ppBlockRef = &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPREPARE,
			InstanceId:  instanceId,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		ppSender = &protocol.SenderSignatureBuilder{
			MemberId:  ppSigner.MemberId,
			Signature: ppSigner.KeyManager.SignConsensusMessage(context.Background(), ppBlockRef.BlockHeight, ppBlockRef.Build().Raw()),
		}
	}
	preparedProof := &protocol.PreparedProofBuilder{
		PreprepareBlockRef: ppBlockRef,
		PreprepareSender:   ppSender,
		PrepareBlockRef:    pBlockRef,
		PrepareSenders:     pSenders,
	}

	return preparedProof.Build()
}

func APreparedProofByMessages(PPMessage *interfaces.PreprepareMessage, PMessages []*interfaces.PrepareMessage) *protocol.PreparedProof {
	preparedMessages := &preparedmessages.PreparedMessages{
		PreprepareMessage: PPMessage,
		PrepareMessages:   PMessages,
	}

	return messagesfactory.CreatePreparedProofBuilderFromPreparedMessages(preparedMessages).Build()
}
