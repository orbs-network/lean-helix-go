package leanhelix

import "github.com/orbs-network/lean-helix-go/types"

type MessageFactory interface {
	CreatePreprepareMessage(term types.BlockHeight, view types.ViewCounter, block *types.Block)
}

type messageFactory struct {
	CalculateBlockHash func(block *types.Block) types.BlockHash
	keyManager         KeyManager
	MyPK               types.PublicKey
}

func NewMessageFactory(calculateBlockHash func(block *types.Block) types.BlockHash, keyManager KeyManager) *messageFactory {
	return &messageFactory{
		CalculateBlockHash: calculateBlockHash,
		keyManager:         keyManager,
		MyPK:               keyManager.MyPublicKey(),
	}
}

func (mf *messageFactory) CreatePreprepareMessage(term types.BlockHeight, view types.ViewCounter, block *types.Block) *PrePrepareMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockMessageContent := &BlockMessageContent{
		MessageType: MESSAGE_TYPE_PREPREPARE,
		Term:        term,
		View:        view,
		BlockHash:   blockHash,
	}

	signaturePair := &SignaturePair{
		SignerPublicKey:  mf.MyPK,
		ContentSignature: mf.keyManager.SignBlockMessageContent(blockMessageContent),
	}

	result := &PrePrepareMessage{
		BlockRefMessage: &BlockRefMessage{
			Content:       blockMessageContent,
			SignaturePair: signaturePair,
		},
		Block: block,
	}

	return result
}

func (mf *messageFactory) CreatePrepareMessage(term types.BlockHeight, view types.ViewCounter, block *types.Block) *PrepareMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockMessageContent := &BlockMessageContent{
		MessageType: MESSAGE_TYPE_PREPARE,
		Term:        term,
		View:        view,
		BlockHash:   blockHash,
	}

	signaturePair := &SignaturePair{
		SignerPublicKey:  mf.MyPK,
		ContentSignature: mf.keyManager.SignBlockMessageContent(blockMessageContent),
	}

	result := &PrepareMessage{
		BlockRefMessage: &BlockRefMessage{
			Content:       blockMessageContent,
			SignaturePair: signaturePair,
		},
	}

	return result
}

func (mf *messageFactory) CreateCommitMessage(term types.BlockHeight, view types.ViewCounter, block *types.Block) *CommitMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockMessageContent := &BlockMessageContent{
		MessageType: MESSAGE_TYPE_COMMIT,
		Term:        term,
		View:        view,
		BlockHash:   blockHash,
	}

	signaturePair := &SignaturePair{
		SignerPublicKey:  mf.MyPK,
		ContentSignature: mf.keyManager.SignBlockMessageContent(blockMessageContent),
	}

	result := &CommitMessage{
		BlockRefMessage: &BlockRefMessage{
			Content:       blockMessageContent,
			SignaturePair: signaturePair,
		},
	}

	return result
}

func (mf *messageFactory) CreateViewChangeMessage(term types.BlockHeight, view types.ViewCounter, prepared *PreparedMessages) *ViewChangeMessage {
	var (
		preparedProof *PreparedProof
		block         *types.Block
	)
	if prepared != nil {
		preparedProof = generatePreparedProof(prepared)
		block = prepared.PreprepareMessage.Block
	}

	content := &ViewChangeMessageContent{
		MessageType:   MESSAGE_TYPE_VIEW_CHANGE,
		Term:          term,
		View:          view,
		PreparedProof: preparedProof,
	}

	signaturePair := &SignaturePair{
		SignerPublicKey:  mf.MyPK,
		ContentSignature: mf.keyManager.SignViewChangeMessage(content),
	}
	result := &ViewChangeMessage{
		SignaturePair: signaturePair,
		Block:         block,
		Content:       content,
	}

	return result
}

func generatePreparedProof(prepared *PreparedMessages) *PreparedProof {

	blockRefMessageFromPrePrepare := &PrePrepareMessage{
		BlockRefMessage: &BlockRefMessage{
			Content:       prepared.PreprepareMessage.Content,
			SignaturePair: prepared.PreprepareMessage.SignaturePair,
		},
	}

	blockRefMessageFromPrepares := make([]*PrepareMessage, len(prepared.PrepareMessages))
	for _, msg := range prepared.PrepareMessages {

		blockRefMessageFromPrepare := &PrepareMessage{
			BlockRefMessage: &BlockRefMessage{
				SignaturePair: msg.SignaturePair,
				Content:       msg.Content,
			},
		}
		blockRefMessageFromPrepares = append(blockRefMessageFromPrepares, blockRefMessageFromPrepare)
	}

	return &PreparedProof{
		PreprepareBlockRefMessage: blockRefMessageFromPrePrepare,
		PrepareBlockRefMessages:   blockRefMessageFromPrepares,
	}
}
