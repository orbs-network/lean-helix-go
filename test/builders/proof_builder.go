package builders

import (
	. "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
)

func CreatePreparedProof(
	ppKeyManager KeyManager,
	pKeyManagers []KeyManager,
	height primitives.BlockHeight,
	view primitives.View,
	blockHash primitives.Uint256) *PreparedProof {

	var ppBlockRef *BlockRefBuilder
	var pBlockRef *BlockRefBuilder
	var ppSender *SenderSignatureBuilder
	var pSenders []*SenderSignatureBuilder

	if len(pKeyManagers) == 0 {
		pBlockRef = nil
		pSenders = nil
	} else {
		pBlockRef = &BlockRefBuilder{
			MessageType: LEAN_HELIX_PREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders = make([]*SenderSignatureBuilder, len(pKeyManagers))
		for i, mgr := range pKeyManagers {
			dataToSign := pBlockRef.Build().Raw()
			sig, err := mgr.Sign(dataToSign)
			if err != nil {
				return nil
			}
			pSenders[i] = &SenderSignatureBuilder{
				SenderPublicKey: mgr.MyPublicKey(),
				Signature:       sig,
			}
		}
	}
	if ppKeyManager == nil {
		ppBlockRef = nil
		ppSender = nil
	} else {
		ppBlockRef = &BlockRefBuilder{
			MessageType: LEAN_HELIX_PREPREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		dataToSign := ppBlockRef.Build().Raw()
		sig, err := ppKeyManager.Sign(dataToSign)
		if err != nil {
			return nil
		}
		ppSender = &SenderSignatureBuilder{
			SenderPublicKey: ppKeyManager.MyPublicKey(),
			Signature:       sig,
		}
	}
	preparedProof := &PreparedProofBuilder{
		PreprepareBlockRef: ppBlockRef,
		PreprepareSender:   ppSender,
		PrepareBlockRef:    pBlockRef,
		PrepareSenders:     pSenders,
	}

	return preparedProof.Build()
}

func APreparedProofByMessages(PPMessage *PreprepareMessage, PMessages []*PrepareMessage) *PreparedProof {
	preparedMessages := &PreparedMessages{
		PreprepareMessage: PPMessage,
		PrepareMessages:   PMessages,
	}

	return CreatePreparedProofBuilderFromPreparedMessages(preparedMessages).Build()
}
