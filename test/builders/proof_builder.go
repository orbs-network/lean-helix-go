package builders

import (
	. "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type MessageSigner struct {
	KeyManager KeyManager
	MemberId   primitives.MemberId
}

func CreatePreparedProof(
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
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders = make([]*protocol.SenderSignatureBuilder, len(pSigners))
		for i, mgr := range pSigners {
			pSenders[i] = &protocol.SenderSignatureBuilder{
				MemberId:  mgr.MemberId,
				Signature: mgr.KeyManager.Sign(pBlockRef.Build().Raw()),
			}
		}
	}
	if ppSigner == nil {
		ppBlockRef = nil
		ppSender = nil
	} else {
		ppBlockRef = &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		ppSender = &protocol.SenderSignatureBuilder{
			MemberId:  ppSigner.MemberId,
			Signature: ppSigner.KeyManager.Sign(ppBlockRef.Build().Raw()),
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

func APreparedProofByMessages(PPMessage *PreprepareMessage, PMessages []*PrepareMessage) *protocol.PreparedProof {
	preparedMessages := &PreparedMessages{
		PreprepareMessage: PPMessage,
		PrepareMessages:   PMessages,
	}

	return CreatePreparedProofBuilderFromPreparedMessages(preparedMessages).Build()
}
