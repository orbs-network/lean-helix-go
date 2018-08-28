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
			Content:       blockMessageContent,
			SignaturePair: signaturePair,
		},
		Block: block,
	}

	return result
}

func CreatePrepareMessage(km lh.KeyManager, term lh.BlockHeight, view lh.ViewCounter, block *lh.Block) *lh.PrepareMessage {
	blockHash := blockUtils.CalculateBlockHash(block)

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
		Content:       blockMessageContent,
		SignaturePair: signaturePair,
	}

	return result
}

func CreateCommitMessage(km lh.KeyManager, term lh.BlockHeight, view lh.ViewCounter, block *lh.Block) *lh.CommitMessage {
	blockHash := blockUtils.CalculateBlockHash(block)

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
		Content:       blockMessageContent,
		SignaturePair: signaturePair,
	}

	return result
}
