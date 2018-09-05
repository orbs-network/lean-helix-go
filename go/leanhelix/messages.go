package leanhelix

type MessageType string

const (
	MESSAGE_TYPE_PREPREPARE  MessageType = "preprepare"
	MESSAGE_TYPE_PREPARE     MessageType = "prepare"
	MESSAGE_TYPE_COMMIT      MessageType = "commit"
	MESSAGE_TYPE_VIEW_CHANGE MessageType = "view-change"
	MESSAGE_TYPE_NEW_VIEW    MessageType = "new-view"
)

type Message interface {
	SignaturePair() SignaturePair
}

type BlockRefMessage struct {
	SignaturePair *SignaturePair
	Content       *BlockMessageContent
}

type PrePrepareMessage struct {
	*BlockRefMessage
	Block *Block
}

type PrepareMessage struct {
	*BlockRefMessage
}

type CommitMessage struct {
	*BlockRefMessage
}

type ViewChangeMessage struct {
	SignaturePair *SignaturePair
	Block         *Block
	Content       *ViewChangeMessageContent
}

type NewViewMessage struct {
	SignaturePair     *SignaturePair
	PreprepareMessage *PrePrepareMessage
	Content           *NewViewContent
}

type NewViewContent struct {
	MessageType MessageType
	Term        BlockHeight
	View        ViewCounter
	Votes       []*ViewChangeVote
}

type ViewChangeVote struct {
	SignaturePair *SignaturePair
	Content       *ViewChangeMessageContent
}

type PreparedMessages struct {
	PreprepareMessage *PrePrepareMessage
	PrepareMessages   []*PrepareMessage
}

type SignaturePair struct {
	SignerPublicKey  PublicKey
	ContentSignature string
}

type BlockMessageContent struct {
	MessageType MessageType
	Term        BlockHeight
	View        ViewCounter
	BlockHash   BlockHash
}

type ViewChangeMessageContent struct {
	MessageType   MessageType
	Term          BlockHeight
	View          ViewCounter
	PreparedProof *PreparedProof
}

type PreparedProof struct {
	PreprepareBlockRefMessage *PrePrepareMessage
	PrepareBlockRefMessages   []*PrepareMessage
}
