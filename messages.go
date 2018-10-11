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
	Serializable
	SignedHeader() BlockRef
	Sender() SenderSignature
	Block() Block
}

type PrepareMessage interface {
	Serializable
	SignedHeader() BlockRef
	Sender() SenderSignature
}

type CommitMessage interface {
	Serializable
	SignedHeader() BlockRef
	Sender() SenderSignature
}

type ViewChangeMessage interface {
	Serializable
	SignedHeader() ViewChangeHeader
	Sender() SenderSignature
	Block() Block
}

type NewViewMessage interface {
	Serializable
	SignedHeader() NewViewHeader
	PreprepareMessage() PreprepareMessage
	Sender() SenderSignature
}

// END MESSAGE TYPES //

// START MESSAGE PARTS TYPES //

type BlockRef interface {
	Serializable
	HasMessageType
	BlockHeight() BlockHeight
	View() View
	BlockHash() BlockHash
}

type ViewChangeHeader interface {
	Serializable
	HasMessageType
	BlockHeight() BlockHeight
	View() View
	PreparedProof() PreparedProof
}

type SenderSignature interface {
	Serializable
	SenderPublicKey() PublicKey
	Signature() Signature
}

type HasMessageType interface {
	MessageType() MessageType
}

type Serializable interface {
	Serialize() []byte
}

// TODO this is different from definition of LeanHelixPreparedProof in lean_helix.mb.go:448 in orbs-spec
type PreparedProof interface {
	Serializable
	PPBlockRef() BlockRef
	PBlockRef() BlockRef
	PPSender() SenderSignature
	PSenders() []SenderSignature
}

type NewViewHeader interface {
	Serializable
	HasMessageType
	BlockHeight() BlockHeight
	View() View
	ViewChangeConfirmations() []ViewChangeConfirmation
}

type ViewChangeConfirmation interface {
	Serializable
	SignedHeader() ViewChangeHeader
	Sender() SenderSignature
}

// END MESSAGE PART TYPES

type MessageTransporter interface {
	SenderSignature
	HasMessageType
}

//type MessageFactory interface {
//	CreatePreprepareMessage(blockHeight BlockHeight, view View, block Block) PreprepareMessage
//	CreatePrepareMessage(blockHeight BlockHeight, view View, blockHash BlockHash) PrepareMessage
//	CreateCommitMessage(blockHeight BlockHeight, view View, blockHash BlockHash) CommitMessage
//	CreateViewChangeMessage(blockHeight BlockHeight, view View, preparedMessages []PreprepareMessage) ViewChangeMessage
//	CreateNewViewMessage(blockHeight BlockHeight, view View, preprepareMessage PreprepareMessage, viewChangeConfirmations []ViewChangeConfirmation) NewViewMessage
//	//CreatePreparedProof(preprepare PreprepareMessage, prepares []PrepareMessage) PreparedProof
//}
//
type MessageFactory interface {
	// Message creation methods

	CreatePreprepareMessage(blockRef BlockRef, sender SenderSignature, block Block) PreprepareMessage
	CreatePrepareMessage(blockRef BlockRef, sender SenderSignature) PrepareMessage
	CreateCommitMessage(blockRef BlockRef, sender SenderSignature) CommitMessage
	CreateViewChangeMessage(vcHeader ViewChangeHeader, sender SenderSignature, block Block) ViewChangeMessage
	CreateNewViewMessage(preprepareMessage PreprepareMessage, nvHeader NewViewHeader, sender SenderSignature) NewViewMessage

	// Auxiliary methods

	CreateSenderSignature(sender []byte, signature []byte) SenderSignature
	CreateBlockRef(messageType int, blockHeight int, view int, blockHash []byte) BlockRef
	CreateNewViewHeader(messageType int, blockHeight int, view int, confirmations []ViewChangeConfirmation) NewViewHeader
	CreateViewChangeConfirmation(vcHeader ViewChangeHeader, sender SenderSignature) ViewChangeConfirmation
	CreateViewChangeHeader(blockHeight int, view int, proof PreparedProof) ViewChangeHeader
	CreatePreparedProof(ppBlockRef BlockRef, pBlockRef BlockRef, ppSender SenderSignature, pSenders []SenderSignature) PreparedProof

	// TODO Remove old methods once not needed
}

type InternalMessageFactory interface {
	// TODO USe TDD to decide on methods here
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
