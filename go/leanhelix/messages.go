package leanhelix

type MessageType string

const (
	MESSAGE_TYPE_PREPREPARE  MessageType = "preprepare"
	MESSAGE_TYPE_PREPARE     MessageType = "prepare"
	MESSAGE_TYPE_COMMIT      MessageType = "commit"
	MESSAGE_TYPE_VIEW_CHANGE MessageType = "view-change"
	MESSAGE_TYPE_NEW_VIEW    MessageType = "new-view"
)

type BlockRefMessage struct {
	Content       *BlockMessageContent
	SignaturePair *SignaturePair
}

type PrePrepareMessage struct {
	*BlockRefMessage
	Block *Block
}

type PrepareMessage BlockRefMessage

type CommitMessage BlockRefMessage

type ViewChangeMessage struct {
	Content       *ViewChangeMessageContent
	SignaturePair *SignaturePair
	Block         *Block
}

type NewViewMessage struct {
	Content           *NewViewMessageContent
	SignaturePair     *SignaturePair
	PrePrepareMessage PrePrepareMessage
}

type MessageContent struct {
	MessageType MessageType
	Term        BlockHeight
	View        ViewCounter
}

type BlockMessageContent struct {
	MessageType MessageType
	Term        BlockHeight
	View        ViewCounter
	BlockHash   BlockHash
}

type SignaturePair struct {
	SignerPublicKey  PublicKey
	ContentSignature string
}

type ViewChangeMessageContent struct {
	MessageType   *MessageType
	Term          BlockHeight
	View          ViewCounter
	PreparedProof *PreparedProof
}

type PreparedProof struct {
	preprepareBlockRefMessage *BlockRefMessage
	prepareBlockRefMessages   []BlockRefMessage
}

type NewViewMessageContent struct {
	MessageType *MessageType
	Term        BlockHeight
	View        ViewCounter
	Votes       []ViewChangeVote
}

type ViewChangeVote struct {
	Content       *ViewChangeMessageContent
	SignaturePair *SignaturePair
}
