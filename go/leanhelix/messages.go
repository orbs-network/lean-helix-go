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
	*BlockMessageContent
	SignaturePair *SignaturePair
}

type PrePrepareMessage struct {
	*BlockRefMessage
	Block *Block
}

type PrepareMessage BlockRefMessage

type PreparedMessages struct {
	PreprepareMessage *PrePrepareMessage
	PrepareMessages   []*PrepareMessage
}

type CommitMessage BlockRefMessage

// TODO For now I want "Block" to be explicit - it's not as integral part
// of ViewChangeMessage as the other fields are

type ViewChangeMessage struct {
	*ViewChangeMessageContent
	*SignaturePair
	Block *Block
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
	MessageType   MessageType
	Term          BlockHeight
	View          ViewCounter
	PreparedProof *PreparedProof
}

type PreparedProof struct {
	PreprepareBlockRefMessage *BlockRefMessage
	PrepareBlockRefMessages   []*BlockRefMessage
}

type NewViewMessageContent struct {
	MessageType MessageType
	Term        BlockHeight
	View        ViewCounter
	Votes       []ViewChangeVote
}

type ViewChangeVote struct {
	Content       *ViewChangeMessageContent
	SignaturePair *SignaturePair
}
