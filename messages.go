package leanhelix

import "github.com/orbs-network/lean-helix-go/types"

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
	Block *types.Block
}

type PrepareMessage struct {
	*BlockRefMessage
}

type CommitMessage struct {
	*BlockRefMessage
}

type ViewChangeMessage struct {
	SignaturePair *SignaturePair
	Block         *types.Block
	Content       *ViewChangeMessageContent
}

type NewViewMessage struct {
	SignaturePair     *SignaturePair
	PreprepareMessage *PrePrepareMessage
	Content           *NewViewContent
}

type NewViewContent struct {
	MessageType MessageType
	Term        types.BlockHeight
	View        types.ViewCounter
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
	SignerPublicKey  types.PublicKey
	ContentSignature string
}

type BlockMessageContent struct {
	MessageType MessageType
	Term        types.BlockHeight
	View        types.ViewCounter
	BlockHash   types.BlockHash
}

type ViewChangeMessageContent struct {
	MessageType   MessageType
	Term          types.BlockHeight
	View          types.ViewCounter
	PreparedProof *PreparedProof
}

type PreparedProof struct {
	PreprepareBlockRefMessage *PrePrepareMessage
	PrepareBlockRefMessages   []*PrepareMessage
}
