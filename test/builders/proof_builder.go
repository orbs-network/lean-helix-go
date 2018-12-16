package builders

import (
	. "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

func CreatePreparedProof(
	ppKeyManager KeyManager,
	pKeyManagers []KeyManager,
	height primitives.BlockHeight,
	view primitives.View,
	blockHash primitives.BlockHash) *protocol.PreparedProof {

	var ppBlockRef *protocol.BlockRefBuilder
	var pBlockRef *protocol.BlockRefBuilder
	var ppSender *protocol.SenderSignatureBuilder
	var pSenders []*protocol.SenderSignatureBuilder

	if len(pKeyManagers) == 0 {
		pBlockRef = nil
		pSenders = nil
	} else {
		pBlockRef = &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders = make([]*protocol.SenderSignatureBuilder, len(pKeyManagers))
		for i, mgr := range pKeyManagers {
			pSenders[i] = &protocol.SenderSignatureBuilder{
				MemberId:  mgr.MyPublicKey(),
				Signature: mgr.Sign(pBlockRef.Build().Raw()),
			}
		}
	}
	if ppKeyManager == nil {
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
			MemberId:  ppKeyManager.MyPublicKey(),
			Signature: ppKeyManager.Sign(ppBlockRef.Build().Raw()),
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
