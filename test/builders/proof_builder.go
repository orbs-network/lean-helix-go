package builders

import (
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
	networkId primitives.NetworkId,
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
			NetworkId:   networkId,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders = make([]*protocol.SenderSignatureBuilder, len(pSigners))
		for i, mgr := range pSigners {
			pSenders[i] = &protocol.SenderSignatureBuilder{
				MemberId:  mgr.MemberId,
				Signature: mgr.KeyManager.SignConsensusMessage(pBlockRef.BlockHeight, pBlockRef.Build().Raw()),
			}
		}
	}
	if ppSigner == nil {
		ppBlockRef = nil
		ppSender = nil
	} else {
		ppBlockRef = &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPREPARE,
			NetworkId:   networkId,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		ppSender = &protocol.SenderSignatureBuilder{
			MemberId:  ppSigner.MemberId,
			Signature: ppSigner.KeyManager.SignConsensusMessage(ppBlockRef.BlockHeight, ppBlockRef.Build().Raw()),
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
