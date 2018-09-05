package leanhelix

type MessageFactory interface {
	CreatePreprepareMessage(term BlockHeight, view ViewCounter, block *Block)
}

type messageFactory struct {
	CalculateBlockHash func(block *Block) BlockHash
	keyManager         KeyManager
	MyPK               PublicKey
}

func NewMessageFactory(calculateBlockHash func(block *Block) BlockHash, keyManager KeyManager) *messageFactory {
	return &messageFactory{
		CalculateBlockHash: calculateBlockHash,
		keyManager:         keyManager,
		MyPK:               keyManager.MyPublicKey(),
	}
}

func (mf *messageFactory) CreatePreprepareMessage(term BlockHeight, view ViewCounter, block *Block) *PrePrepareMessage {
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

func (mf *messageFactory) CreatePrepareMessage(term BlockHeight, view ViewCounter, block *Block) *PrepareMessage {
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

func (mf *messageFactory) CreateCommitMessage(term BlockHeight, view ViewCounter, block *Block) *CommitMessage {
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

func (mf *messageFactory) CreateViewChangeMessage(term BlockHeight, view ViewCounter, prepared *PreparedMessages) *ViewChangeMessage {
	var (
		preparedProof *PreparedProof
		block         *Block
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
