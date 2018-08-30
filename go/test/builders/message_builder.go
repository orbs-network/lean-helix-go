package builders

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
)

func CreatePrePrepareMessage(km lh.KeyManager, term lh.BlockHeight, view lh.ViewCounter, block *lh.Block) *lh.PrePrepareMessage {
	blockHash := CalculateBlockHash(block)

	blockMessageContent := &lh.BlockMessageContent{
		MessageType: lh.MESSAGE_TYPE_PREPREPARE,
		Term:        term,
		View:        view,
		BlockHash:   blockHash,
	}

	signaturePair := &lh.SignaturePair{
		SignerPublicKey:  km.MyPublicKey(),
		ContentSignature: km.SignBlockMessageContent(blockMessageContent),
	}

	result := &lh.PrePrepareMessage{
		BlockRefMessage: &lh.BlockRefMessage{
			BlockMessageContent: blockMessageContent,
			SignaturePair:       signaturePair,
		},
		Block: block,
	}

	return result
}

func CreatePrepareMessage(km lh.KeyManager, term lh.BlockHeight, view lh.ViewCounter, block *lh.Block) *lh.PrepareMessage {
	blockHash := CalculateBlockHash(block)

	blockMessageContent := &lh.BlockMessageContent{
		MessageType: lh.MESSAGE_TYPE_PREPARE,
		Term:        term,
		View:        view,
		BlockHash:   blockHash,
	}

	signaturePair := &lh.SignaturePair{
		SignerPublicKey:  km.MyPublicKey(),
		ContentSignature: km.SignBlockMessageContent(blockMessageContent),
	}

	result := &lh.PrepareMessage{
		BlockMessageContent: blockMessageContent,
		SignaturePair:       signaturePair,
	}

	return result
}

func CreateCommitMessage(km lh.KeyManager, term lh.BlockHeight, view lh.ViewCounter, block *lh.Block) *lh.CommitMessage {
	blockHash := CalculateBlockHash(block)

	blockMessageContent := &lh.BlockMessageContent{
		MessageType: lh.MESSAGE_TYPE_COMMIT,
		Term:        term,
		View:        view,
		BlockHash:   blockHash,
	}

	signaturePair := &lh.SignaturePair{
		SignerPublicKey:  km.MyPublicKey(),
		ContentSignature: km.SignBlockMessageContent(blockMessageContent),
	}

	result := &lh.CommitMessage{
		BlockMessageContent: blockMessageContent,
		SignaturePair:       signaturePair,
	}

	return result
}

// TODO km should be ptr

func CreateViewChangeMessage(km lh.KeyManager, term lh.BlockHeight, view lh.ViewCounter, prepared *lh.PreparedMessages) *lh.ViewChangeMessage {
	var (
		preparedProof *lh.PreparedProof
		block         *lh.Block
	)
	if prepared != nil {
		preparedProof = generatePreparedProof(prepared)
		block = prepared.PreprepareMessage.Block
	}

	content := &lh.ViewChangeMessageContent{
		MessageType:   lh.MESSAGE_TYPE_VIEW_CHANGE,
		Term:          term,
		View:          view,
		PreparedProof: preparedProof,
	}

	signaturePair := &lh.SignaturePair{
		SignerPublicKey:  km.MyPublicKey(),
		ContentSignature: km.SignViewChangeMessage(content),
	}

	result := &lh.ViewChangeMessage{
		ViewChangeMessageContent: content,
		SignaturePair:            signaturePair,
		Block:                    block,
	}

	return result
}

func generatePreparedProof(prepared *lh.PreparedMessages) *lh.PreparedProof {

	blockRefMessageFromPrePrepare := &lh.BlockRefMessage{
		BlockMessageContent: prepared.PreprepareMessage.BlockMessageContent,
		SignaturePair:       prepared.PreprepareMessage.SignaturePair,
	}

	blockRefMessageFromPrepares := make([]*lh.BlockRefMessage, len(prepared.PrepareMessages))
	for _, msg := range prepared.PrepareMessages {
		blockRefMessageFromPrepare := &lh.BlockRefMessage{
			BlockMessageContent: msg.BlockMessageContent,
			SignaturePair:       msg.SignaturePair,
		}
		blockRefMessageFromPrepares = append(blockRefMessageFromPrepares, blockRefMessageFromPrepare)
	}

	return &lh.PreparedProof{
		PreprepareBlockRefMessage: blockRefMessageFromPrePrepare,
		PrepareBlockRefMessages:   blockRefMessageFromPrepares,
	}
}
