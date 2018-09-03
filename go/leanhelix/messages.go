package leanhelix

type MessageType string

const (
	MESSAGE_TYPE_PREPREPARE  MessageType = "preprepare"
	MESSAGE_TYPE_PREPARE     MessageType = "prepare"
	MESSAGE_TYPE_COMMIT      MessageType = "commit"
	MESSAGE_TYPE_VIEW_CHANGE MessageType = "view-change"
	MESSAGE_TYPE_NEW_VIEW    MessageType = "new-view"
)

// TODO how to impl this???
type Message interface {
	SignaturePair() SignaturePair
	Content() MessageContent
}

type MessageContent interface {
	MessageType()
	Term()
	View()
}

type BlockMessageContent interface {
	BlockHash() BlockHash
}

type BlockRefMessage interface {
	Message
	BlockMessageContent
}

type PrePrepareMessage struct {
	Message
	Block *Block
}

type PrepareMessage struct {
	Message
	BlockHash BlockHash
}

type CommitMessage struct {
	Message
	BlockHash BlockHash
}

type PreparedMessages struct {
	PreprepareMessage *PrePrepareMessage
	PrepareMessages   []*PrepareMessage
}

type ViewChangeMessage struct {
	Message
	Block         *Block
	PreparedProof *PreparedProof
}

type NewViewMessage struct {
	Content           *NewViewMessageContent
	SignaturePair     *SignaturePair
	PrePrepareMessage PrePrepareMessage
}

/*
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
*/
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
	PrepareBlockRefMessages   []*PrepareMessage
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
