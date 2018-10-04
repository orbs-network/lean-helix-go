package leanhelix

type MessageType string

const (
	MESSAGE_TYPE_PREPREPARE  MessageType = "preprepare"
	MESSAGE_TYPE_PREPARE     MessageType = "prepare"
	MESSAGE_TYPE_COMMIT      MessageType = "commit"
	MESSAGE_TYPE_VIEW_CHANGE MessageType = "view-change"
	MESSAGE_TYPE_NEW_VIEW    MessageType = "new-view"
)

// START MESSAGE TYPES //

type PreprepareMessage interface {
	SignedHeader() BlockRef
	Sender() SenderSignature
	Block() Block
}

type PrepareMessage interface {
	SignedHeader() BlockRef
	Sender() SenderSignature
}

type CommitMessage interface {
	SignedHeader() BlockRef
	Sender() SenderSignature
	// TODO Add RandomSeedShare?
}

type ViewChangeHeader interface {
	HasMessageType
	BlockHeight() BlockHeight
	View() View
	PreparedProof() PreparedProof
}

type ViewChangeMessage interface {
	SignedHeader() ViewChangeHeader
	Sender() SenderSignature
	Block() Block
}

type NewViewMessage interface {
	SignedHeader() NewViewHeader
	Sender() SenderSignature
}

// END MESSAGE TYPES //

type SenderSignature interface {
	SenderPublicKey() PublicKey
	Signature() Signature
}

type HasMessageType interface {
	MessageType() MessageType
}

type MessageTransporter interface {
	SenderSignature
	HasMessageType
}

type BlockRef interface {
	HasMessageType
	BlockHeight() BlockHeight
	View() View
	BlockHash() BlockHash // TODO Gad: rename this to "current block hash"???
}

// TODO this is different from definition of LeanHelixPreparedProof in lean_helix.mb.go:448 in orbs-spec
type PreparedProof interface {
	PreprepareMessage() PreprepareMessage
	PrepareMessages() []PrepareMessage
}

type NewViewHeader interface {
	HasMessageType
	BlockHeight() BlockHeight
	View() View
	ViewChangeConfirmations() []ViewChangeConfirmation
}

type ViewChangeConfirmation interface {
	Sender() SenderSignature
	SignedHeader() ViewChangeHeader
}

//type MessageType string
//
//const (
//	MESSAGE_TYPE_PREPREPARE  MessageType = "preprepare"
//	MESSAGE_TYPE_PREPARE     MessageType = "prepare"
//	MESSAGE_TYPE_COMMIT      MessageType = "commit"
//	MESSAGE_TYPE_VIEW_CHANGE MessageType = "view-change"
//	MESSAGE_TYPE_NEW_VIEW    MessageType = "new-view"
//)

//type Message interface {
//	SignaturePair() SignaturePair
//}
//
//type BlockRefMessage struct {
//	SignaturePair *SignaturePair
//	Content       *BlockMessageContent
//}
//
//type PrePrepareMessage struct {
//	*BlockRefMessage
//	Block *types.Block
//}
//
//type PrepareMessage struct {
//	*BlockRefMessage
//}
//
//type CommitMessage struct {
//	*BlockRefMessage
//}
//
//type ViewChangeMessage struct {
//	Content       *ViewChangeMessageContent
//	SignaturePair *SignaturePair
//	Block         *types.Block // optional
//}
//
//type NewViewMessage struct {
//	SignaturePair     *SignaturePair
//	PreprepareMessage *PrePrepareMessage
//	Content           *NewViewContent
//}
//
//type NewViewContent struct {
//	MessageType   MessageType
//	BlockHeight          types.BlockHeight
//	View          types.View
//	Confirmations []*ViewChangeConfirmation
//}
//
//type ViewChangeConfirmation struct {
//	Content       *ViewChangeMessageContent
//	SignaturePair *SignaturePair
//}
//
//type PreparedMessages struct {
//	PreprepareMessage *PrePrepareMessage
//	PrepareMessages   []*PrepareMessage
//}
//
//type SignaturePair struct {
//	SignerPublicKey  types.PublicKey
//	ContentSignature string
//}
//
//type BlockMessageContent struct {
//	MessageType MessageType
//	BlockHeight        types.BlockHeight
//	View        types.View
//	BlockHash   types.BlockHash
//}
//
//type ViewChangeMessageContent struct {
//	MessageType   MessageType
//	BlockHeight          types.BlockHeight
//	View          types.View
//	PreparedProof *PreparedProof
//}
//
//type PreparedProof struct {
//	PreprepareBlockRefMessage *PrePrepareMessage
//	PrepareBlockRefMessages   []*PrepareMessage
//}
